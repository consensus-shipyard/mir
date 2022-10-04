// Package libp2p implements the Mir transport interface using libp2p.
package libp2p

//go:generate go run ./gen/gen.go

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	mirnet "github.com/filecoin-project/mir/pkg/net"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	"github.com/filecoin-project/mir/pkg/types"
)

const (
	ProtocolID             = "/mir/0.0.1"
	maxConnectingTimeout   = 1000 * time.Millisecond
	maxRetryTimeout        = 2 * time.Second
	maxRetries             = 20
	noLoggingErrorAttempts = 2
	PermanentAddrTTL       = math.MaxInt64 - iota
)

type TransportMessage struct {
	Sender  string
	Payload []byte
}

var _ mirnet.Transport = &Transport{}

var ErrUnknownNode = errors.New("unknown node")
var ErrNilStream = errors.New("stream has not been opened")

type connInfo struct {
	AddrInfo  *peer.AddrInfo
	Stream    network.Stream
	IsOpening bool
}

type Transport struct {
	host             host.Host
	ownID            types.NodeID
	connWg           *sync.WaitGroup
	incomingMessages chan *events.EventList
	logger           logging.Logger
	conns            map[types.NodeID]*connInfo
	connsLock        sync.RWMutex
}

func NewTransport(h host.Host, ownID types.NodeID, logger logging.Logger) (*Transport, error) {
	if logger == nil {
		logger = logging.ConsoleErrorLogger
	}

	return &Transport{
		connWg:           &sync.WaitGroup{},
		incomingMessages: make(chan *events.EventList),
		conns:            make(map[types.NodeID]*connInfo),
		logger:           logger,
		ownID:            ownID,
		host:             h,
	}, nil
}

func (t *Transport) Start() error {
	t.logger.Log(logging.LevelDebug, "starting libp2p transport", "ownID", t.ownID, "listen", t.host.Addrs())
	t.host.SetStreamHandler(ProtocolID, t.mirHandler)
	return nil
}

func (t *Transport) Stop() {
	t.logger.Log(logging.LevelDebug, "stopping libp2p transport", "ownID", t.ownID)
	defer t.logger.Log(logging.LevelDebug, "stopping libp2p transport finished.", "ownID", t.ownID)

	t.host.RemoveStreamHandler(ProtocolID)

	t.connsLock.Lock()
	for nodeID, addr := range t.conns {
		if err := t.host.Network().ClosePeer(addr.AddrInfo.ID); err != nil {
			t.logger.Log(logging.LevelError, "could not close old connection to node", "src", t.ownID, "dst", nodeID, "err", err)
			continue
		}
	}
	t.connsLock.Unlock()

	t.connWg.Wait()
}

func (t *Transport) CloseOldConnections(newNodes map[types.NodeID]types.NodeAddress) {
	t.connWg.Add(1)

	go func() {
		defer t.connWg.Done()

		t.connsLock.Lock()
		defer t.connsLock.Unlock()

		for nodeID, dest := range t.conns {
			// Close an old connection to a node if we don't need to connect to it further.
			if _, foundInNewNodes := newNodes[nodeID]; !foundInNewNodes {
				t.logger.Log(logging.LevelDebug, "closing old connection", "src", t.ownID, "dst", nodeID)

				if err := t.host.Network().ClosePeer(dest.AddrInfo.ID); err != nil {
					t.logger.Log(logging.LevelError, "could not close old connection to node", "src", t.ownID, "dst", nodeID, "err", err)
					continue
				}

				delete(t.conns, nodeID)
			}
		}
	}()
}

