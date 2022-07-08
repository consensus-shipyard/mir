package libp2ptransport

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/multiformats/go-multiaddr"
	"google.golang.org/protobuf/proto"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	"github.com/filecoin-project/mir/pkg/types"
)

const (
	ID = "/mir/0.0.1"
)

type TransportMessage struct {
	Sender  string
	Payload []byte
}

type Transport struct {
	host              host.Host
	ownID             types.NodeID
	membership        map[types.NodeID]string
	incomingMessages  chan *events.EventList
	outboundStreamsMx sync.Mutex
	outboundStreams   map[types.NodeID]network.Stream
	logger            logging.Logger
}

func New(h host.Host, membership map[types.NodeID]string, ownID types.NodeID, logger logging.Logger) *Transport {
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
	t.logger.Log(logging.LevelDebug, "node transport starting on", "addr", t.host.Addrs())
	t.host.SetStreamHandler(ID, t.mirHandler)
	return nil
}

func (t *Transport) Stop() {
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
}

func (t *Transport) Connect(ctx context.Context) {
	wg := sync.WaitGroup{}

	for nodeID, nodeAddr := range t.membership {
		if nodeID == t.ownID {
			continue
		}

		wg.Add(1)
		go func(nodeID types.NodeID, nodeAddr string) {
			defer wg.Done()

			t.logger.Log(logging.LevelDebug, fmt.Sprintf("%s is connecting to %s", t.ownID.Pb(), nodeID.Pb()))

			maddr, err := multiaddr.NewMultiaddr(nodeAddr)
			if err != nil {
				t.logger.Log(logging.LevelError,
					fmt.Sprintf("failed to parse %v for node %v: %v", nodeAddr, nodeID, err))
			}

			// Extract the peer ID from the multiaddr.
			info, err := peer.AddrInfoFromP2pAddr(maddr)
			if err != nil {
				t.logger.Log(logging.LevelError,
					fmt.Sprintf("failed to parse addr %v for node %v: %v", maddr, nodeID, err))
			}

			t.host.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)

			s, err := t.openStream(ctx, info.ID)
			if err != nil {
				t.logger.Log(logging.LevelError, fmt.Sprintf("couldn't open stream: %v", err))
				return
			}
			t.addOutboundStream(nodeID, s)
		}(nodeID, nodeAddr)
	}

	wg.Wait()
}

func (t *Transport) openStream(ctx context.Context, p peer.ID) (network.Stream, error) {
	timeout := 500 * time.Millisecond
	for {
		sctx, cancel := context.WithTimeout(ctx, timeout)
		s, err := t.host.NewStream(sctx, p, ID)
		cancel()
		if err != nil {
			t.logger.Log(logging.LevelError, err.Error())
			select {
			case <-time.After(timeout):
				continue
			case <-ctx.Done():
				return nil, fmt.Errorf("context closed")
			}
		}
		return s, nil
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

	return err
}

func (t *Transport) mirHandler(s network.Stream) {
	t.logger.Log(logging.LevelDebug, fmt.Sprintf("mir handler for %s started", s.ID()))
	defer t.logger.Log(logging.LevelDebug, fmt.Sprintf("mir handler for %s stopped", s.ID()))

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

	_, found := t.outboundStreams[nodeID]
	if found {
		t.logger.Log(logging.LevelWarn, fmt.Sprintf("stream to %s already extists\n", nodeID.Pb()))
	}
	t.outboundStreams[nodeID] = s
}

func (t *Transport) ApplyEvents(
	ctx context.Context,
	eventList *events.EventList,
) error {
	iter := eventList.Iterator()
	for event := iter.Next(); event != nil; event = iter.Next() {
		switch e := event.Type.(type) {
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
