// Package libp2p implements the Mir transport interface using libp2p.
package libp2p

//go:generate go run ./gen/gen.go

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	mirnet "github.com/filecoin-project/mir/pkg/net"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	"github.com/filecoin-project/mir/pkg/types"
)

type TransportMessage struct {
	Sender  string
	Payload []byte
}

var _ mirnet.Transport = &Transport{}

var ErrUnknownNode = errors.New("unknown node")
var ErrNilStream = errors.New("stream has not been opened")

type connInfo struct {
	Connecting bool
	AddrInfo   *peer.AddrInfo
	Stream     network.Stream
}

func (ci *connInfo) isConnected(h host.Host) bool {
	return ci.Stream != nil && h.Network().Connectedness(ci.AddrInfo.ID) == network.Connected
}

type newStream struct {
	NodeID types.NodeID
	Stream network.Stream
}

type Transport struct {
	params           Params
	host             host.Host
	ownID            types.NodeID
	logger           logging.Logger
	incomingMessages chan *events.EventList
	conns            map[types.NodeID]*connInfo
	connsLock        sync.RWMutex
	stopChan         chan struct{}
	streamChan       chan *newStream
	wg               sync.WaitGroup
}

func NewTransport(params Params, h host.Host, ownID types.NodeID, logger logging.Logger) (*Transport, error) {
	if logger == nil {
		logger = logging.ConsoleErrorLogger
	}

	return &Transport{
		params:           params,
		ownID:            ownID,
		host:             h,
		incomingMessages: make(chan *events.EventList),
		logger:           logger,
		conns:            make(map[types.NodeID]*connInfo),
		stopChan:         make(chan struct{}),
		streamChan:       make(chan *newStream, 1),
	}, nil
}

func (t *Transport) Start() error {
	t.logger.Log(logging.LevelDebug, "starting libp2p transport", "src", t.ownID, "listen", t.host.Addrs())
	t.runStreamUpdater()
	t.host.SetStreamHandler(t.params.ProtocolID, t.mirHandler)
	return nil
}

func (t *Transport) Stop() {
	t.logger.Log(logging.LevelDebug, "stopping libp2p transport", "src", t.ownID)
	defer t.logger.Log(logging.LevelDebug, "libp2p transport stopped", "src", t.ownID)

	close(t.stopChan)

	t.host.RemoveStreamHandler(t.params.ProtocolID)

	t.connsLock.Lock()
	for nodeID, c := range t.conns {
		if c.Stream != nil {
			if err := c.Stream.Close(); err != nil {
				t.logger.Log(logging.LevelError, "could not close stream", "src", t.ownID, "dst", nodeID, "err", err)
			}
		}
		delete(t.conns, nodeID)
	}
	t.connsLock.Unlock()

	for _, c := range t.host.Network().Conns() {
		for _, s := range c.GetStreams() {
			if s.Protocol() == t.params.ProtocolID {
				err := c.Close()
				if err != nil {
					t.logger.Log(logging.LevelError, "could not close connection", "src", t.ownID, "dst", c.RemotePeer().String(), "err", err)
				}
			}
		}
	}

	t.wg.Wait()
}

func (t *Transport) CloseOldConnections(newNodes map[types.NodeID]types.NodeAddress) {
	t.connsLock.Lock()
	defer t.connsLock.Unlock()

	for nodeID, c := range t.conns {
		// Close the old stream to a node.
		if _, foundInNewNodes := newNodes[nodeID]; !foundInNewNodes {
			t.logger.Log(logging.LevelDebug, "closing old connection", "src", t.ownID, "dst", nodeID)

			if c.Stream != nil {
				if err := c.Stream.Close(); err != nil {
					t.logger.Log(logging.LevelError, "could not close old stream to node", "src", t.ownID, "dst", nodeID, "err", err)
				}
			}

			delete(t.conns, nodeID)
		}
	}
}

func (t *Transport) Connect(nodes map[types.NodeID]types.NodeAddress) {
	if len(nodes) == 0 {
		t.logger.Log(logging.LevelWarn, "no nodes to connect to")
		return
	}

	parsedAddrInfo := make(map[types.NodeID]*peer.AddrInfo)
	for nodeID, addr := range nodes {
		info, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			t.logger.Log(logging.LevelWarn, "connect: failed to parse addr", "src", t.ownID, "dest", nodeID, "addr", addr, "err", err)
			continue
		}
		parsedAddrInfo[nodeID] = info
	}

	t.connsLock.Lock()
	toConnect := make([]types.NodeID, 0, len(parsedAddrInfo))
	for nodeID := range parsedAddrInfo {
		if nodeID == t.ownID {
			continue
		}
		conn, found := t.conns[nodeID]
		if !found || conn.Stream == nil {
			t.conns[nodeID] = &connInfo{AddrInfo: parsedAddrInfo[nodeID], Stream: nil}
			toConnect = append(toConnect, nodeID)
		}

	}
	t.connsLock.Unlock()

	if len(toConnect) > 0 {
		t.logger.Log(logging.LevelDebug, "connecting to new nodes", "src", t.ownID, "to", toConnect)
		t.connect(toConnect)
	}
}