func (t *Transport) Connect(ctx context.Context, nodes map[types.NodeID]types.NodeAddress) {
	if len(nodes) == 0 {
		t.logger.Log(logging.LevelWarn, "no nodes to connect to")
		return
	}

	t.connsLock.Lock()
	for nodeID, addr := range nodes {
		if nodeID == t.ownID {
			continue
		}
		_, found := t.conns[nodeID]
		if !found {
			info, err := peer.AddrInfoFromP2pAddr(addr)
			if err != nil {
				t.logger.Log(logging.LevelWarn, "connect: failed to parse addr", "src", t.ownID, "dest", nodeID, "addr", addr, "err", err)
				continue
			}
			t.conns[nodeID] = &connInfo{
				AddrInfo:  info,
				IsOpening: false,
				Stream:    nil,
			}

		}
	}
	t.connsLock.Unlock()

	t.connect(ctx, nodes)
}

func (t *Transport) Send(ctx context.Context, dest types.NodeID, msg *messagepb.Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	out, err := t.encode(msg)
	if err != nil {
		return err
	}

	err = t.sendPayload(dest, out)
	if err != nil {
		t.reconnect(ctx, dest)
		return errors.Wrapf(err, "%s failed to send data to %s", t.ownID, dest)
	}

	return nil
}

func (t *Transport) ImplementsModule() {}

func (t *Transport) EventsOut() <-chan *events.EventList {
	return t.incomingMessages
}

func (t *Transport) ApplyEvents(ctx context.Context, eventList *events.EventList) error {
	iter := eventList.Iterator()
	for event := iter.Next(); event != nil; event = iter.Next() {

		switch e := event.Type.(type) {
		case *eventpb.Event_Init:
			// no actions on init
		case *eventpb.Event_SendMessage:
			for _, destID := range e.SendMessage.Destinations {
				if types.NodeID(destID) == t.ownID {
					// Send message to myself bypassing the network.
					// The sending must be done in its own goroutine in case writing to gt.incomingMessages blocks.
					// (Processing of input events must be non-blocking.)
					receivedEvent := events.MessageReceived(
						types.ModuleID(e.SendMessage.Msg.DestModule),
						t.ownID,
						e.SendMessage.Msg,
					)
					go func() {
						select {
						case t.incomingMessages <- events.ListOf(receivedEvent):
						case <-ctx.Done():
						}
					}()
				} else {
					// Send message to another node.
					if err := t.Send(ctx, types.NodeID(destID), e.SendMessage.Msg); err != nil {
						t.logger.Log(logging.LevelError, "failed to send a message", "err", err)
					}
				}
			}
		default:
			return fmt.Errorf("unexpected event: %T", event.Type)
		}
	}

	return nil
}

func (t *Transport) connect(ctx context.Context, nodes map[types.NodeID]types.NodeAddress) {
	t.connWg.Add(1)

	go func() {
		defer t.connWg.Done()

		t.logger.Log(logging.LevelDebug, "started connecting nodes", "src", t.ownID)
		defer t.logger.Log(logging.LevelDebug, "finished connecting nodes", "src", t.ownID)

		wg := &sync.WaitGroup{}
		wg.Add(len(nodes))

		for nodeID := range nodes {
			if nodeID == t.ownID {
				// Do not establish a real connection with own node.
				wg.Done()
				continue
			}

			if t.streamExists(nodeID) {
				t.logger.Log(logging.LevelInfo, "stream to node already exists", "src", t.ownID, "dst", nodeID)
				wg.Done()
				continue
			}

			go t.connectToNode(ctx, nodeID, wg)
		}
		wg.Wait()
	}()
}

func (t *Transport) reconnect(ctx context.Context, node types.NodeID) {
	info, found := t.getNodeAddrInfo(node)
	if !found {
		t.logger.Log(logging.LevelError, "failed to get node address", "src", t.ownID, "dst", node)
	} else {
		if err := t.host.Network().ClosePeer(info.ID); err != nil {
			t.logger.Log(logging.LevelError, "could not close the connection to node", "src", t.ownID, "dst", node, "err", err)
		}
		t.connWg.Add(1)
		go t.connectToNode(ctx, node, t.connWg)
	}
}

