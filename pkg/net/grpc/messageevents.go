package grpc

import (
	"encoding/json"
	"fmt"

	t "github.com/filecoin-project/mir/pkg/types"
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

func (om *OutgoingMessage) Dest() t.ModuleID {
	return om.LocalDestModule
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
	srcModule t.ModuleID
	dstModule t.ModuleID
	srcNode   t.NodeID
	msgData   []byte
}

func (m *MessageReceived) Src() t.ModuleID {
	return m.srcModule
}

func (m *MessageReceived) Dest() t.ModuleID {
	return m.dstModule
}

func (m *MessageReceived) SrcNode() t.NodeID {
	return m.srcNode
}

func (m *MessageReceived) DestModule() t.ModuleID {
	return m.dstModule
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
	return m.msgData
}
