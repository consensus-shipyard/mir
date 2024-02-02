package wsInterceptor

import (
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	wsServer "github.com/filecoin-project/mir/samples/blockchain/wsinterceptor/wsServer.go"
	"google.golang.org/protobuf/proto"
)

/**
 * Websocket interceptor
 * =====================
 *
 * The websocket interceptor intercepts all events and sends them to a websocket server.
 * Any connected client can then receive these events by subscribing to the websocket server.
 *
 * The interceptor proto defines events which are specificly intended for the interceptor and not used by the actual blockchain.
 * Since these events don't have a destination module, they are sent to the "devnull" module.
 * However, all events are intercepted and sent to the websocket server. The interceptor proto is simply for "extra" events.
 *
 * The interceptor proto defines two such events:
 * - TreeUpdate: This event is sent by the blockchain manager (BCM) when the blockchain is updated. It contains all blocks in the blockchain and the id of the new head.
 * - StateUpdate: This event is sent by the application when it computes the state for the newest head of the blockchain.
 */

type eventFilterFn func(*eventpb.Event) bool

type wsInterceptor struct {
	server      *wsServer.WsServer
	eventFilter eventFilterFn
	logger      logging.Logger
}

func (i *wsInterceptor) Intercept(events *events.EventList) error {
	for _, e := range events.Slice() {
		if i.eventFilter(e) {
			payload, err := proto.Marshal(e)
			if err != nil {
				// log?
				return err
			}
			i.server.SendChan <- wsServer.WsMessage{
				MessageType: 2,
				Payload:     []byte(payload),
			}
		}
	}

	return nil
}

func NewWsInterceptor(eventFilter eventFilterFn, port int, logger logging.Logger) *wsInterceptor {
	server := wsServer.NewWsServer(port, logger)
	logger.Log(logging.LevelInfo, "Starting websocker interceptor server")
	go server.StartServers()
	return &wsInterceptor{
		server:      server,
		eventFilter: eventFilter,
		logger:      logger,
	}
}
