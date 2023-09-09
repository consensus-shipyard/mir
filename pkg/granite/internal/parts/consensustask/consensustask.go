package consensustask

import (
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/granite"
	"github.com/filecoin-project/mir/pkg/granite/common"
	"github.com/filecoin-project/mir/pkg/logging"
	eventpbdsl "github.com/filecoin-project/mir/pkg/pb/eventpb/dsl"
	granitepbmsgs "github.com/filecoin-project/mir/pkg/pb/granitepb/msgs"
	granitepbtypes "github.com/filecoin-project/mir/pkg/pb/granitepb/types"
	transportpbdsl "github.com/filecoin-project/mir/pkg/pb/transportpb/dsl"
	"github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
)

// IncludeConsensusTask registers event handlers for the core logic of Granite.
func IncludeConsensusTask(
	m dsl.Module,
	mc common.ModuleConfig,
	params *common.ModuleParams,
	state *common.State,
	logger logging.Logger,
) {

	// TODO for now we immediately send a dummy propose UponInit, replace to whatever module provides the proposal
	eventpbdsl.UponInit(m, func() error {
		proposal := []byte("test") // TODO get proposal

		var signature []byte // TODO for now we assume a dummy signature
		signature, err := params.Crypto.Sign([][]byte{proposal, params.InstanceUID, state.Round.Bytes()})

		if err != nil {
			return err
		}

		//sendConverge
		transportpbdsl.SendMessage(m, mc.Net, granitepbmsgs.ConsensusMsg(mc.Self, granite.PROPOSE, state.Round, proposal, nil, signature), maputil.GetKeys(params.Membership.Nodes))
		return nil
	})
}

// TODO Actually implement this function with the core logic of Granite
func MsgValidated(m dsl.Module,
	state *common.State,
	from types.NodeID,
	msg *granitepbtypes.ConsensusMsg) {
	//TODO implement discarding messages for old rounds and buffering msgs for new rounds

	switch msg.MsgType {
	case granite.CONVERGE:
		ConvergeValidated(
			m,
			state,
			from,
			msg.Round,
			msg.Data,
			msg.Ticket,
			msg.Signature,
		)
	case granite.PROPOSE:
		ProposeValidated(
			m,
			state,
			from,
			msg.Round,
			msg.Data,
			msg.Signature,
		)
	case granite.PREPARE:
		PrepareValidated(
			m,
			state,
			from,
			msg.Round,
			msg.Data,
			msg.Signature,
		)
	case granite.COMMIT:
		CommitValidated(
			m,
			state,
			from,
			msg.Round,
			msg.Data,
			msg.Signature,
		)
	}
}

func ConvergeValidated(m dsl.Module,
	state *common.State,
	from types.NodeID,
	round granite.RoundNr,
	data []uint8,
	ticket *granitepbtypes.Ticket,
	signature []uint8,
) error {
	// TODO check signature, ticket, round
	// TODO other checks...

	// TODO Store for when the timer comes to check for lowes ticket

	return nil
}

func ProposeValidated(m dsl.Module,
	state *common.State,
	from types.NodeID,
	round granite.RoundNr,
	data []uint8,
	signature []uint8,
) error {
	// TODO check signature, ticket, round
	// TODO other checks...

	// TODO check if we have enough messages to move to next step

	return nil
}

func PrepareValidated(m dsl.Module,
	state *common.State,
	from types.NodeID,
	round granite.RoundNr,
	data []uint8,
	signature []uint8,
) error {
	// TODO check signature, ticket, round
	// TODO other checks...

	// TODO check if we have enough messages to move to next step

	return nil
}

func CommitValidated(m dsl.Module,
	state *common.State,
	from types.NodeID,
	round granite.RoundNr,
	data []uint8,
	signature []uint8,
) error {
	// TODO check signature, ticket, round
	// TODO other checks...

	// TODO check if we have enough messages to move to next step
	// TODO Implement await 2Delta after sending CONVERGE, probably by calling a function that calls the timer module
	return nil
}
