package threshcryptopbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	types "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	events "github.com/filecoin-project/mir/pkg/pb/threshcryptopb/events"
	types2 "github.com/filecoin-project/mir/pkg/pb/threshcryptopb/types"
	types1 "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for emitting events.

func SignShare(m dsl.Module, next []*types.Event, destModule types1.ModuleID, data [][]uint8, origin *types2.SignShareOrigin) {
	dsl.EmitMirEvent(m, events.SignShare(next, destModule, data, origin))
}

func SignShareResult(m dsl.Module, next []*types.Event, destModule types1.ModuleID, signatureShare []uint8, origin *types2.SignShareOrigin) {
	dsl.EmitMirEvent(m, events.SignShareResult(next, destModule, signatureShare, origin))
}

func VerifyShare(m dsl.Module, next []*types.Event, destModule types1.ModuleID, data [][]uint8, signatureShare []uint8, nodeId types1.NodeID, origin *types2.VerifyShareOrigin) {
	dsl.EmitMirEvent(m, events.VerifyShare(next, destModule, data, signatureShare, nodeId, origin))
}

func VerifyShareResult(m dsl.Module, next []*types.Event, destModule types1.ModuleID, ok bool, error string, origin *types2.VerifyShareOrigin) {
	dsl.EmitMirEvent(m, events.VerifyShareResult(next, destModule, ok, error, origin))
}

func VerifyFull(m dsl.Module, next []*types.Event, destModule types1.ModuleID, data [][]uint8, fullSignature []uint8, origin *types2.VerifyFullOrigin) {
	dsl.EmitMirEvent(m, events.VerifyFull(next, destModule, data, fullSignature, origin))
}

func VerifyFullResult(m dsl.Module, next []*types.Event, destModule types1.ModuleID, ok bool, error string, origin *types2.VerifyFullOrigin) {
	dsl.EmitMirEvent(m, events.VerifyFullResult(next, destModule, ok, error, origin))
}

func Recover(m dsl.Module, next []*types.Event, destModule types1.ModuleID, data [][]uint8, signatureShares [][]uint8, origin *types2.RecoverOrigin) {
	dsl.EmitMirEvent(m, events.Recover(next, destModule, data, signatureShares, origin))
}

func RecoverResult(m dsl.Module, next []*types.Event, destModule types1.ModuleID, fullSignature []uint8, ok bool, error string, origin *types2.RecoverOrigin) {
	dsl.EmitMirEvent(m, events.RecoverResult(next, destModule, fullSignature, ok, error, origin))
}