func (t *Transport) connectToNode(ctx context.Context, nodeID types.NodeID, wg *sync.WaitGroup) {
	defer wg.Done()

	if t.isConnectionInProgress(nodeID) {
		t.logger.Log(logging.LevelDebug, "connection is in progress or ready", "src", t.ownID, "dst", nodeID)
		return
	}

	t.setConnectionInProgress(nodeID)
	defer t.clearConnectionInProgress(nodeID)

	t.logger.Log(logging.LevelDebug, fmt.Sprintf("node %v is connecting to node %v", t.ownID, nodeID))

	info, found := t.getNodeAddrInfo(nodeID)
	if !found {
		t.logger.Log(logging.LevelError, "failed to get node address", "src", t.ownID, "dst", nodeID)
		return
	}

	t.host.Peerstore().AddAddrs(info.ID, info.Addrs, PermanentAddrTTL)

	s, err := t.openStream(ctx, info.ID)
	if err != nil {
		t.logger.Log(logging.LevelError, "failed to open stream to node", "addr", info, "node", nodeID, "err", err)
		return
	}

	t.addOutboundStream(nodeID, s)
	t.logger.Log(logging.LevelDebug, fmt.Sprintf("node %s has connected to node %s", t.ownID.Pb(), nodeID.Pb()))
}

func (t *Transport) openStream(ctx context.Context, dest peer.ID) (network.Stream, error) {
	// We need the simplest retry mechanism due to the fact that the underlying libp2p's NewStream function dials once:
	// https://github.com/libp2p/go-libp2p/blob/7828f3e0797e0a7b7033fa5e8be9b94f57a4c173/p2p/net/swarm/swarm.go#L358
	t.logger.Log(logging.LevelDebug, "opening stream to peer", "src", t.ownID, "dst", dest)

	// The implementation is based on the openStream() function from the RemoteTracer:
	// https://github.com/libp2p/go-libp2p-pubsub/blob/cbb7bfc1f182e0b765d2856f6a0ea73e34d93602/tracer.go#L280
	var s network.Stream
	var err error
	for i := 0; i < maxRetries; i++ {
		sctx, cancel := context.WithTimeout(ctx, maxConnectingTimeout)

		s, err = t.host.NewStream(sctx, dest, ProtocolID)
		cancel()

		if err == nil {
			return s, nil
		}

		if i >= noLoggingErrorAttempts {
			t.logger.Log(
				logging.LevelError, fmt.Sprintf("%s failed to open stream to %s: %v", t.ownID, dest, err))
		} else {
			t.logger.Log(
				logging.LevelInfo, fmt.Sprintf("%s failed to open stream to %s: %v", t.ownID, dest, err))
		}

		delay := time.NewTimer(maxRetryTimeout)
		select {
		case <-delay.C:
			continue
		case <-ctx.Done():
			if !delay.Stop() {
				<-delay.C
			}
			return nil, fmt.Errorf("%s opening stream to %s: context closed", t.ownID, dest)
		}
	}
	return nil, fmt.Errorf("%s failed to open stream to %s: %w", t.ownID, dest, err)
}

func (t *Transport) sendPayload(dest types.NodeID, payload []byte) error {
	t.connsLock.RLock()
	defer t.connsLock.RUnlock()

	// There are two cases when we get an error:
	// 1. We don't have the node ID in the nodes table. E.g. we didn't call Connect().
	// 2. We created an entry for the node but have not opened a connection.
	// But if we added the node via Connect, then it should open the connection.
	// It doesn't make sense to open a new one here.
	stream, err := t.getNodeStreamAndInfoWithoutLock(dest)
	if err != nil {
		return err
	}

	if _, err := stream.Write(payload); err != nil {
		return err
	}

	return nil
}

func (t *Transport) encode(msg *messagepb.Message) ([]byte, error) {
	p, err := proto.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	tm := TransportMessage{t.ownID.Pb(), p}
	buf := new(bytes.Buffer)
	if err = tm.MarshalCBOR(buf); err != nil {
		return nil, fmt.Errorf("failed to CBOR marshal message: %w", err)
	}
	return buf.Bytes(), nil
}

