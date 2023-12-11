package events

import (
	"fmt"

	t "github.com/filecoin-project/mir/pkg/types"
)

type serializableMessageReceived struct {
	mirEvent
	Sender  t.NodeID
	Payload RawMessage
}

func (smr *serializableMessageReceived) MessageReceived() *MessageReceived {
	return &MessageReceived{
		mirEvent: smr.mirEvent,
		Sender:   smr.Sender,
		Payload:  smr.Payload,
	}
}

type MessageReceived struct {
	mirEvent
	Sender  t.NodeID
	Payload Message
}

func (e *MessageReceived) serializable() (*serializableMessageReceived, error) {
	data, err := e.Payload.ToBytes()
	if err != nil {
		return nil, err
	}

	return &serializableMessageReceived{
		mirEvent: e.mirEvent,
		Sender:   e.Sender,
		Payload:  data,
	}, nil
}

func NewMessageReceived(dest t.ModuleID, sender t.NodeID, payload Message) *MessageReceived {
	return &MessageReceived{
		mirEvent: mirEvent{DestModule: dest},
		Sender:   sender,
		Payload:  payload,
	}
}

func NewMessageReceivedWithSrc(src t.ModuleID, dest t.ModuleID, sender t.NodeID, payload Message) *MessageReceived {
	e := NewMessageReceived(dest, sender, payload)
	e.SrcModule = src
	return e
}

func (e *MessageReceived) NewSrc(newSrc t.ModuleID) Event {
	newE := *e
	e.SrcModule = newSrc
	return &newE
}

func (e *MessageReceived) NewDest(newDest t.ModuleID) Event {
	newE := *e
	e.DestModule = newDest
	return &newE
}

func (e *MessageReceived) ToBytes() ([]byte, error) {
	serializable, err := e.serializable()
	if err != nil {
		return nil, err
	}

	return serialize(serializable)
}

func (e *MessageReceived) ToString() string {
	data, err := e.ToBytes()
	if err != nil {
		return fmt.Sprintf("unmarshalableEvent(%+v)", e)
	}

	return string(data)
}
