package libp2ptransport

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/multiformats/go-multiaddr"
	"google.golang.org/protobuf/proto"

	"github.com/filecoin-project/mir/pkg/events"
	grpc "github.com/filecoin-project/mir/pkg/grpctransport"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	"github.com/filecoin-project/mir/pkg/types"
)

const TransportProtocolID = "/mir/0.0.1"

type TransportMessage struct {
	Payload []byte
}

type Transport struct {
	host             host.Host
	ownID            types.NodeID
	membership       map[types.NodeID]string
	incomingMessages chan *events.EventList
	streamsMx        sync.Mutex
	streams          map[types.NodeID]*bufio.ReadWriter
	logger           logging.Logger
}

func New(h host.Host, membership map[types.NodeID]string, ownID types.NodeID, logger logging.Logger) *Transport {
	if logger == nil {
		logger = logging.ConsoleErrorLogger
	}

	return &Transport{
		incomingMessages: make(chan *events.EventList),
		streams:          make(map[types.NodeID]*bufio.ReadWriter),
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

func (t *Transport) Start(ctx context.Context) error {
	t.logger.Log(logging.LevelDebug, "starting on", "addr", t.host.Addrs())
	t.host.SetStreamHandler(TransportProtocolID, t.streamHandler)
	return nil
}

func (t *Transport) Connect(ctx context.Context) error {
	wg := sync.WaitGroup{}

	for nodeID, nodeAddr := range t.membership {
		if nodeID == t.ownID {
			continue
		}
		if strings.Compare(nodeID.Pb(), t.ownID.Pb()) == -1 {
			continue
		}

		wg.Add(1)
		go func(nodeID types.NodeID, nodeAddr string) {
			defer wg.Done()

			t.logger.Log(logging.LevelDebug, fmt.Sprintf("%s is connecting to %s\n", t.ownID.Pb(), nodeID.Pb()))

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

			var stream network.Stream

			next := time.NewTicker(700 * time.Millisecond)
			defer next.Stop()

			connected := false
			attempt := 0
			for !connected {
				select {
				case <-next.C:
					stream, err = t.host.NewStream(ctx, info.ID, TransportProtocolID)
					if err != nil {
						t.logger.Log(logging.LevelError,
							fmt.Sprintf("attempt %d to open stream to peer %v: %v", attempt, info.ID, err))
						attempt++
						continue
					}

					t.logger.Log(logging.LevelInfo, fmt.Sprintf("%s opened a stream to %s\n", t.ownID.Pb(), nodeID.Pb()))
					t.addStream(nodeID, stream)
					connected = true

					go t.receiver(t.streams[nodeID])
				}
			}
		}(nodeID, nodeAddr)
	}

	wg.Wait()

	return nil
}

func (t *Transport) Send(dest types.NodeID, msg *messagepb.Message) error {
	payload := grpc.GrpcMessage{
		Sender: t.ownID.Pb(),
		Msg:    msg,
	}

	bs, err := proto.Marshal(&payload)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	rw, ok := t.streams[dest]
	if !ok {
		return fmt.Errorf("failed to get stream for node %v", dest)
	}

	mb := TransportMessage{
		bs,
	}

	buf := new(bytes.Buffer)
	err = mb.MarshalCBOR(buf)
	if err != nil {
		return err
	}

	_, err = rw.Write(buf.Bytes())
	if err == nil {
		rw.Flush()
	}
	if err != nil {
		return err
	}

	return err
}

func (t *Transport) getIDByPeerID(peerID string) types.NodeID {
	for k, v := range t.membership {
		if strings.Contains(v, peerID) {
			return k
		}
	}
	panic("ID was not found by peer ID")
}

func (t *Transport) receiver(rw *bufio.ReadWriter) {
	for {
		var req TransportMessage
		err := req.UnmarshalCBOR(rw)
		if err != nil {
			t.logger.Log(logging.LevelError, "failed to read Mir transport request", "err", err)
			return
		}

		var grpcMsg grpc.GrpcMessage

		if err := proto.Unmarshal(req.Payload, &grpcMsg); err != nil {
			t.logger.Log(logging.LevelError, "failed to unmarshall Mir transport request", "err", err)
			return
		}

		t.incomingMessages <- events.ListOf(
			events.MessageReceived(types.ModuleID(grpcMsg.Msg.DestModule), types.NodeID(grpcMsg.Sender), grpcMsg.Msg),
		)
	}
}

func (t *Transport) streamHandler(stream network.Stream) {
	t.logger.Log(logging.LevelDebug, "stream handler started")
	defer t.logger.Log(logging.LevelDebug, "stream handler stopped")

	nodeID := t.getIDByPeerID(stream.Conn().RemotePeer().String())
	t.addStream(nodeID, stream)
	go t.receiver(t.streams[nodeID])
}

func (t *Transport) addStream(nodeID types.NodeID, stream network.Stream) {
	t.streamsMx.Lock()
	defer t.streamsMx.Unlock()

	_, found := t.streams[nodeID]
	if found {
		t.logger.Log(logging.LevelWarn, fmt.Sprintf("stream to %s already extists\n", nodeID.Pb()))
	}
	t.streams[nodeID] = bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
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