func (t *Transport) decode(s io.Reader) (*messagepb.Message, types.NodeID, error) {
	var tm TransportMessage
	err := tm.UnmarshalCBOR(s)
	if err != nil {
		return nil, "", err
	}

	var msg messagepb.Message
	if err := proto.Unmarshal(tm.Payload, &msg); err != nil {
		return nil, "", err
	}
	return &msg, types.NodeID(tm.Sender), nil
}

func (t *Transport) mirHandler(s network.Stream) {
	t.logger.Log(logging.LevelDebug, "mir handler started", "src", t.ownID, "dst", s.ID())
	defer t.logger.Log(logging.LevelDebug, "mir handler stopped", "src", t.ownID, "dst", s.ID())

	defer func() {
		t.logger.Log(logging.LevelDebug, "mir handler is closing stream", "src", t.ownID, "dst", s.ID())
		err := s.Close()
		if err != nil {
			t.logger.Log(logging.LevelError, "closing stream", "src", t.ownID, "dst", s.ID(), "err", err)
		}
	}() // nolint

	for {
		msg, sender, err := t.decode(s)
		if err != nil {
			t.logger.Log(logging.LevelError, "failed to read mir message", "src", t.ownID, "sender", sender, "err", err)
			return
		}
		t.connsLock.RLock()
		senderNode, ok := t.conns[sender]
		t.connsLock.RUnlock()
		if !ok {
			return
		}
		senderNodeID := senderNode.AddrInfo.ID.String()
		senderPeerID := s.Conn().RemotePeer().String()
		if senderNodeID != senderPeerID {
			t.logger.Log(logging.LevelWarn,
				"failed to validate sender",
				"src", t.ownID, "sender", sender, "senderNodeID", senderNodeID, "senderPeerID", senderPeerID)
			return
		}
		t.incomingMessages <- events.ListOf(
			events.MessageReceived(types.ModuleID(msg.DestModule), sender, msg),
		)
	}
}

func (t *Transport) addOutboundStream(nodeID types.NodeID, s network.Stream) {
	t.connsLock.Lock()
	defer t.connsLock.Unlock()

	node, found := t.conns[nodeID]
	if !found {
		t.logger.Log(logging.LevelWarn, "addOutboundStream: failed to find the node", "src", t.ownID, "node", nodeID)
		return
	}
	node.Stream = s
}

func (t *Transport) streamExists(nodeID types.NodeID) bool {
	t.connsLock.RLock()
	defer t.connsLock.RUnlock()

	conn, found := t.conns[nodeID]
	return found && conn.Stream != nil
}

func (t *Transport) getNodeStreamAndInfoWithoutLock(nodeID types.NodeID) (network.Stream, error) {
	node, found := t.conns[nodeID]
	if !found {
		return nil, ErrUnknownNode
	}

	if node.Stream == nil {
		return nil, ErrNilStream
	}
	return node.Stream, nil
}

func (t *Transport) getNodeAddrInfo(nodeID types.NodeID) (*peer.AddrInfo, bool) {
	t.connsLock.RLock()
	defer t.connsLock.RUnlock()

	node, found := t.conns[nodeID]
	if !found {
		return nil, false
	}
	return node.AddrInfo, true
}

func (t *Transport) isConnectionInProgress(nodeID types.NodeID) bool {
	t.connsLock.RLock()
	defer t.connsLock.RUnlock()

	conn, found := t.conns[nodeID]
	return found && conn.IsOpening
}

func (t *Transport) setConnectionInProgress(nodeID types.NodeID) {
	t.connsLock.Lock()
	defer t.connsLock.Unlock()

	info, found := t.conns[nodeID]
	if !found {
		return
	}
	info.IsOpening = true
}

func (t *Transport) clearConnectionInProgress(nodeID types.NodeID) {
	t.connsLock.Lock()
	defer t.connsLock.Unlock()

	info, found := t.conns[nodeID]
	if !found {
		return
	}
	info.IsOpening = false
}
