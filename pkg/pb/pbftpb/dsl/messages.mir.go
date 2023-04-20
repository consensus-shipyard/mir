package pbftpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	types4 "github.com/filecoin-project/mir/pkg/orderers/types"
	dsl1 "github.com/filecoin-project/mir/pkg/pb/ordererpb/dsl"
	types2 "github.com/filecoin-project/mir/pkg/pb/ordererpb/types"
	types "github.com/filecoin-project/mir/pkg/pb/pbftpb/types"
	types3 "github.com/filecoin-project/mir/pkg/trantor/types"
	types1 "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for processing net messages.

func UponMessageReceived[W types.Message_TypeWrapper[M], M any](m dsl.Module, handler func(from types1.NodeID, msg *M) error) {
	dsl1.UponMessageReceived[*types2.Message_Pbft](m, func(from types1.NodeID, msg *types.Message) error {
		w, ok := msg.Type.(W)
		if !ok {
			return nil
		}

		return handler(from, w.Unwrap())
	})
}

func UponPreprepareReceived(m dsl.Module, handler func(from types1.NodeID, sn types3.SeqNr, view types4.ViewNr, data []uint8, aborted bool) error) {
	UponMessageReceived[*types.Message_Preprepare](m, func(from types1.NodeID, msg *types.Preprepare) error {
		return handler(from, msg.Sn, msg.View, msg.Data, msg.Aborted)
	})
}

func UponPrepareReceived(m dsl.Module, handler func(from types1.NodeID, sn types3.SeqNr, view types4.ViewNr, digest []uint8) error) {
	UponMessageReceived[*types.Message_Prepare](m, func(from types1.NodeID, msg *types.Prepare) error {
		return handler(from, msg.Sn, msg.View, msg.Digest)
	})
}

func UponCommitReceived(m dsl.Module, handler func(from types1.NodeID, sn types3.SeqNr, view types4.ViewNr, digest []uint8) error) {
	UponMessageReceived[*types.Message_Commit](m, func(from types1.NodeID, msg *types.Commit) error {
		return handler(from, msg.Sn, msg.View, msg.Digest)
	})
}

func UponDoneReceived(m dsl.Module, handler func(from types1.NodeID, digests [][]uint8) error) {
	UponMessageReceived[*types.Message_Done](m, func(from types1.NodeID, msg *types.Done) error {
		return handler(from, msg.Digests)
	})
}

func UponCatchUpRequestReceived(m dsl.Module, handler func(from types1.NodeID, digest []uint8, sn types3.SeqNr) error) {
	UponMessageReceived[*types.Message_CatchUpRequest](m, func(from types1.NodeID, msg *types.CatchUpRequest) error {
		return handler(from, msg.Digest, msg.Sn)
	})
}

func UponCatchUpResponseReceived(m dsl.Module, handler func(from types1.NodeID, resp *types.Preprepare) error) {
	UponMessageReceived[*types.Message_CatchUpResponse](m, func(from types1.NodeID, msg *types.CatchUpResponse) error {
		return handler(from, msg.Resp)
	})
}

func UponSignedViewChangeReceived(m dsl.Module, handler func(from types1.NodeID, viewChange *types.ViewChange, signature []uint8) error) {
	UponMessageReceived[*types.Message_SignedViewChange](m, func(from types1.NodeID, msg *types.SignedViewChange) error {
		return handler(from, msg.ViewChange, msg.Signature)
	})
}

func UponPreprepareRequestReceived(m dsl.Module, handler func(from types1.NodeID, digest []uint8, sn types3.SeqNr) error) {
	UponMessageReceived[*types.Message_PreprepareRequest](m, func(from types1.NodeID, msg *types.PreprepareRequest) error {
		return handler(from, msg.Digest, msg.Sn)
	})
}

func UponMissingPreprepareReceived(m dsl.Module, handler func(from types1.NodeID, preprepare *types.Preprepare) error) {
	UponMessageReceived[*types.Message_MissingPreprepare](m, func(from types1.NodeID, msg *types.MissingPreprepare) error {
		return handler(from, msg.Preprepare)
	})
}

func UponNewViewReceived(m dsl.Module, handler func(from types1.NodeID, view types4.ViewNr, viewChangeSenders []string, signedViewChanges []*types.SignedViewChange, preprepareSeqNrs []types3.SeqNr, preprepares []*types.Preprepare) error) {
	UponMessageReceived[*types.Message_NewView](m, func(from types1.NodeID, msg *types.NewView) error {
		return handler(from, msg.View, msg.ViewChangeSenders, msg.SignedViewChanges, msg.PreprepareSeqNrs, msg.Preprepares)
	})
}
