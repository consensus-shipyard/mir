package eventpbtypes

import (
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"

	mirreflect "github.com/filecoin-project/mir/codegen/mirreflect"
	types15 "github.com/filecoin-project/mir/codegen/model/types"
	types13 "github.com/filecoin-project/mir/pkg/pb/apppb/types"
	types5 "github.com/filecoin-project/mir/pkg/pb/availabilitypb/batchdbpb/types"
	types4 "github.com/filecoin-project/mir/pkg/pb/availabilitypb/types"
	types6 "github.com/filecoin-project/mir/pkg/pb/batchfetcherpb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/bcbpb/types"
	types8 "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	types12 "github.com/filecoin-project/mir/pkg/pb/cryptopb/types"
	eventpb "github.com/filecoin-project/mir/pkg/pb/eventpb"
	types9 "github.com/filecoin-project/mir/pkg/pb/factorypb/types"
	types1 "github.com/filecoin-project/mir/pkg/pb/hasherpb/types"
	types10 "github.com/filecoin-project/mir/pkg/pb/isspb/types"
	types3 "github.com/filecoin-project/mir/pkg/pb/mempoolpb/types"
	types11 "github.com/filecoin-project/mir/pkg/pb/ordererpb/types"
	pingpongpb "github.com/filecoin-project/mir/pkg/pb/pingpongpb"
	types7 "github.com/filecoin-project/mir/pkg/pb/threshcryptopb/types"
	types14 "github.com/filecoin-project/mir/pkg/pb/transportpb/types"
	types16 "github.com/filecoin-project/mir/pkg/timer/types"
	types17 "github.com/filecoin-project/mir/pkg/trantor/types"
	types "github.com/filecoin-project/mir/pkg/types"
	reflectutil "github.com/filecoin-project/mir/pkg/util/reflectutil"
)

type Event struct {
	DestModule types.ModuleID
	Type       Event_Type
	Next       []*Event
}

type Event_Type interface {
	mirreflect.GeneratedType
	isEvent_Type()
	Pb() eventpb.Event_Type
}

type Event_TypeWrapper[T any] interface {
	Event_Type
	Unwrap() *T
}

func Event_TypeFromPb(pb eventpb.Event_Type) Event_Type {
	switch pb := pb.(type) {
	case *eventpb.Event_Init:
		return &Event_Init{Init: InitFromPb(pb.Init)}
	case *eventpb.Event_Timer:
		return &Event_Timer{Timer: TimerEventFromPb(pb.Timer)}
	case *eventpb.Event_Hasher:
		return &Event_Hasher{Hasher: types1.EventFromPb(pb.Hasher)}
	case *eventpb.Event_Bcb:
		return &Event_Bcb{Bcb: types2.EventFromPb(pb.Bcb)}
	case *eventpb.Event_Mempool:
		return &Event_Mempool{Mempool: types3.EventFromPb(pb.Mempool)}
	case *eventpb.Event_Availability:
		return &Event_Availability{Availability: types4.EventFromPb(pb.Availability)}
	case *eventpb.Event_BatchDb:
		return &Event_BatchDb{BatchDb: types5.EventFromPb(pb.BatchDb)}
	case *eventpb.Event_BatchFetcher:
		return &Event_BatchFetcher{BatchFetcher: types6.EventFromPb(pb.BatchFetcher)}
	case *eventpb.Event_ThreshCrypto:
		return &Event_ThreshCrypto{ThreshCrypto: types7.EventFromPb(pb.ThreshCrypto)}
	case *eventpb.Event_Checkpoint:
		return &Event_Checkpoint{Checkpoint: types8.EventFromPb(pb.Checkpoint)}
	case *eventpb.Event_Factory:
		return &Event_Factory{Factory: types9.EventFromPb(pb.Factory)}
	case *eventpb.Event_Iss:
		return &Event_Iss{Iss: types10.EventFromPb(pb.Iss)}
	case *eventpb.Event_Orderer:
		return &Event_Orderer{Orderer: types11.EventFromPb(pb.Orderer)}
	case *eventpb.Event_Crypto:
		return &Event_Crypto{Crypto: types12.EventFromPb(pb.Crypto)}
	case *eventpb.Event_App:
		return &Event_App{App: types13.EventFromPb(pb.App)}
	case *eventpb.Event_Transport:
		return &Event_Transport{Transport: types14.EventFromPb(pb.Transport)}
	case *eventpb.Event_PingPong:
		return &Event_PingPong{PingPong: pb.PingPong}
	case *eventpb.Event_TestingString:
		return &Event_TestingString{TestingString: pb.TestingString}
	case *eventpb.Event_TestingUint:
		return &Event_TestingUint{TestingUint: pb.TestingUint}
	}
	return nil
}

