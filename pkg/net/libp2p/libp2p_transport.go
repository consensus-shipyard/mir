// Package libp2p implements the Mir transport interface using libp2p.
package libp2p

//go:generate go run ./gen/gen.go

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
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
var ErrNilStream = errors.New("Stream has not been opened")

type nodeInfo struct {
	ID        types.NodeID
	Addr      types.NodeAddress
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
	nodes            map[types.NodeID]*nodeInfo
	nodesLock        sync.RWMutex
}

func NewTransport(h host.Host, ownID types.NodeID, logger logging.Logger) (*Transport, error) {
	if logger == nil {
		logger = logging.ConsoleErrorLogger
	}

	return &Transport{
		connWg:           &sync.WaitGroup{},
		incomingMessages: make(chan *events.EventList),
		nodes:            make(map[types.NodeID]*nodeInfo),
		logger:           logger,
		ownID:            ownID,
		host:             h,
	}, nil
}

func (t *Transport) Start() error {
	t.logger.Log(logging.LevelDebug, "mir handler is starting", "ownID", t.ownID, "listen", t.host.Addrs())
	t.host.SetStreamHandler(ProtocolID, t.mirHandler)
	return nil
}

func (t *Transport) Stop() {
	t.logger.Log(logging.LevelDebug, "Stopping libp2p transport.", "ownID", t.ownID)
	defer t.logger.Log(logging.LevelDebug, "Stopping libp2p transport finished.", "ownID", t.ownID)

	t.host.RemoveStreamHandler(ProtocolID)

	if err := t.host.Close(); err != nil {
		t.logger.Log(logging.LevelError, "Could not close libp2p", "ownID", t.ownID, "err", err)
	} else {
		t.logger.Log(logging.LevelDebug, "libp2p host closed", "ownID", t.ownID)
	}

	t.connWg.Wait()
}

func (t *Transport) CloseOldConnections(newNodes map[types.NodeID]types.NodeAddress) {
	t.connWg.Add(1)

	go func() {
		defer t.connWg.Done()

		t.nodesLock.Lock()
		defer t.nodesLock.Unlock()

		for nodeID, dest := range t.nodes {
			// Close an old connection to a node if we don't need to connect to it further.
			if _, foundInNewNodes := newNodes[nodeID]; !foundInNewNodes {
				t.logger.Log(logging.LevelDebug, "Closing old connection", "src", t.ownID, "dst", nodeID)

				if err := t.host.Network().ClosePeer(dest.AddrInfo.ID); err != nil {
					t.logger.Log(logging.LevelError, "Could not close old connection to node", "src", t.ownID, "dst", nodeID, "err", err)
					continue
				}

				delete(t.nodes, nodeID)
			}
		}
	}()
}

func (t *Transport) Connect(ctx context.Context, nodes map[types.NodeID]types.NodeAddress) {
	if len(nodes) == 0 {
		t.logger.Log(logging.LevelWarn, "no nodes to connect to")
		return
	}

	t.nodesLock.Lock()
	for nodeID, addr := range nodes {
		if nodeID == t.ownID {
			continue
		}
		_, found := t.nodes[nodeID]
		if !found {
			info, err := peer.AddrInfoFromP2pAddr(addr)
			if err != nil {
				t.logger.Log(logging.LevelWarn, "connect: failed to parse addr", "src", t.ownID, "dest", nodeID, "addr", addr, "err", err)
				continue
			}
			t.nodes[nodeID] = &nodeInfo{
				ID:        nodeID,
				Addr:      addr,
				AddrInfo:  info,
				IsOpening: false,
				Stream:    nil,
			}

		}
	}
	t.nodesLock.Unlock()

	t.connect(ctx, nodes)
}

