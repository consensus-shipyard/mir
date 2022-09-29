// Package libp2p implements libp2p-based transport for Mir protocol framework.
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
	maxConnectingTimeout   = 200 * time.Millisecond
	retryTimeout           = 2 * time.Second
	retryAttempts          = 20
	noLoggingErrorAttempts = 2
	PermanentAddrTTL       = math.MaxInt64 - iota
)

type TransportMessage struct {
	Sender  string
	Payload []byte
}

var _ mirnet.Transport = &Transport{}

type nodeInfo struct {
	ID        types.NodeID
	Addr      types.NodeAddress
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

type nodeInfoError struct {
	DestNode types.NodeID
}

func (e *nodeInfoError) Error() string {
	return fmt.Sprintf("failed to get info for node %s on sending", e.DestNode)
}

type streamInfoError struct {
	DestNode types.NodeID
}

func (e *streamInfoError) Error() string {
	return fmt.Sprintf("no open stream for node %s", e.DestNode)
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

func (t *Transport) ImplementsModule() {}

func (t *Transport) EventsOut() <-chan *events.EventList {
	return t.incomingMessages
}

func (t *Transport) Start() error {
	t.logger.Log(logging.LevelDebug, "mir handler is starting", "ownID", t.ownID, "addrs", t.host.Addrs())
	t.host.SetStreamHandler(ProtocolID, t.mirHandler)
	return nil
}

func (t *Transport) Stop() {
	t.logger.Log(logging.LevelDebug, "Stopping libp2p transport.", "ownID", t.ownID)
	defer t.logger.Log(logging.LevelDebug, "Stopping libp2p transport finished.", "ownID", t.ownID)

	t.nodesLock.Lock()
	defer t.nodesLock.Unlock()

	for id, info := range t.nodes {
		if info.Stream == nil {
			continue
		}
		t.logger.Log(logging.LevelDebug, "Closing connection", "to", id)

		if err := info.Stream.Close(); err != nil {
			t.logger.Log(logging.LevelError, "Could not close connection.", "nodeID", id, "err", err)
			continue
		}

		t.logger.Log(logging.LevelDebug, "Closed connection", "to", id)
	}

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

		for nodeID, nodeInfo := range t.nodes {
			// Close an old connection to a node if we don't need to connect to it further.
			if _, foundInNewNodes := newNodes[nodeID]; !foundInNewNodes {
				t.logger.Log(logging.LevelDebug, "Closing old connection", "src", t.ownID, "dst", nodeID)

				if nodeInfo.Stream != nil {
					info, err := peer.AddrInfoFromP2pAddr(nodeInfo.Addr)
					if err != nil {
						t.logger.Log(logging.LevelError, "failed to parse addr", "src", t.ownID, "addr", nodeInfo.Addr, "err", err)
						continue
					}
					err = t.host.Network().ClosePeer(info.ID)
					if err != nil {
						t.logger.Log(logging.LevelError, "Could not close old connection to node", "src", t.ownID, "dst", nodeID, "err", err)
						continue
					}
				} else {
					t.logger.Log(logging.LevelDebug, "No stream for old connection", "src", t.ownID, "dst", nodeID)
				}

				t.nodesLock.Lock()
				delete(t.nodes, nodeID)
				t.nodesLock.Unlock()
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
	for id, addr := range nodes {
		_, found := t.nodes[id]
		if !found {
			t.nodes[id] = &nodeInfo{
				ID:        id,
				Addr:      addr,
				IsOpening: false,
				Stream:    nil,
			}

		}
	}
	t.nodesLock.Unlock()

	t.connWg.Add(1)
	go t.connect(ctx, nodes)
}

func (t *Transport) connect(ctx context.Context, nodes map[types.NodeID]types.NodeAddress) {
	t.logger.Log(logging.LevelDebug, "started connecting nodes")

	defer t.connWg.Done()
	defer func() {
		t.logger.Log(logging.LevelDebug, "finished connecting nodes")
	}()

	wg := &sync.WaitGroup{}
	wg.Add(len(nodes))

	for nodeID := range nodes {
		if nodeID == t.ownID {
			// Do not establish a real connection with own node.
			wg.Done()
			continue
		}

		if t.streamExists(nodeID) {
			t.logger.Log(logging.LevelInfo, "stream to node already exists", "dst", nodeID)
			wg.Done()
			continue
		}

		go t.connectToNode(ctx, nodeID, wg)
	}

	wg.Wait()
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

	addr, found := t.getAddr(node)
	if !found {
		t.logger.Log(logging.LevelError, "failed to get node address", "src", t.ownID, "dst", node)
		return
	}
	info, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		t.logger.Log(logging.LevelError, "failed to parse addr", "addr", addr, "err", err)
		return
	}

	t.host.Peerstore().AddAddrs(info.ID, info.Addrs, PermanentAddrTTL)

	s, err := t.openStream(ctx, info.ID)
	if err != nil {
		t.logger.Log(logging.LevelError, "failed to open stream to node", "addr", addr, "node", node, "err", err)
		return
	}

	t.addOutboundStream(node, s)
	t.logger.Log(logging.LevelDebug, fmt.Sprintf("node %s has connected to node %s", t.ownID.Pb(), node.Pb()))
}

func (t *Transport) openStream(ctx context.Context, p peer.ID) (network.Stream, error) {
	var streamErr error
	for i := 0; i < retryAttempts; i++ {
		sctx, cancel := context.WithTimeout(ctx, maxConnectingTimeout)

		s, streamErr := t.host.NewStream(sctx, p, ProtocolID)
		cancel()

		if streamErr == nil {
			return s, nil
		}

		if i >= noLoggingErrorAttempts {
			t.logger.Log(
				logging.LevelError, fmt.Sprintf("failed to open stream to %s, retry in %s", p, retryTimeout.String()), "src", t.ownID)
		} else {
			t.logger.Log(
				logging.LevelInfo, fmt.Sprintf("failed to open stream to %s, retry in %s", p, retryTimeout.String()), "src", t.ownID)
		}

		delay := time.NewTimer(retryTimeout)
		select {
		case <-delay.C:
			continue
		case <-ctx.Done():
			if !delay.Stop() {
				<-delay.C
			}
			return nil, fmt.Errorf("libp2p opening stream: context closed")
		}
	}
	return nil, fmt.Errorf("failed to open stream to %s: %w", p, streamErr)
}

func (t *Transport) Send(ctx context.Context, dest types.NodeID, payload *messagepb.Message) error {
	var nodeErr *nodeInfoError
	var streamErr *streamInfoError
	s, err := t.getStream(dest)
	if errors.As(err, &nodeErr) {
		return err
	}
	if errors.As(err, &streamErr) {
		t.connWg.Add(1)
		go t.connectToNode(ctx, dest, t.connWg)
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
		w.Flush()
	}
	// If we cannot write to the stream, or we don't have a connection to that peer then try to
	// connect to the peer.
	if err != nil || len(t.host.Network().ConnsToPeer(s.Conn().RemotePeer())) == 0 {
		t.connWg.Add(1)
		go t.connectToNode(ctx, dest, t.connWg)
		return fmt.Errorf("failed to write data to stream, reopening")
	}

	return nil
}

func (t *Transport) mirHandler(s network.Stream) {
	t.logger.Log(logging.LevelDebug, "mir handler started", "streamID", s.ID())
	defer t.logger.Log(logging.LevelDebug, "mir handler stopped", "streamID", s.ID())

	defer func() {
		t.logger.Log(logging.LevelDebug, "mir handler is closing stream", "streamID", s.ID())
		err := s.Close()
		if err != nil {
			t.logger.Log(logging.LevelError, "closing stream", "streamID", s.ID(), "err", err)
		}
	}() // nolint

	for {
		var msg TransportMessage
		err := msg.UnmarshalCBOR(s)
		if err != nil {
			t.logger.Log(logging.LevelError, "failed to read mir transport request", "err", err)
			return
		}

		var payload messagepb.Message

		if err := proto.Unmarshal(msg.Payload, &payload); err != nil {
			t.logger.Log(logging.LevelError, "failed to unmarshall mir transport request", "err", err)
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

	v, found := t.nodes[nodeID]
	if !found {
		return
	}
	v.Stream = s
}

func (t *Transport) streamExists(nodeID types.NodeID) bool {
	t.nodesLock.RLock()
	defer t.nodesLock.RUnlock()

	v, found := t.nodes[nodeID]
	return found && v.Stream != nil
}

func (t *Transport) getStream(nodeID types.NodeID) (network.Stream, error) {
	t.nodesLock.RLock()
	defer t.nodesLock.RUnlock()

	_, found := t.nodes[nodeID]
	if !found {
		return nil, &nodeInfoError{nodeID}
	}

	v, found := t.nodes[nodeID]
	if !found || v.Stream == nil {
		return nil, &streamInfoError{nodeID}
	}
	return v.Stream, nil
}

func (t *Transport) getAddr(nodeID types.NodeID) (types.NodeAddress, bool) {
	t.nodesLock.RLock()
	defer t.nodesLock.RUnlock()

	v, found := t.nodes[nodeID]
	if !found {
		return nil, false
	}
	return v.Addr, true
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
					if err := t.Send(ctx, types.NodeID(destID), e.SendMessage.Msg); err != nil { // nolint
						// TODO: Handle sending errors (and remove "nolint" comment above).
						//       Also, this violates the non-blocking operation of ApplyEvents method. Fix it.
					}
				}
			}
		default:
			return fmt.Errorf("unexpected type of Net event: %T", event.Type)
		}
	}

	return nil
}