type Event_Init struct {
	Init *Init
}

func (*Event_Init) isEvent_Type() {}

func (w *Event_Init) Unwrap() *Init {
	return w.Init
}

func (w *Event_Init) Pb() eventpb.Event_Type {
	return &eventpb.Event_Init{Init: (w.Init).Pb()}
}

func (*Event_Init) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Init]()}
}

type Event_Timer struct {
	Timer *TimerEvent
}

func (*Event_Timer) isEvent_Type() {}

func (w *Event_Timer) Unwrap() *TimerEvent {
	return w.Timer
}

func (w *Event_Timer) Pb() eventpb.Event_Type {
	return &eventpb.Event_Timer{Timer: (w.Timer).Pb()}
}

func (*Event_Timer) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Timer]()}
}

type Event_Hasher struct {
	Hasher *types1.Event
}

func (*Event_Hasher) isEvent_Type() {}

func (w *Event_Hasher) Unwrap() *types1.Event {
	return w.Hasher
}

func (w *Event_Hasher) Pb() eventpb.Event_Type {
	return &eventpb.Event_Hasher{Hasher: (w.Hasher).Pb()}
}

func (*Event_Hasher) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Hasher]()}
}

type Event_Bcb struct {
	Bcb *types2.Event
}

func (*Event_Bcb) isEvent_Type() {}

func (w *Event_Bcb) Unwrap() *types2.Event {
	return w.Bcb
}

func (w *Event_Bcb) Pb() eventpb.Event_Type {
	return &eventpb.Event_Bcb{Bcb: (w.Bcb).Pb()}
}

func (*Event_Bcb) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Bcb]()}
}

type Event_Mempool struct {
	Mempool *types3.Event
}

func (*Event_Mempool) isEvent_Type() {}

func (w *Event_Mempool) Unwrap() *types3.Event {
	return w.Mempool
}

func (w *Event_Mempool) Pb() eventpb.Event_Type {
	return &eventpb.Event_Mempool{Mempool: (w.Mempool).Pb()}
}

func (*Event_Mempool) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Mempool]()}
}

type Event_Availability struct {
	Availability *types4.Event
}

func (*Event_Availability) isEvent_Type() {}

func (w *Event_Availability) Unwrap() *types4.Event {
	return w.Availability
}

func (w *Event_Availability) Pb() eventpb.Event_Type {
	return &eventpb.Event_Availability{Availability: (w.Availability).Pb()}
}

func (*Event_Availability) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Availability]()}
}

type Event_BatchDb struct {
	BatchDb *types5.Event
}

func (*Event_BatchDb) isEvent_Type() {}

func (w *Event_BatchDb) Unwrap() *types5.Event {
	return w.BatchDb
}

func (w *Event_BatchDb) Pb() eventpb.Event_Type {
	return &eventpb.Event_BatchDb{BatchDb: (w.BatchDb).Pb()}
}

func (*Event_BatchDb) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_BatchDb]()}
}

type Event_BatchFetcher struct {
	BatchFetcher *types6.Event
}

func (*Event_BatchFetcher) isEvent_Type() {}

