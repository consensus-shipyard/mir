package grpc

import (
	es "github.com/go-errors/errors"

	messagepbtypes "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	transportpbevents "github.com/filecoin-project/mir/pkg/pb/transportpb/events"
	"github.com/filecoin-project/mir/stdevents"
	"github.com/filecoin-project/mir/stdtypes"
)

// selfConnection represents a connection of a node to itself.
// It bypasses the network completely and feeds sent messages directly into the code that handles message delivery.
type selfConnection struct {
	ownAddr     string
	msgBuffer   chan *GrpcMessage
	deliverChan chan<- *stdtypes.EventList
	stop        chan struct{}
	done        chan struct{}
}

// newSelfConnection returns a connection to self.
// Addr is the own address and deliverChan is the channel to which delivered messages need to be written.
// Messages sent to this connection will be buffered and eventually written to deliverChan in form of MessageDelivered
// events (unless the buffer fills up, in which case sent messages will be dropped.)
func newSelfConnection(params Params, ownAddr string, deliverChan chan<- *stdtypes.EventList) (*selfConnection, error) {
	conn := &selfConnection{
		ownAddr:     ownAddr,
		msgBuffer:   make(chan *GrpcMessage, params.ConnectionBufferSize),
		deliverChan: deliverChan,
		stop:        make(chan struct{}),
		done:        make(chan struct{}),
	}

	go conn.process()

	return conn, nil
}

// Address returns the network address of the node itself, as it is itself on the other side of the connection.
func (conn *selfConnection) Address() string {
	return conn.ownAddr
}

// Send feeds the given message directly to the sink of delivered messages.
// Send is non-blocking and if the buffer for delivered messages is full, the message is dropped.
func (conn *selfConnection) Send(msg *GrpcMessage) error {

	select {
	case conn.msgBuffer <- msg:
		return nil
	default:
		return es.Errorf("send buffer full")
	}

}

// Close makes the connection stop processing messages. No messages will be delivered by it after Close returns.
func (conn *selfConnection) Close() {

	// Do nothing if connection already has been closed.
	select {
	case <-conn.stop:
		return
	default:
	}

	// Stop processing and wait until it finishes.
	close(conn.stop)
	<-conn.done
}

// Wait returns a channel and a function.
// Since this is a connection to self, there is no underlying network stream to wait for.
// If the connection is closed, the channel will contain a single error value written to it and will never be closed.
// Otherwise, the channel will be closed without any values.
// The returned function has no effect.
func (conn *selfConnection) Wait() (chan error, func()) {
	result := make(chan error, 1)
	select {
	case <-conn.stop:
		result <- es.Errorf("connection closed")
	default:
		close(result)
	}
	return result, func() {}
}

// process shovels messages from the message buffer to the deliver channel.
func (conn *selfConnection) process() {
	// When done, make the Close method return.
	defer close(conn.done)

	for {
		select {
		case <-conn.stop:
			return
		case grpcMsg := <-conn.msgBuffer:

			var rcvEvent stdtypes.Event
			var destModule stdtypes.ModuleID
			switch msg := grpcMsg.Type.(type) {
			case *GrpcMessage_PbMsg:
				rcvEvent = transportpbevents.MessageReceived(
					destModule,
					stdtypes.NodeID(grpcMsg.Sender),
					messagepbtypes.MessageFromPb(msg.PbMsg),
				).Pb()
			case *GrpcMessage_RawMsg:
				destModule = stdtypes.ModuleID(msg.RawMsg.DestModule)
				rcvEvent = stdevents.NewMessageReceived(
					destModule,
					stdtypes.NodeID(grpcMsg.Sender),
					stdtypes.RawMessage(msg.RawMsg.Data),
				)
			}

			select {
			case <-conn.stop:
				return
			case conn.deliverChan <- stdtypes.ListOf(rcvEvent):
				// Nothing to do in this case, message has been delivered.
			}
		}
	}
}
