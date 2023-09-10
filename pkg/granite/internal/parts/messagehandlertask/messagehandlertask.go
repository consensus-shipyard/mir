package messagehandlertask

import (
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/granite"
	"github.com/filecoin-project/mir/pkg/granite/common"
	"github.com/filecoin-project/mir/pkg/granite/internal/parts/consensustask"
	"github.com/filecoin-project/mir/pkg/logging"
	granitepbdsl "github.com/filecoin-project/mir/pkg/pb/granitepb/dsl"
	granitepbtypes "github.com/filecoin-project/mir/pkg/pb/granitepb/types"
	"github.com/filecoin-project/mir/pkg/types"
)

// IncludeMessageHandlerTask registers event handlers for the buffering and validation of Granite messages before delivering them to the core logic.
func IncludeMessageHandlerTask(
	m dsl.Module,
	mc *common.ModuleConfig,
	params *common.ModuleParams,
	state *common.State,
	logger logging.Logger,
) {

	granitepbdsl.UponConsensusMsgReceived(m, func(from types.NodeID, msgType granite.MsgType, round granite.RoundNr, data []uint8, ticket *granitepbtypes.Ticket, signature []uint8) error {
		if _, ok := state.UnvalidatedMsgs.Get(msgType, round, from); ok {
			return nil // message already received or duplicate, discarding
		} else if _, ok := state.ValidatedMsgs.Get(msgType, round, from); ok {
			return nil // message already received or duplicate, discarding
		}

		//Verify signature
		toVerify := [][]byte{data, params.InstanceUID, round.Bytes()}
		if msgType == granite.CONVERGE {
			if ticket == nil {
				logger.Log(logging.LevelWarn, "Ticket is nil")
				return nil
			}

			toVerify = append(toVerify, ticket.Data)
		}

		if err := params.Crypto.Verify(toVerify, signature, from); err != nil {
			logger.Log(logging.LevelWarn, "Signature verification failed")
			return nil
		}

		//TODO add check for external validity -- in mir probably asking the availability module

		msg := &granitepbtypes.ConsensusMsg{
			Round:     round,
			Data:      data,
			Ticket:    ticket,
			Signature: signature,
		}

		state.UnvalidatedMsgs.StoreMessage(msgType, round, from, msg)

		for msg, source, ok := state.FindNewValid(); ok; msg, source, ok = state.FindNewValid() {
			state.UnvalidatedMsgs.RemoveMessage(msgType, round, from)
			state.ValidatedMsgs.StoreMessage(msgType, round, from, msg)

			consensustask.MsgValidated(m,
				state,
				params,
				mc,
				source,
				msg,
			)
		}

		return nil
	})

}