func (w *Event_BatchFetcher) Unwrap() *types6.Event {
	return w.BatchFetcher
}

func (w *Event_BatchFetcher) Pb() eventpb.Event_Type {
	return &eventpb.Event_BatchFetcher{BatchFetcher: (w.BatchFetcher).Pb()}
}

func (*Event_BatchFetcher) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_BatchFetcher]()}
}

type Event_ThreshCrypto struct {
	ThreshCrypto *types7.Event
}

func (*Event_ThreshCrypto) isEvent_Type() {}

func (w *Event_ThreshCrypto) Unwrap() *types7.Event {
	return w.ThreshCrypto
}

func (w *Event_ThreshCrypto) Pb() eventpb.Event_Type {
	return &eventpb.Event_ThreshCrypto{ThreshCrypto: (w.ThreshCrypto).Pb()}
}

func (*Event_ThreshCrypto) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_ThreshCrypto]()}
}

type Event_Checkpoint struct {
	Checkpoint *types8.Event
}

func (*Event_Checkpoint) isEvent_Type() {}

func (w *Event_Checkpoint) Unwrap() *types8.Event {
	return w.Checkpoint
}

func (w *Event_Checkpoint) Pb() eventpb.Event_Type {
	return &eventpb.Event_Checkpoint{Checkpoint: (w.Checkpoint).Pb()}
}

func (*Event_Checkpoint) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Checkpoint]()}
}

type Event_Factory struct {
	Factory *types9.Event
}

func (*Event_Factory) isEvent_Type() {}

func (w *Event_Factory) Unwrap() *types9.Event {
	return w.Factory
}

func (w *Event_Factory) Pb() eventpb.Event_Type {
	return &eventpb.Event_Factory{Factory: (w.Factory).Pb()}
}

func (*Event_Factory) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Factory]()}
}

type Event_Iss struct {
	Iss *types10.Event
}

func (*Event_Iss) isEvent_Type() {}

func (w *Event_Iss) Unwrap() *types10.Event {
	return w.Iss
}

func (w *Event_Iss) Pb() eventpb.Event_Type {
	return &eventpb.Event_Iss{Iss: (w.Iss).Pb()}
}

func (*Event_Iss) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Iss]()}
}

type Event_Orderer struct {
	Orderer *types11.Event
}

func (*Event_Orderer) isEvent_Type() {}

func (w *Event_Orderer) Unwrap() *types11.Event {
	return w.Orderer
}

func (w *Event_Orderer) Pb() eventpb.Event_Type {
	return &eventpb.Event_Orderer{Orderer: (w.Orderer).Pb()}
}

func (*Event_Orderer) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Orderer]()}
}

type Event_Crypto struct {
	Crypto *types12.Event
}

func (*Event_Crypto) isEvent_Type() {}

func (w *Event_Crypto) Unwrap() *types12.Event {
	return w.Crypto
}

func (w *Event_Crypto) Pb() eventpb.Event_Type {
	return &eventpb.Event_Crypto{Crypto: (w.Crypto).Pb()}
}

func (*Event_Crypto) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Crypto]()}
}

type Event_App struct {
	App *types13.Event
}

func (*Event_App) isEvent_Type() {}

func (w *Event_App) Unwrap() *types13.Event {
	return w.App
}

func (w *Event_App) Pb() eventpb.Event_Type {
	return &eventpb.Event_App{App: (w.App).Pb()}
}

func (*Event_App) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_App]()}
}

type Event_Transport struct {
	Transport *types14.Event
}

func (*Event_Transport) isEvent_Type() {}

func (w *Event_Transport) Unwrap() *types14.Event {
	return w.Transport
}

func (w *Event_Transport) Pb() eventpb.Event_Type {
	return &eventpb.Event_Transport{Transport: (w.Transport).Pb()}
}

func (*Event_Transport) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Transport]()}
}

type Event_PingPong struct {
	PingPong *pingpongpb.Event
}

func (*Event_PingPong) isEvent_Type() {}

