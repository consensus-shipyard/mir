package grpc

import (
	"encoding/json"
	"fmt"

	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/stdtypes"
)

type Serializable interface {
	ToBytes() ([]byte, error)
}

type OutgoingMessage struct {
	SrcModule        t.ModuleID
	LocalDestModule  t.ModuleID
	DestNodes        []t.NodeID
	RemoteDestModule t.ModuleID
	Data             Serializable
}

func NewOutgoingMessage(msg Serializable, localDestModule t.ModuleID, remoteDestModule t.ModuleID, destinations []t.NodeID) *OutgoingMessage {

	return &OutgoingMessage{
		SrcModule:        "pingpong",
		LocalDestModule:  localDestModule,
		DestNodes:        destinations,
		RemoteDestModule: remoteDestModule,
		Data:             msg,
	}
}

func (om *OutgoingMessage) Src() t.ModuleID {
	return om.SrcModule
}

func (om *OutgoingMessage) NewSrc(newSrc t.ModuleID) stdtypes.Event {
	newOM := *om
	newOM.SrcModule = newSrc
	return &newOM
}

func (om *OutgoingMessage) Dest() t.ModuleID {
	return om.LocalDestModule
}

func (om *OutgoingMessage) NewDest(dest t.ModuleID) stdtypes.Event {
	newOM := *om
	newOM.LocalDestModule = dest
	return &newOM
}

func (om *OutgoingMessage) ToBytes() ([]byte, error) {
	return json.Marshal(om)
}

func (om *OutgoingMessage) ToString() string {
	jsonBytes, err := json.Marshal(om)
	if err != nil {
		return fmt.Sprintf("%+v", *om)
	}

	return string(jsonBytes)
}

type MessageReceived struct {
	SrcModule  t.ModuleID
	DstModule  t.ModuleID
	SourceNode t.NodeID
	MsgData    []byte
}

func (m *MessageReceived) Src() t.ModuleID {
	return m.SrcModule
}

func (m *MessageReceived) NewSrc(newSrc t.ModuleID) stdtypes.Event {
	newEvent := *m
	newEvent.SrcModule = newSrc
	return &newEvent
}

func (m *MessageReceived) Dest() t.ModuleID {
	return m.DstModule
}

func (m *MessageReceived) NewDest(newDest t.ModuleID) stdtypes.Event {
	newEvent := *m
	newEvent.DstModule = newDest
	return &newEvent
}

func (m *MessageReceived) SrcNode() t.NodeID {
	return m.SourceNode
}

func (m *MessageReceived) DestModule() t.ModuleID {
	return m.DstModule
}

func (m *MessageReceived) ToBytes() ([]byte, error) {
	return json.Marshal(m)
}

func (m *MessageReceived) ToString() string {

	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return fmt.Sprintf("%+v", *m)
	}

	return string(jsonBytes)
}

func (m *MessageReceived) Data() []byte {
	return m.MsgData
}
