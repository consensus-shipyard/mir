package stdevents

import (
	"fmt"

	"github.com/filecoin-project/mir/stdtypes"
)

type serializableSendMessage struct {
	mirEvent
	DestNodes        []stdtypes.NodeID
	RemoteDestModule stdtypes.ModuleID
	Payload          stdtypes.RawMessage
}

func (ssm *serializableSendMessage) SendMessage() *SendMessage {
	return &SendMessage{
		mirEvent:         ssm.mirEvent,
		DestNodes:        ssm.DestNodes,
		RemoteDestModule: ssm.RemoteDestModule,
		Payload:          ssm.Payload,
	}
}

type SendMessage struct {
	mirEvent
	DestNodes        []stdtypes.NodeID
	RemoteDestModule stdtypes.ModuleID
	Payload          stdtypes.Message
}

func (e *SendMessage) serializable() (*serializableSendMessage, error) {
	data, err := e.Payload.ToBytes()
	if err != nil {
		return nil, err
	}

	return &serializableSendMessage{
		mirEvent:         e.mirEvent,
		DestNodes:        e.DestNodes,
		RemoteDestModule: e.RemoteDestModule,
		Payload:          data,
	}, nil
}

func NewSendMessage(
	message stdtypes.Message,
	localDestModule stdtypes.ModuleID,
	remoteDestModule stdtypes.ModuleID,
	destNodes ...stdtypes.NodeID,
) *SendMessage {
	return &SendMessage{
		mirEvent:         mirEvent{DestModule: localDestModule},
		DestNodes:        destNodes,
		RemoteDestModule: remoteDestModule,
		Payload:          message,
	}
}

func NewSendMessageWithSrc(
	srcModule stdtypes.ModuleID,
	message stdtypes.Message,
	localDestModule stdtypes.ModuleID,
	remoteDestModule stdtypes.ModuleID,
	destNodes ...stdtypes.NodeID,
) *SendMessage {
	e := NewSendMessage(message, localDestModule, remoteDestModule, destNodes...)
	e.SrcModule = srcModule
	return e
}

func (e *SendMessage) NewSrc(newSrc stdtypes.ModuleID) stdtypes.Event {
	newE := *e
	e.SrcModule = newSrc
	return &newE
}

func (e *SendMessage) NewDest(newDest stdtypes.ModuleID) stdtypes.Event {
	newE := *e
	e.DestModule = newDest
	return &newE
}

func (e *SendMessage) ToBytes() ([]byte, error) {
	serializable, err := e.serializable()
	if err != nil {
		return nil, err
	}

	return serialize(serializable)
}

func (e *SendMessage) ToString() string {
	data, err := e.ToBytes()
	if err != nil {
		return fmt.Sprintf("unmarshalableEvent(%+v)", e)
	}

	return string(data)
}