func (w *Event_PingPong) Unwrap() *pingpongpb.Event {
	return w.PingPong
}

func (w *Event_PingPong) Pb() eventpb.Event_Type {
	return &eventpb.Event_PingPong{PingPong: w.PingPong}
}

func (*Event_PingPong) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_PingPong]()}
}

type Event_TestingString struct {
	TestingString *wrapperspb.StringValue
}

func (*Event_TestingString) isEvent_Type() {}

func (w *Event_TestingString) Unwrap() *wrapperspb.StringValue {
	return w.TestingString
}

func (w *Event_TestingString) Pb() eventpb.Event_Type {
	return &eventpb.Event_TestingString{TestingString: w.TestingString}
}

func (*Event_TestingString) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_TestingString]()}
}

type Event_TestingUint struct {
	TestingUint *wrapperspb.UInt64Value
}

func (*Event_TestingUint) isEvent_Type() {}

func (w *Event_TestingUint) Unwrap() *wrapperspb.UInt64Value {
	return w.TestingUint
}

func (w *Event_TestingUint) Pb() eventpb.Event_Type {
	return &eventpb.Event_TestingUint{TestingUint: w.TestingUint}
}

func (*Event_TestingUint) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_TestingUint]()}
}

func EventFromPb(pb *eventpb.Event) *Event {
	return &Event{
		DestModule: (types.ModuleID)(pb.DestModule),
		Type:       Event_TypeFromPb(pb.Type),
		Next: types15.ConvertSlice(pb.Next, func(t *eventpb.Event) *Event {
			return EventFromPb(t)
		}),
	}
}

func (m *Event) Pb() *eventpb.Event {
	return &eventpb.Event{
		DestModule: (string)(m.DestModule),
		Type:       (m.Type).Pb(),
		Next: types15.ConvertSlice(m.Next, func(t *Event) *eventpb.Event {
			return (t).Pb()
		}),
	}
}

func (*Event) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event]()}
}

type Init struct{}

func InitFromPb(pb *eventpb.Init) *Init {
	return &Init{}
}

func (m *Init) Pb() *eventpb.Init {
	return &eventpb.Init{}
}

func (*Init) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Init]()}
}

type TimerEvent struct {
	Type TimerEvent_Type
}

type TimerEvent_Type interface {
	mirreflect.GeneratedType
	isTimerEvent_Type()
	Pb() eventpb.TimerEvent_Type
}

type TimerEvent_TypeWrapper[T any] interface {
	TimerEvent_Type
	Unwrap() *T
}

func TimerEvent_TypeFromPb(pb eventpb.TimerEvent_Type) TimerEvent_Type {
	switch pb := pb.(type) {
	case *eventpb.TimerEvent_Delay:
		return &TimerEvent_Delay{Delay: TimerDelayFromPb(pb.Delay)}
	case *eventpb.TimerEvent_Repeat:
		return &TimerEvent_Repeat{Repeat: TimerRepeatFromPb(pb.Repeat)}
	case *eventpb.TimerEvent_GarbageCollect:
		return &TimerEvent_GarbageCollect{GarbageCollect: TimerGarbageCollectFromPb(pb.GarbageCollect)}
	}
	return nil
}

type TimerEvent_Delay struct {
	Delay *TimerDelay
}

func (*TimerEvent_Delay) isTimerEvent_Type() {}

func (w *TimerEvent_Delay) Unwrap() *TimerDelay {
	return w.Delay
}

func (w *TimerEvent_Delay) Pb() eventpb.TimerEvent_Type {
	return &eventpb.TimerEvent_Delay{Delay: (w.Delay).Pb()}
}

func (*TimerEvent_Delay) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.TimerEvent_Delay]()}
}

type TimerEvent_Repeat struct {
	Repeat *TimerRepeat
}

func (*TimerEvent_Repeat) isTimerEvent_Type() {}

func (w *TimerEvent_Repeat) Unwrap() *TimerRepeat {
	return w.Repeat
}

