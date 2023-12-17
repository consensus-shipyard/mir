package grpc

import (
	"encoding/json"
	"fmt"

	"github.com/filecoin-project/mir/stdtypes"
)

type Serializable interface {
	ToBytes() ([]byte, error)
}

type OutgoingMessage struct {
	SrcModule        stdtypes.ModuleID
	LocalDestModule  stdtypes.ModuleID
	DestNodes        []stdtypes.NodeID
	RemoteDestModule stdtypes.ModuleID
	Data             Serializable
}

func NewOutgoingMessage(msg Serializable, localDestModule stdtypes.ModuleID, remoteDestModule stdtypes.ModuleID, destinations []stdtypes.NodeID) *OutgoingMessage {

	return &OutgoingMessage{
		SrcModule:        "pingpong",
		LocalDestModule:  localDestModule,
		DestNodes:        destinations,
		RemoteDestModule: remoteDestModule,
		Data:             msg,
	}
}

func (om *OutgoingMessage) Src() stdtypes.ModuleID {
	return om.SrcModule
}

func (om *OutgoingMessage) NewSrc(newSrc stdtypes.ModuleID) stdtypes.Event {
	newOM := *om
	newOM.SrcModule = newSrc
	return &newOM
}

func (om *OutgoingMessage) Dest() stdtypes.ModuleID {
	return om.LocalDestModule
}

func (om *OutgoingMessage) NewDest(dest stdtypes.ModuleID) stdtypes.Event {
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
	SrcModule  stdtypes.ModuleID
	DstModule  stdtypes.ModuleID
	SourceNode stdtypes.NodeID
	MsgData    []byte
}

func (m *MessageReceived) Src() stdtypes.ModuleID {
	return m.SrcModule
}

func (m *MessageReceived) NewSrc(newSrc stdtypes.ModuleID) stdtypes.Event {
	newEvent := *m
	newEvent.SrcModule = newSrc
	return &newEvent
}

func (m *MessageReceived) Dest() stdtypes.ModuleID {
	return m.DstModule
}

func (m *MessageReceived) NewDest(newDest stdtypes.ModuleID) stdtypes.Event {
	newEvent := *m
	newEvent.DstModule = newDest
	return &newEvent
}

func (m *MessageReceived) SrcNode() stdtypes.NodeID {
	return m.SourceNode
}

func (m *MessageReceived) DestModule() stdtypes.ModuleID {
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
