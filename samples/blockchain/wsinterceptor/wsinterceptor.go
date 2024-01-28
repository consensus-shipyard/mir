package wsinterceptor

import (
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/samples/blockchain/wsinterceptor/wsserver.go"
	"google.golang.org/protobuf/proto"
)

type eventFilterFn func(*eventpb.Event) bool

type wsInterceptor struct {
	server      *wsserver.WsServer
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
			i.server.SendChan <- wsserver.WsMessage{
				MessageType: 2,
				Payload:     []byte(payload),
			}
		}
	}

	return nil
}

func NewWsInterceptor(eventFilter eventFilterFn, port int, logger logging.Logger) *wsInterceptor {
	server := wsserver.NewWsServer(port, logger)
	logger.Log(logging.LevelInfo, "Starting websocker interceptor server")
	go server.StartServers()
	return &wsInterceptor{
		server:      server,
		eventFilter: eventFilter,
		logger:      logger,
	}
}