func (w *TimerEvent_Repeat) Pb() eventpb.TimerEvent_Type {
	return &eventpb.TimerEvent_Repeat{Repeat: (w.Repeat).Pb()}
}

func (*TimerEvent_Repeat) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.TimerEvent_Repeat]()}
}

type TimerEvent_GarbageCollect struct {
	GarbageCollect *TimerGarbageCollect
}

func (*TimerEvent_GarbageCollect) isTimerEvent_Type() {}

func (w *TimerEvent_GarbageCollect) Unwrap() *TimerGarbageCollect {
	return w.GarbageCollect
}

func (w *TimerEvent_GarbageCollect) Pb() eventpb.TimerEvent_Type {
	return &eventpb.TimerEvent_GarbageCollect{GarbageCollect: (w.GarbageCollect).Pb()}
}

func (*TimerEvent_GarbageCollect) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.TimerEvent_GarbageCollect]()}
}

func TimerEventFromPb(pb *eventpb.TimerEvent) *TimerEvent {
	return &TimerEvent{
		Type: TimerEvent_TypeFromPb(pb.Type),
	}
}

func (m *TimerEvent) Pb() *eventpb.TimerEvent {
	return &eventpb.TimerEvent{
		Type: (m.Type).Pb(),
	}
}

func (*TimerEvent) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.TimerEvent]()}
}

type TimerDelay struct {
	EventsToDelay []*Event
	Delay         types16.Duration
}

func TimerDelayFromPb(pb *eventpb.TimerDelay) *TimerDelay {
	return &TimerDelay{
		EventsToDelay: types15.ConvertSlice(pb.EventsToDelay, func(t *eventpb.Event) *Event {
			return EventFromPb(t)
		}),
		Delay: (types16.Duration)(pb.Delay),
	}
}

func (m *TimerDelay) Pb() *eventpb.TimerDelay {
	return &eventpb.TimerDelay{
		EventsToDelay: types15.ConvertSlice(m.EventsToDelay, func(t *Event) *eventpb.Event {
			return (t).Pb()
		}),
		Delay: (uint64)(m.Delay),
	}
}

func (*TimerDelay) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.TimerDelay]()}
}

type TimerRepeat struct {
	EventsToRepeat []*Event
	Delay          types16.Duration
	RetentionIndex types17.RetentionIndex
}

func TimerRepeatFromPb(pb *eventpb.TimerRepeat) *TimerRepeat {
	return &TimerRepeat{
		EventsToRepeat: types15.ConvertSlice(pb.EventsToRepeat, func(t *eventpb.Event) *Event {
			return EventFromPb(t)
		}),
		Delay:          (types16.Duration)(pb.Delay),
		RetentionIndex: (types17.RetentionIndex)(pb.RetentionIndex),
	}
}

func (m *TimerRepeat) Pb() *eventpb.TimerRepeat {
	return &eventpb.TimerRepeat{
		EventsToRepeat: types15.ConvertSlice(m.EventsToRepeat, func(t *Event) *eventpb.Event {
			return (t).Pb()
		}),
		Delay:          (uint64)(m.Delay),
		RetentionIndex: (uint64)(m.RetentionIndex),
	}
}

func (*TimerRepeat) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.TimerRepeat]()}
}

type TimerGarbageCollect struct {
	RetentionIndex types17.RetentionIndex
}

func TimerGarbageCollectFromPb(pb *eventpb.TimerGarbageCollect) *TimerGarbageCollect {
	return &TimerGarbageCollect{
		RetentionIndex: (types17.RetentionIndex)(pb.RetentionIndex),
	}
}

func (m *TimerGarbageCollect) Pb() *eventpb.TimerGarbageCollect {
	return &eventpb.TimerGarbageCollect{
		RetentionIndex: (uint64)(m.RetentionIndex),
	}
}

func (*TimerGarbageCollect) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.TimerGarbageCollect]()}
}