func (t *Transport) connect(ctx context.Context, nodes map[types.NodeID]types.NodeAddress) {
	t.connWg.Add(1)
	defer t.connWg.Done()

	go func() {
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

func (t *Transport) connectToNode(ctx context.Context, node types.NodeID, wg *sync.WaitGroup) {
	defer wg.Done()

	if t.isConnectionToNodeInProgress(node) {
		t.logger.Log(logging.LevelDebug, "connecting to node is in progress", "src", t.ownID, "dst", node)
		return
	}

	t.setConnectionInProgress(node)
	defer t.clearConnectionInProgress(node)

	t.logger.Log(logging.LevelDebug, fmt.Sprintf("node %v is connecting to node %v", t.ownID, node))

	info, found := t.getNodeAddrInfo(node)
	if !found {
		t.logger.Log(logging.LevelError, "failed to get node address", "src", t.ownID, "dst", node)
		return
	}

	t.host.Peerstore().AddAddrs(info.ID, info.Addrs, PermanentAddrTTL)

	s, err := t.openStream(ctx, info.ID)
	if err != nil {
		t.logger.Log(logging.LevelError, "failed to open stream to node", "addr", info, "node", node, "err", err)
		return
	}

	t.addOutboundStream(node, s)
	t.logger.Log(logging.LevelDebug, fmt.Sprintf("node %s has connected to node %s", t.ownID.Pb(), node.Pb()))
}

func (t *Transport) openStream(ctx context.Context, dest peer.ID) (network.Stream, error) {
	//  We need the simplest retry mechanism due to the fact that the underlying libp2p's NewStream function dials once:
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

func (t *Transport) Send(ctx context.Context, dest types.NodeID, payload *messagepb.Message) error {
	var err error

	// There are two cases when we get an error:
	// 1. We don't have the node ID in the nodes table. E.g. we didn't call Connect().
	// 2. We created an entry for the node but have not opened a connection.
	// But if we added the node via Connect, then it should open the connection.
	// It doesn't make sense to open a new one here.
	s, info, err := t.getNodeStreamAndInfo(dest)
	if err != nil {
		return err
	}

	outBytes, err := proto.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	msg := TransportMessage{t.ownID.Pb(), outBytes}
	buf := new(bytes.Buffer)
	err = msg.MarshalCBOR(buf)
	if err != nil {
		return fmt.Errorf("failed to CBOR marshal message: %w", err)
	}

	if ctx.Err() != nil {
		return err
	}

	w := bufio.NewWriter(s)
	_, err = w.Write(buf.Bytes())
	if err == nil {
		err = w.Flush()
	}
	// If we cannot write to the stream then close the connection and reconnect.
	if err != nil {
		if err := t.host.Network().ClosePeer(info.ID); err != nil {
			t.logger.Log(logging.LevelError, "could not close the connection to node", "src", t.ownID, "dst", dest, "err", err)
		}
		t.connWg.Add(1)
		go t.connectToNode(ctx, dest, t.connWg)
		return errors.Wrapf(err, "%s failed to send data to %s", t.ownID, dest)
	}

	return nil
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
		var msg TransportMessage
		err := msg.UnmarshalCBOR(s)
		if err != nil {
			t.logger.Log(logging.LevelError, "failed to read mir request", "src", t.ownID, "dst", s.ID(), "err", err)
			return
		}

		var payload messagepb.Message

		if err := proto.Unmarshal(msg.Payload, &payload); err != nil {
			t.logger.Log(logging.LevelError, "failed to unmarshall mir request", "src", t.ownID, "dst", s.ID(), "err", err)
			return
		}

		t.incomingMessages <- events.ListOf(
			events.MessageReceived(types.ModuleID(payload.DestModule), types.NodeID(msg.Sender), &payload),
		)
	}
}

func (t *Transport) addOutboundStream(nodeID types.NodeID, s network.Stream) {
	t.nodesLock.Lock()
	defer t.nodesLock.Unlock()

	node, found := t.nodes[nodeID]
	if !found {
		t.logger.Log(logging.LevelWarn, "addOutboundStream: failed to find the node", "src", t.ownID, "node", nodeID)
		return
	}
	node.Stream = s
}

func (t *Transport) streamExists(nodeID types.NodeID) bool {
	t.nodesLock.RLock()
	defer t.nodesLock.RUnlock()

	v, found := t.nodes[nodeID]
	return found && v.Stream != nil
}

func (t *Transport) getNodeStreamAndInfo(nodeID types.NodeID) (network.Stream, *peer.AddrInfo, error) {
	t.nodesLock.Lock()
	defer t.nodesLock.Unlock()

	_, found := t.nodes[nodeID]
	if !found {
		return nil, nil, ErrUnknownNode
	}

	node, found := t.nodes[nodeID]
	if !found ||
		node.Stream == nil {
		return nil, nil, ErrNilStream
	}
	return node.Stream, node.AddrInfo, nil
}

func (t *Transport) getNodeAddrInfo(nodeID types.NodeID) (*peer.AddrInfo, bool) {
	t.nodesLock.RLock()
	defer t.nodesLock.RUnlock()

	node, found := t.nodes[nodeID]
	if !found {
		return nil, false
	}
	return node.AddrInfo, true
}

func (t *Transport) isConnectionToNodeInProgress(nodeID types.NodeID) bool {
	t.nodesLock.RLock()
	defer t.nodesLock.RUnlock()

	info, found := t.nodes[nodeID]
	return found && info.IsOpening
}

func (t *Transport) setConnectionInProgress(nodeID types.NodeID) {
	t.nodesLock.Lock()
	defer t.nodesLock.Unlock()

	info, found := t.nodes[nodeID]
	if !found {
		return
	}
	info.IsOpening = true
}

func (t *Transport) clearConnectionInProgress(nodeID types.NodeID) {
	t.nodesLock.Lock()
	defer t.nodesLock.Unlock()

	info, found := t.nodes[nodeID]
	if !found {
		return
	}
	info.IsOpening = false
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
