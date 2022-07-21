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
	"google.golang.org/protobuf/proto"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	mirnet "github.com/filecoin-project/mir/pkg/net"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	"github.com/filecoin-project/mir/pkg/types"
)

const (
	ID                = "/mir/0.0.1"
	defaultMaxTimeout = 300 * time.Millisecond
	PermanentAddrTTL  = math.MaxInt64 - iota
)

type TransportMessage struct {
	Sender  string
	Payload []byte
}

var _ mirnet.Transport = &Transport{}

type Transport struct {
	host              host.Host
	ownID             types.NodeID
	membership        map[types.NodeID]multiaddr.Multiaddr
	incomingMessages  chan *events.EventList
	outboundStreamsMx sync.Mutex
	outboundStreams   map[types.NodeID]network.Stream
	logger            logging.Logger
}

func NewTransport(h host.Host, membership map[types.NodeID]multiaddr.Multiaddr, ownID types.NodeID, logger logging.Logger) *Transport {
	if logger == nil {
		logger = logging.ConsoleErrorLogger
	}
	return &Transport{
		incomingMessages: make(chan *events.EventList),
		outboundStreams:  make(map[types.NodeID]network.Stream),
		logger:           logger,
		membership:       membership,
		ownID:            ownID,
		host:             h,
	}
}

func (t *Transport) ImplementsModule() {}

func (t *Transport) EventsOut() <-chan *events.EventList {
	return t.incomingMessages
}

func (t *Transport) Start() error {
	t.logger.Log(logging.LevelDebug, fmt.Sprintf("node %s handler starting on %v", t.ownID, t.host.Addrs()))
	t.host.SetStreamHandler(ID, t.mirHandler)
	return nil
}

func (t *Transport) Stop() {
	t.logger.Log(logging.LevelDebug, "Stopping libp2p transport.")
	defer t.logger.Log(logging.LevelDebug, "libp2p transport stopped.")

	t.outboundStreamsMx.Lock()
	defer t.outboundStreamsMx.Unlock()

	for id, s := range t.outboundStreams {
		if s == nil {
			continue
		}
		t.logger.Log(logging.LevelDebug, "Closing connection", "to", id)
		if err := s.Close(); err != nil {
			t.logger.Log(logging.LevelWarn, fmt.Sprintf("Could not close connection to node %v: %v", id, err))
			continue
		}
		t.logger.Log(logging.LevelDebug, "Closed connection", "to", id)
	}

	if err := t.host.Close(); err != nil {
		t.logger.Log(logging.LevelError, fmt.Sprintf("Could not close libp2p %v: %v", t.ownID, err))
	} else {
		t.logger.Log(logging.LevelDebug, fmt.Sprintf("node %v libp2p host closed", t.ownID))
	}
}

func (t *Transport) Connect(ctx context.Context) {
	wg := sync.WaitGroup{}
	wg.Add(len(t.membership) - 1)

	for nodeID, nodeAddr := range t.membership {
		if nodeID == t.ownID {
			continue
		}

		go func(nodeID types.NodeID, nodeAddr multiaddr.Multiaddr) {
			defer wg.Done()

			t.logger.Log(logging.LevelDebug, fmt.Sprintf("node %s is connecting to node %s", t.ownID.Pb(), nodeID.Pb()))

			// Extract the peer ID from the multiaddr.
			info, err := peer.AddrInfoFromP2pAddr(nodeAddr)
			if err != nil {
				t.logger.Log(logging.LevelError,
					fmt.Sprintf("failed to parse addr %v for node %v: %v", nodeAddr, nodeID, err))
				return
			}

			t.host.Peerstore().AddAddrs(info.ID, info.Addrs, PermanentAddrTTL)

			s, err := t.openStream(ctx, info.ID)
			if err != nil {
				t.logger.Log(logging.LevelError, fmt.Sprintf("couldn't open stream: %v", err))
				return
			}
			t.addOutboundStream(nodeID, s)
			t.logger.Log(logging.LevelDebug, fmt.Sprintf("node %s has connected to node %s", t.ownID.Pb(), nodeID.Pb()))

		}(nodeID, nodeAddr)
	}

	wg.Wait()
}

func (t *Transport) openStream(ctx context.Context, p peer.ID) (network.Stream, error) {
	for {
		sctx, cancel := context.WithTimeout(ctx, defaultMaxTimeout)
		s, err := t.host.NewStream(sctx, p, ID)
		cancel()

		if err == nil {
			return s, nil
		}

		t.logger.Log(logging.LevelError, fmt.Sprintf("failed to open stream: %v", err))

		delay := time.NewTimer(defaultMaxTimeout)

		select {
		case <-delay.C:
			continue
		case <-ctx.Done():
			if !delay.Stop() {
				<-delay.C
			}
			return nil, fmt.Errorf("context closed")
		}
	}
}

func (t *Transport) Send(dest types.NodeID, payload *messagepb.Message) error {
	outBytes, err := proto.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	t.outboundStreamsMx.Lock()
	s, ok := t.outboundStreams[dest]
	t.outboundStreamsMx.Unlock()
	if !ok {
		return fmt.Errorf("failed to get stream for node %v", dest)
	}

	w := bufio.NewWriter(bufio.NewWriter(s))

	msg := TransportMessage{
		t.ownID.Pb(),
		outBytes,
	}

	buf := new(bytes.Buffer)
	err = msg.MarshalCBOR(buf)
	if err != nil {
		return err
	}

	_, err = w.Write(buf.Bytes())
	if err == nil {
		w.Flush()
	}
	if err != nil {
		return err
	}

	return nil
}

func (t *Transport) mirHandler(s network.Stream) {
	t.logger.Log(logging.LevelDebug, fmt.Sprintf("mir handler for %s started", s.ID()))
	defer t.logger.Log(logging.LevelDebug, fmt.Sprintf("mir handler for %s stopped", s.ID()))

	defer s.Close() // nolint

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

		t.logger.Log(logging.LevelDebug, "sent to channel", "msg type=", fmt.Sprintf("%T", payload.Type))
	}
}

func (t *Transport) addOutboundStream(nodeID types.NodeID, s network.Stream) {
	t.outboundStreamsMx.Lock()
	defer t.outboundStreamsMx.Unlock()

	if _, found := t.outboundStreams[nodeID]; found {
		t.logger.Log(logging.LevelWarn, fmt.Sprintf("stream to %s already extists\n", nodeID.Pb()))
	}
	t.outboundStreams[nodeID] = s
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
