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
	"github.com/multiformats/go-multiaddr"
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

type Transport struct {
	host                    host.Host
	ownID                   types.NodeID
	connWg                  *sync.WaitGroup
	incomingMessages        chan *events.EventList
	outboundStreamsMx       sync.Mutex
	outboundStreams         map[types.NodeID]network.Stream
	connectionsInProgress   map[types.NodeID]bool
	connectionsInProgressMx sync.Mutex
	logger                  logging.Logger
}

func NewTransport(h host.Host, ownID types.NodeID, logger logging.Logger) (*Transport, error) {
	if logger == nil {
		logger = logging.ConsoleErrorLogger
	}

	return &Transport{
		connWg:                &sync.WaitGroup{},
		incomingMessages:      make(chan *events.EventList),
		outboundStreams:       make(map[types.NodeID]network.Stream),
		connectionsInProgress: make(map[types.NodeID]bool),
		logger:                logger,
		ownID:                 ownID,
		host:                  h,
	}, nil
}

func (t *Transport) ImplementsModule() {}

func (t *Transport) EventsOut() <-chan *events.EventList {
	return t.incomingMessages
}

func (t *Transport) Start() error {
	t.logger.Log(logging.LevelDebug, fmt.Sprintf("node %s handler starting on %v", t.ownID, t.host.Addrs()))
	t.host.SetStreamHandler(ProtocolID, t.mirHandler)
	return nil
}

func (t *Transport) Stop() {
	t.logger.Log(logging.LevelDebug, "Stopping libp2p transport.")
	defer t.logger.Log(logging.LevelDebug, "Stopping libp2p transport finished.")

	t.outboundStreamsMx.Lock()
	defer t.outboundStreamsMx.Unlock()

	for id, s := range t.outboundStreams {
		if s == nil {
			continue
		}
		t.logger.Log(logging.LevelDebug, "Closing connection", "to", id)

		if err := s.Close(); err != nil {
			t.logger.Log(logging.LevelError, fmt.Sprintf("Could not close connection to node %v: %v", id, err))
			continue
		}

		t.logger.Log(logging.LevelDebug, "Closed connection", "to", id)
	}

	t.host.RemoveStreamHandler(ProtocolID)

	if err := t.host.Close(); err != nil {
		t.logger.Log(logging.LevelError, fmt.Sprintf("Could not close libp2p %v: %v", t.ownID, err))
	} else {
		t.logger.Log(logging.LevelDebug, fmt.Sprintf("node %v libp2p host closed", t.ownID))
	}

	t.connWg.Wait()
}

func (t *Transport) CloseOldConnections(ctx context.Context, nextNodes map[types.NodeID]types.NodeAddress) {
	t.outboundStreamsMx.Lock()
	defer t.outboundStreamsMx.Unlock()

	for id, s := range t.outboundStreams {
		if s == nil {
			continue
		}

		// Close an old connection to a node if we don't need to connect to this node further.
		if _, newConn := nextNodes[id]; !newConn {
			t.logger.Log(logging.LevelDebug, "Closing old connection", "to", id)

			if err := s.Close(); err != nil {
				t.logger.Log(logging.LevelError, fmt.Sprintf("Could not close old connection to node %v: %v", id, err))
				continue
			}

			// If we had that connection then remove it.
			delete(t.outboundStreams, id)

			t.logger.Log(logging.LevelDebug, "Closed old connection", "to", id)
		}
	}
}

func (t *Transport) Connect(ctx context.Context, nodes map[types.NodeID]types.NodeAddress) {
	if len(nodes) == 0 {
		t.logger.Log(logging.LevelWarn, "no nodes to connect to")
		return
	}
	t.connWg.Add(1)
	go t.connect(ctx, nodes)
}

func (t *Transport) syncConnect(ctx context.Context, nodes map[types.NodeID]types.NodeAddress) {
	if len(nodes) == 0 {
		t.logger.Log(logging.LevelWarn, "no nodes to connect to")
		return
	}
	t.connWg.Add(1)
	t.connect(ctx, nodes)
}

func (t *Transport) connect(ctx context.Context, nodes map[types.NodeID]types.NodeAddress) {
	t.logger.Log(logging.LevelDebug, "started sync connecting nodes")

	defer t.connWg.Done()
	defer func() {
		t.logger.Log(logging.LevelDebug, "finished sync connecting nodes")
	}()

	wg := &sync.WaitGroup{}
	wg.Add(len(nodes))

	for nodeID, nodeAddr := range nodes {
		if nodeID == t.ownID {
			// Do not establish a real connection with own node.
			wg.Done()
			continue
		}

		if t.streamExists(nodeID) {
			wg.Done()
			t.logger.Log(logging.LevelInfo, fmt.Sprintf("stream to %s already exists", nodeID.Pb()))
			continue
		}

		go t.connectToNode(ctx, nodeID, nodeAddr, wg)
	}

	wg.Wait()
}

