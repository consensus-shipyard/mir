package messagepbtypes

import (
	mirreflect "github.com/filecoin-project/mir/codegen/mirreflect"
	mscpb "github.com/filecoin-project/mir/pkg/pb/availabilitypb/mscpb"
	types1 "github.com/filecoin-project/mir/pkg/pb/bcbpb/types"
	checkpointpb "github.com/filecoin-project/mir/pkg/pb/checkpointpb"
	isspb "github.com/filecoin-project/mir/pkg/pb/isspb"
	messagepb "github.com/filecoin-project/mir/pkg/pb/messagepb"
	ordererspb "github.com/filecoin-project/mir/pkg/pb/ordererspb"
	pingpongpb "github.com/filecoin-project/mir/pkg/pb/pingpongpb"
	types "github.com/filecoin-project/mir/pkg/types"
	reflectutil "github.com/filecoin-project/mir/pkg/util/reflectutil"
)

type Message struct {
	DestModule types.ModuleID
	Type       Message_Type
}

type Message_Type interface {
	mirreflect.GeneratedType
	isMessage_Type()
	Pb() messagepb.Message_Type
}

type Message_TypeWrapper[T any] interface {
	Message_Type
	Unwrap() *T
}

func Message_TypeFromPb(pb messagepb.Message_Type) Message_Type {
	switch pb := pb.(type) {
	case *messagepb.Message_Iss:
		return &Message_Iss{Iss: pb.Iss}
	case *messagepb.Message_Bcb:
		return &Message_Bcb{Bcb: types1.MessageFromPb(pb.Bcb)}
	case *messagepb.Message_MultisigCollector:
		return &Message_MultisigCollector{MultisigCollector: pb.MultisigCollector}
	case *messagepb.Message_Pingpong:
		return &Message_Pingpong{Pingpong: pb.Pingpong}
	case *messagepb.Message_Checkpoint:
		return &Message_Checkpoint{Checkpoint: pb.Checkpoint}
	case *messagepb.Message_SbMessage:
		return &Message_SbMessage{SbMessage: pb.SbMessage}
	}
	return nil
}

type Message_Iss struct {
	Iss *isspb.ISSMessage
}

func (*Message_Iss) isMessage_Type() {}

func (w *Message_Iss) Unwrap() *isspb.ISSMessage {
	return w.Iss
}

func (w *Message_Iss) Pb() messagepb.Message_Type {
	return &messagepb.Message_Iss{Iss: w.Iss}
}

func (*Message_Iss) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*messagepb.Message_Iss]()}
}

type Message_Bcb struct {
	Bcb *types1.Message
}

func (*Message_Bcb) isMessage_Type() {}

func (w *Message_Bcb) Unwrap() *types1.Message {
	return w.Bcb
}

func (w *Message_Bcb) Pb() messagepb.Message_Type {
	return &messagepb.Message_Bcb{Bcb: (w.Bcb).Pb()}
}

func (*Message_Bcb) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*messagepb.Message_Bcb]()}
}

type Message_MultisigCollector struct {
	MultisigCollector *mscpb.Message
}

func (*Message_MultisigCollector) isMessage_Type() {}

func (w *Message_MultisigCollector) Unwrap() *mscpb.Message {
	return w.MultisigCollector
}

func (w *Message_MultisigCollector) Pb() messagepb.Message_Type {
	return &messagepb.Message_MultisigCollector{MultisigCollector: w.MultisigCollector}
}

func (*Message_MultisigCollector) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*messagepb.Message_MultisigCollector]()}
}

type Message_Pingpong struct {
	Pingpong *pingpongpb.Message
}

func (*Message_Pingpong) isMessage_Type() {}

func (w *Message_Pingpong) Unwrap() *pingpongpb.Message {
	return w.Pingpong
}

func (w *Message_Pingpong) Pb() messagepb.Message_Type {
	return &messagepb.Message_Pingpong{Pingpong: w.Pingpong}
}

func (*Message_Pingpong) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*messagepb.Message_Pingpong]()}
}

type Message_Checkpoint struct {
	Checkpoint *checkpointpb.Message
}

func (*Message_Checkpoint) isMessage_Type() {}

func (w *Message_Checkpoint) Unwrap() *checkpointpb.Message {
	return w.Checkpoint
}

func (w *Message_Checkpoint) Pb() messagepb.Message_Type {
	return &messagepb.Message_Checkpoint{Checkpoint: w.Checkpoint}
}

func (*Message_Checkpoint) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*messagepb.Message_Checkpoint]()}
}

type Message_SbMessage struct {
	SbMessage *ordererspb.SBInstanceMessage
}

func (*Message_SbMessage) isMessage_Type() {}

func (w *Message_SbMessage) Unwrap() *ordererspb.SBInstanceMessage {
	return w.SbMessage
}

func (w *Message_SbMessage) Pb() messagepb.Message_Type {
	return &messagepb.Message_SbMessage{SbMessage: w.SbMessage}
}

func (*Message_SbMessage) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*messagepb.Message_SbMessage]()}
}

func MessageFromPb(pb *messagepb.Message) *Message {
	return &Message{
		DestModule: (types.ModuleID)(pb.DestModule),
		Type:       Message_TypeFromPb(pb.Type),
	}
}

func (m *Message) Pb() *messagepb.Message {
	return &messagepb.Message{
		DestModule: (string)(m.DestModule),
		Type:       (m.Type).Pb(),
	}
}

func (*Message) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*messagepb.Message]()}
}
