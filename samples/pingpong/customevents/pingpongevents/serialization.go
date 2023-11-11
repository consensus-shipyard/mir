// This code should eventually be generated from the event definitions in events.go

package pingpongevents

import (
	"encoding/json"
	"fmt"

	es "github.com/go-errors/errors"
)

// TODO: When code is generated, this probably won't be needed and the ToBytes() methods can directly be generated
// for each type separately.
// In this case, it would be
// func (p *Ping) ToBytes() {...}
// func (p *Pong) ToBytes() {...}
// and the Whole PingPongMessage wrapper can be removed.
// The code using these events would just directly use pingpongevents.Ping{...} instead of pingpongevents.Message(pingpongevents.Ping{...})
type PingPongMessage struct {
	msg any
}

func Message(payload any) *PingPongMessage {
	return &PingPongMessage{msg: payload}
}

type serializedMessage struct {
	MsgType string
	MsgData []byte
}

func (ppm *PingPongMessage) ToBytes() ([]byte, error) {
	msgData, err := json.Marshal(ppm.msg)
	if err != nil {
		return nil, es.Errorf("could not marshal message data")
	}

	// TODO: Once this code is generated, add a check here and return an error if an unknown type is being serialized.
	//   Concretely here, check that the type of ppm.msg is Ping or Pong.
	message := serializedMessage{
		MsgType: fmt.Sprintf("%T", ppm.msg),
		MsgData: msgData,
	}

	data, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func ParseMessage(data []byte) (any, error) {
	var msg serializedMessage

	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}

	var innerMessage any
	var err error
	switch msg.MsgType {
	case "pingpongevents.Ping", "*pingpongevents.Ping":
		var ping Ping
		err = json.Unmarshal(msg.MsgData, &ping)
		innerMessage = &ping
	case "pingpongevents.Pong", "*pingpongevents.Pong":
		var pong Pong
		err = json.Unmarshal(msg.MsgData, &pong)
		innerMessage = &pong
	default:
		return nil, es.Errorf("unknown message type: %v", msg.MsgType)
	}
	if err != nil {
		return nil, es.Errorf("failed unmarshalling data: %w", err)
	}

	return innerMessage, nil
}