// WaitFor polls the current connection state and returns when at least n connections have been established.
func (t *Transport) WaitFor(n int) {
	for {
		t.connsLock.RLock()
		numConnections := 0
		for _, ci := range t.conns {
			if ci.isConnected(t.host) {
				numConnections++
			}
		}
		t.connsLock.RUnlock()

		// We subtract one, as we always assume to be connected to ourselves
		// and the connection to self does not appear in t.conns.
		if numConnections >= n-1 {
			return
		}
		time.Sleep(t.params.ConnWaitPollInterval)
	}
}

// ConnectSync is a convenience method that triggers the connection process
// and waits for n connections to be established before returning.
// Equivalent to calling Connect and WaitFor.
func (t *Transport) ConnectSync(nodes map[types.NodeID]types.NodeAddress, n int) {
	t.Connect(nodes)
	t.WaitFor(n)
}

func (t *Transport) Send(dest types.NodeID, msg *messagepb.Message) error {
	select {
	case <-t.stopChan:
		return fmt.Errorf("transport stopped")
	default:
	}

	out, err := t.encode(msg)
	if err != nil {
		return err
	}

	err = t.sendPayload(dest, out)
	if err != nil {
		t.connect([]types.NodeID{dest})
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
					if err := t.Send(types.NodeID(destID), e.SendMessage.Msg); err != nil {
						t.logger.Log(logging.LevelWarn, "failed to send a message", "err", err)
					}
				}
			}
		default:
			return fmt.Errorf("unexpected event: %T", event.Type)
		}
	}

	return nil
}

func (t *Transport) runStreamUpdater() {
	t.wg.Add(1)
	go func() {
		defer t.wg.Done()

		t.logger.Log(logging.LevelDebug, "stream updater started", "src", t.ownID)
		defer t.logger.Log(logging.LevelDebug, "stream updater stopped", "src", t.ownID)

		for {
			select {
			case <-t.stopChan:
				t.logger.Log(logging.LevelError, "stream updater received stop signal", "src", t.ownID)
				return
			case one := <-t.streamChan:
				nodeID := one.NodeID
				t.connsLock.Lock()
				conn, found := t.conns[nodeID]
				if found {
					if conn.Stream != nil {
						if err := conn.Stream.Close(); err != nil {
							t.logger.Log(logging.LevelError, "could not close stream to node", "src", t.ownID, "dst", nodeID, "err", err)
						}
					}
					conn.Stream = one.Stream
					conn.Connecting = false
				}
				t.connsLock.Unlock()
				t.logger.Log(logging.LevelDebug, "updated stream", "src", t.ownID, "nodeID", nodeID)
			}
		}
	}()
}

func (t *Transport) connect(nodesID []types.NodeID) {
	for i := range nodesID {
		// Do not establish a real connection with own node.
		if nodesID[i] == t.ownID {
			continue
		}

		t.wg.Add(1)
		go t.connectToNode(nodesID[i])
	}
}

func (t *Transport) connectToNode(nodeID types.NodeID) {
	defer t.wg.Done()

	t.connsLock.Lock()
	conn, found := t.conns[nodeID]
	if !found {
		t.connsLock.Unlock()
		t.logger.Log(logging.LevelError, "failed to get node address", "src", t.ownID, "dst", nodeID)
		return
	}

	if conn.Connecting {
		t.connsLock.Unlock()
		t.logger.Log(logging.LevelWarn, "already connecting", "src", t.ownID, "dst", nodeID)
		return
	}

	info := conn.AddrInfo
	t.host.Peerstore().AddAddrs(info.ID, info.Addrs, t.params.PermanentAddrTTL)
	conn.Connecting = true
	t.connsLock.Unlock()

	t.logger.Log(logging.LevelDebug, fmt.Sprintf("node %v is connecting to node %v", t.ownID, nodeID))
	s, err := t.openStream(nodeID, info.ID)
	if err != nil {
		t.logger.Log(logging.LevelError, fmt.Sprintf("node %v failed to connect to node %v: %s", t.ownID, nodeID, err))
		t.connsLock.Lock()
		conn.Connecting = false
		t.connsLock.Unlock()
		return
	}

	t.streamChan <- &newStream{nodeID, s}
	t.logger.Log(logging.LevelDebug, fmt.Sprintf("node %s has connected to node %s", t.ownID.Pb(), nodeID.Pb()))
}