func (t *Transport) reconnect(ctx context.Context, nodes map[types.NodeID]types.NodeAddress) {
	t.logger.Log(logging.LevelDebug, "started reconnecting")

	defer t.connWg.Done()
	defer func() {
		t.logger.Log(logging.LevelDebug, "finished reconnecting")
	}()

	wg := &sync.WaitGroup{}
	wg.Add(len(nodes))

	for nodeID, nodeAddr := range nodes {
		if nodeID == t.ownID {
			// Do not establish a real connection with own node.
			wg.Done()
			continue
		}

		go t.connectToNode(ctx, nodeID, nodeAddr, wg)
	}

	wg.Wait()
}

func (t *Transport) connectToNode(ctx context.Context, id types.NodeID, addr multiaddr.Multiaddr, wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Println(id, addr)

	if t.isConnectionToNodeInProgress(id) {
		t.logger.Log(logging.LevelDebug, fmt.Sprintf("connecting to node %s in progress", id.Pb()))
		return
	}

	t.addConnectionInProgress(id)
	defer t.removeConnectionInProgress(id)

	t.logger.Log(logging.LevelDebug, fmt.Sprintf("node %s is connecting to node %s", t.ownID.Pb(), id.Pb()))

	info, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		t.logger.Log(logging.LevelError, fmt.Sprintf("failed to parse addr %v: %v", addr, err))
		return
	}

	t.host.Peerstore().AddAddrs(info.ID, info.Addrs, PermanentAddrTTL)

	s, err := t.openStream(ctx, info.ID)
	if err != nil {
		t.removeConnectionInProgress(id)
		t.logger.Log(logging.LevelError, "failed to open stream to node %v: %w", addr, err)
		return
	}

	t.addOutboundStream(id, s)
	t.logger.Log(logging.LevelDebug, fmt.Sprintf("node %s has connected to node %s", t.ownID.Pb(), id.Pb()))
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
				logging.LevelError, fmt.Sprintf("failed to open stream to %s, retry in %s", p, retryTimeout.String()))
		} else {
			t.logger.Log(
				logging.LevelInfo, fmt.Sprintf("failed to open stream to %s, retry in %s", p, retryTimeout.String()))
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

func (t *Transport) Send(dest types.NodeID, payload *messagepb.Message) error {
	t.outboundStreamsMx.Lock()
	s, ok := t.outboundStreams[dest]
	t.outboundStreamsMx.Unlock()
	if !ok {
		return errors.Wrap(mirnet.ErrNoStreamForDest, dest.Pb())
	}

	outBytes, err := proto.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	msg := TransportMessage{
		t.ownID.Pb(),
		outBytes,
	}

	buf := new(bytes.Buffer)
	err = msg.MarshalCBOR(buf)
	if err != nil {
		return err
	}

	w := bufio.NewWriter(s)
	_, err = w.Write(buf.Bytes())
	if err == nil {
		w.Flush()
	}
	if err != nil || len(t.host.Network().ConnsToPeer(s.Conn().RemotePeer())) == 0 {
		t.connWg.Add(1)
		addr, err := multiaddr.NewMultiaddr(fmt.Sprintf("%s/p2p/%s", s.Conn().RemoteMultiaddr(), s.Conn().RemotePeer()))
		if err != nil {
			return err
		}
		t.reconnect(context.TODO(), map[types.NodeID]types.NodeAddress{dest: addr})
		return errors.Wrapf(mirnet.ErrWritingFailed, "%v", err)
	}

	return nil
}

func (t *Transport) mirHandler(s network.Stream) {
	t.logger.Log(logging.LevelDebug, fmt.Sprintf("mir handler for %s started", s.ID()))
	defer t.logger.Log(logging.LevelDebug, fmt.Sprintf("mir handler for %s stopped", s.ID()))

	defer func() {
		t.logger.Log(logging.LevelDebug, fmt.Sprintf("mir handler closing stream for %v", s.ID()))
		err := s.Close()
		if err != nil {
			t.logger.Log(logging.LevelError, fmt.Sprintf("closing stream for %v", s.ID()))
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
	t.outboundStreamsMx.Lock()
	defer t.outboundStreamsMx.Unlock()

	t.outboundStreams[nodeID] = s
}

func (t *Transport) streamExists(nodeID types.NodeID) bool {
	t.outboundStreamsMx.Lock()
	defer t.outboundStreamsMx.Unlock()

	_, found := t.outboundStreams[nodeID]
	return found
}

func (t *Transport) isConnectionToNodeInProgress(nodeID types.NodeID) bool {
	t.connectionsInProgressMx.Lock()
	defer t.connectionsInProgressMx.Unlock()

	_, found := t.connectionsInProgress[nodeID]
	return found
}

func (t *Transport) addConnectionInProgress(nodeID types.NodeID) {
	t.connectionsInProgressMx.Lock()
	defer t.connectionsInProgressMx.Unlock()

	t.connectionsInProgress[nodeID] = true
}

func (t *Transport) removeConnectionInProgress(nodeID types.NodeID) {
	t.connectionsInProgressMx.Lock()
	defer t.connectionsInProgressMx.Unlock()

	delete(t.connectionsInProgress, nodeID)
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
					if err := t.Send(types.NodeID(destID), e.SendMessage.Msg); err != nil { // nolint
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
