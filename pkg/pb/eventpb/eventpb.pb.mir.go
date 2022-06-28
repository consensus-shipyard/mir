package eventpb

import (
	bcbpb "github.com/filecoin-project/mir/pkg/pb/bcbpb"
	isspb "github.com/filecoin-project/mir/pkg/pb/isspb"
	mempoolpb "github.com/filecoin-project/mir/pkg/pb/mempoolpb"
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"
)

type Event_Type = isEvent_Type

type Event_TypeWrapper[Ev any] interface {
	Event_Type
	Unwrap() *Ev
}

func (p *Event_Init) Unwrap() *Init {
	return p.Init
}

func (p *Event_Tick) Unwrap() *Tick {
	return p.Tick
}

func (p *Event_WalAppend) Unwrap() *WALAppend {
	return p.WalAppend
}

func (p *Event_WalEntry) Unwrap() *WALEntry {
	return p.WalEntry
}

func (p *Event_WalTruncate) Unwrap() *WALTruncate {
	return p.WalTruncate
}

func (p *Event_WalLoadAll) Unwrap() *WALLoadAll {
	return p.WalLoadAll
}

func (p *Event_NewRequests) Unwrap() *NewRequests {
	return p.NewRequests
}

func (p *Event_HashRequest) Unwrap() *HashRequest {
	return p.HashRequest
}

func (p *Event_HashResult) Unwrap() *HashResult {
	return p.HashResult
}

func (p *Event_SignRequest) Unwrap() *SignRequest {
	return p.SignRequest
}

func (p *Event_SignResult) Unwrap() *SignResult {
	return p.SignResult
}

func (p *Event_VerifyNodeSigs) Unwrap() *VerifyNodeSigs {
	return p.VerifyNodeSigs
}

func (p *Event_NodeSigsVerified) Unwrap() *NodeSigsVerified {
	return p.NodeSigsVerified
}

func (p *Event_RequestReady) Unwrap() *RequestReady {
	return p.RequestReady
}

func (p *Event_SendMessage) Unwrap() *SendMessage {
	return p.SendMessage
}

func (p *Event_MessageReceived) Unwrap() *MessageReceived {
	return p.MessageReceived
}

func (p *Event_Deliver) Unwrap() *Deliver {
	return p.Deliver
}

func (p *Event_Iss) Unwrap() *isspb.ISSEvent {
	return p.Iss
}

func (p *Event_VerifyRequestSig) Unwrap() *VerifyRequestSig {
	return p.VerifyRequestSig
}

func (p *Event_RequestSigVerified) Unwrap() *RequestSigVerified {
	return p.RequestSigVerified
}

func (p *Event_StoreVerifiedRequest) Unwrap() *StoreVerifiedRequest {
	return p.StoreVerifiedRequest
}

func (p *Event_AppSnapshotRequest) Unwrap() *AppSnapshotRequest {
	return p.AppSnapshotRequest
}

func (p *Event_AppSnapshot) Unwrap() *AppSnapshot {
	return p.AppSnapshot
}

func (p *Event_AppRestoreState) Unwrap() *AppRestoreState {
	return p.AppRestoreState
}

func (p *Event_TimerDelay) Unwrap() *TimerDelay {
	return p.TimerDelay
}

func (p *Event_TimerRepeat) Unwrap() *TimerRepeat {
	return p.TimerRepeat
}

func (p *Event_TimerGarbageCollect) Unwrap() *TimerGarbageCollect {
	return p.TimerGarbageCollect
}

func (p *Event_Bcb) Unwrap() *bcbpb.Event {
	return p.Bcb
}

func (p *Event_Mempool) Unwrap() *mempoolpb.Event {
	return p.Mempool
}

func (p *Event_TestingString) Unwrap() *wrapperspb.StringValue {
	return p.TestingString
}

func (p *Event_TestingUint) Unwrap() *wrapperspb.UInt64Value {
	return p.TestingUint
}