func (t *Transport) openStream(dest types.NodeID, p peer.ID) (network.Stream, error) {
	// We need the simplest retry mechanism due to the fact that the underlying libp2p's NewStream function dials once:
	// https://github.com/libp2p/go-libp2p/blob/7828f3e0797e0a7b7033fa5e8be9b94f57a4c173/p2p/net/swarm/swarm.go#L358
	t.logger.Log(logging.LevelDebug, "start opening stream to peer", "src", t.ownID, "dst", dest)
	defer t.logger.Log(logging.LevelDebug, "stop opening stream to peer", "src", t.ownID, "dst", dest)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		select {
		case <-ctx.Done():
			return
		case <-t.stopChan:
			cancel()
		}
	}()

	// The implementation is based on the openStream() function from the RemoteTracer:
	// https://github.com/libp2p/go-libp2p-pubsub/blob/cbb7bfc1f182e0b765d2856f6a0ea73e34d93602/tracer.go#L280
	var s network.Stream
	var err error
	for i := 0; i < t.params.MaxRetries; i++ {
		select {
		case <-t.stopChan:
			return nil, fmt.Errorf("%s opening stream to %s: stop chan closed", t.ownID, dest)
		default:
		}
		sctx, scancel := context.WithTimeout(ctx, t.params.MaxConnectingTimeout)

		s, err = t.host.NewStream(sctx, p, t.params.ProtocolID)
		scancel()
		if err == nil {
			return s, nil
		}

		if i >= t.params.NoLoggingErrorAttempts {
			t.logger.Log(
				logging.LevelError, fmt.Sprintf("%s failed to open stream to %s: %v", t.ownID, dest, err))
		} else {
			t.logger.Log(
				logging.LevelInfo, fmt.Sprintf("%s failed to open stream to %s: %v", t.ownID, dest, err))
		}

		delay := time.NewTimer(t.params.MaxRetryTimeout)
		select {
		case <-delay.C:
			continue
		case <-t.stopChan:
			if !delay.Stop() {
				<-delay.C
			}
			return nil, fmt.Errorf("%s opening stream to %s: stop chan closed", t.ownID, dest)
		}
	}
	return nil, fmt.Errorf("%s failed to open stream to %s: %w", t.ownID, dest, err)
}

func (t *Transport) sendPayload(dest types.NodeID, payload []byte) error {
	// There are two cases when we get an error:
	// 1. We don't have the node ID in the nodes table. E.g. we didn't call Connect().
	// 2. We created an entry for the node but have not opened a connection.
	// But if we added the node via Connect, then it should open the connection.
	// It doesn't make sense to open a new one here.
	s, err := t.getNodeStream(dest)
	if err != nil {
		return err
	}
	if _, err := s.Write(payload); err != nil {
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

func (t *Transport) readAndDecode(s network.Stream) (*messagepb.Message, types.NodeID, error) {
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
	peerID := s.Conn().RemotePeer().String()

	t.logger.Log(logging.LevelDebug, "mir handler started", "src", t.ownID, "dst", peerID)
	defer t.logger.Log(logging.LevelDebug, "mir handler stopped", "src", t.ownID, "dst", peerID)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		select {
		case <-ctx.Done():
			return
		case <-t.stopChan:
			t.logger.Log(logging.LevelDebug, "mir handler stop signal received", "src", t.ownID)
			if err := s.Close(); err != nil {
				t.logger.Log(logging.LevelError, "stream reset", "src", t.ownID)
			}
		}
	}()

	for {
		msg, sender, err := t.readAndDecode(s)
		if err != nil {
			if errors.Is(err, io.EOF) {
				t.logger.Log(logging.LevelDebug, "connection handler received EOF", "src", t.ownID, "dst", peerID)
				return
			}
			t.logger.Log(logging.LevelError, "failed to read mir message", "src", t.ownID, "err", err)
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

		select {
		case t.incomingMessages <- events.ListOf(
			events.MessageReceived(types.ModuleID(msg.DestModule), sender, msg),
		):
		case <-t.stopChan:
			t.logger.Log(logging.LevelError, "mir handler received stop signal", "src", t.ownID)
			return
		}
	}
}

func (t *Transport) getNodeStream(nodeID types.NodeID) (network.Stream, error) {
	t.connsLock.RLock()
	defer t.connsLock.RUnlock()

	node, found := t.conns[nodeID]
	if !found {
		return nil, ErrUnknownNode
	}

	if node.Stream == nil {
		return nil, ErrNilStream
	}
	return node.Stream, nil
}
