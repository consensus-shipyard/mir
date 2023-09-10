package consensustask

import (
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/granite"
	"github.com/filecoin-project/mir/pkg/granite/common"
	"github.com/filecoin-project/mir/pkg/logging"
	eventpbdsl "github.com/filecoin-project/mir/pkg/pb/eventpb/dsl"
	eventpbtypes "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	granitepbdsl "github.com/filecoin-project/mir/pkg/pb/granitepb/dsl"
	granitepbevents "github.com/filecoin-project/mir/pkg/pb/granitepb/events"
	granitepbmsgs "github.com/filecoin-project/mir/pkg/pb/granitepb/msgs"
	granitepbtypes "github.com/filecoin-project/mir/pkg/pb/granitepb/types"
	transportpbdsl "github.com/filecoin-project/mir/pkg/pb/transportpb/dsl"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	timertypes "github.com/filecoin-project/mir/pkg/timer/types"
	"github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
	"github.com/filecoin-project/mir/pkg/util/membutil"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"
	"math/rand"
	"time"
)

// IncludeConsensusTask registers event handlers for the core logic of Granite.
func IncludeConsensusTask(
	m dsl.Module,
	mc *common.ModuleConfig,
	params *common.ModuleParams,
	state *common.State,
	logger logging.Logger,
) {

	// TODO for now we immediately send a dummy propose UponInit, replace to whatever module provides the proposal
	eventpbdsl.UponInit(m, func() error {
		state.Proposal = []byte("test") // TODO get proposal
		state.Value = state.Proposal

		var signature []byte
		signature, err := params.Crypto.Sign([][]byte{state.Value, params.InstanceUID, state.Round.Bytes()})

		if err != nil {
			return err
		}

		//Send PROPOSE (first round Converge is skipped)
		transportpbdsl.SendMessage(m, mc.Net, granitepbmsgs.ConsensusMsg(mc.Self, granite.PROPOSE, state.Round, state.Proposal, nil, signature), maputil.GetKeys(params.Membership.Nodes))
		return nil
	})

	granitepbdsl.UponConvergeTimeout(m, func() error {
		checkConverge(
			m,
			state,
			params,
			mc,
		)

		return nil
	})
}

// TODO Actually implement this function with the core logic of Granite
func MsgValidated(m dsl.Module,
	state *common.State,
	params *common.ModuleParams,
	mc *common.ModuleConfig,
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
			params,
			mc,
			msg.Round,
		)
	case granite.PREPARE:
		PrepareValidated(
			m,
			state,
			params,
			mc,
			msg.Round,
		)
	case granite.COMMIT:
		CommitValidated(
			m,
			state,
			params,
			mc,
			msg.Round,
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
	// Do nothing, we move only once the timer has expired
	//TODO optimization: move if we have received messages from ALL nodes even if timer has not timed out yet

	return nil
}

func checkConverge(
	m dsl.Module,
	state *common.State,
	params *common.ModuleParams,
	mc *common.ModuleConfig,
) error {
	state.Value = getValueMinimumTicket(params.Membership, state.ValidatedMsgs.Msgs[granite.CONVERGE][state.Round])
	transportpbdsl.SendMessage(m, mc.Net, granitepbmsgs.ConsensusMsg(mc.Self, granite.PROPOSE, state.Round, state.Proposal, nil, signature), maputil.GetKeys(params.Membership.Nodes))

	state.Step = granite.PROPOSE

	//It could be that we have already received enough PREPARE messages to move to the next step
	checkPropose(
		m,
		state,
		params,
		mc,
	)
}

func ProposeValidated(m dsl.Module,
	state *common.State,
	params *common.ModuleParams,
	mc *common.ModuleConfig,
	round granite.RoundNr,
) error {
	// TODO check if we have enough messages to move to next step
	if round != state.Round {
		return nil
	}
	if state.Step != granite.PROPOSE {
		return nil
	}

	return checkPropose(m, state, params, mc)
}

func checkPropose(
	m dsl.Module,
	state *common.State,
	params *common.ModuleParams,
	mc *common.ModuleConfig,
) error {
	nodeIDs := maputil.GetKeys(state.ValidatedMsgs.Msgs[granite.PROPOSE][state.Round])

	// Check for strong quorum
	if !membutil.HaveStrongQuorum(params.Membership, nodeIDs) {
		return nil
	}

	mode := findMode(
		sliceutil.Transform(
			maputil.GetValues(state.ValidatedMsgs.Msgs[granite.PROPOSE][state.Round]), func(msg *granitepbtypes.ConsensusMsg) string {
				return string(msg.Data)
			}),
	)
	state.Value = []byte(mode)

	var signature []byte
	signature, err := params.Crypto.Sign([][]byte{state.Value, params.InstanceUID, state.Round.Bytes()})
	if err != nil {
		return err
	}

	//Send PREPARE
	transportpbdsl.SendMessage(m, mc.Net, granitepbmsgs.ConsensusMsg(mc.Self, granite.PREPARE, state.Round, state.Proposal, nil, signature), maputil.GetKeys(params.Membership.Nodes))

	state.Step = granite.PREPARE

	//It could be that we have already received enough PREPARE messages to move to the next step
	checkPrepare(
		m,
		state,
		params,
		mc,
	)
	return nil
}

func PrepareValidated(m dsl.Module,
	state *common.State,
	params *common.ModuleParams,
	mc *common.ModuleConfig,
	round granite.RoundNr,
) error {
	if round != state.Round {
		return nil
	}
	if state.Step != granite.PROPOSE {
		return nil
	}

	checkPrepare(
		m,
		state,
		params,
		mc,
	)
	return nil
}

func checkPrepare(
	m dsl.Module,
	state *common.State,
	params *common.ModuleParams,
	mc *common.ModuleConfig,
) error {
	nodeIDs := maputil.GetKeys(state.ValidatedMsgs.Msgs[granite.PROPOSE][state.Round])

	// Check for strong quorum
	if !membutil.HaveStrongQuorum(params.Membership, nodeIDs) {
		return nil
	}

	updatedValue := getValueWithQuorum(params.Membership, state.ValidatedMsgs.Msgs[granite.PROPOSE][state.Round])
	state.Value = nil
	if updatedValue != nil {
		state.Value = updatedValue
	}

	var signature []byte
	signature, err := params.Crypto.Sign([][]byte{state.Value, params.InstanceUID, state.Round.Bytes()})
	if err != nil {
		return err
	}

	//Send COMMIT
	transportpbdsl.SendMessage(m, mc.Net, granitepbmsgs.ConsensusMsg(mc.Self, granite.COMMIT, state.Round, state.Proposal, nil, signature), maputil.GetKeys(params.Membership.Nodes))

	state.Step = granite.COMMIT

	checkCommit(
		m,
		state,
		params,
		mc,
	)

	return nil
}

func CommitValidated(m dsl.Module,
	state *common.State,
	params *common.ModuleParams,
	mc *common.ModuleConfig,
	round granite.RoundNr,
) error {
	if round != state.Round {
		return nil
	}
	if state.Step != granite.COMMIT {
		return nil
	}

	checkCommit(
		m,
		state,
		params,
		mc,
	)
	return nil
}

func checkCommit(
	m dsl.Module,
	state *common.State,
	params *common.ModuleParams,
	mc *common.ModuleConfig,
) error {
	nodeIDs := maputil.GetKeys(state.ValidatedMsgs.Msgs[granite.PREPARE][state.Round])

	// Check for strong quorum
	if !membutil.HaveStrongQuorum(params.Membership, nodeIDs) {
		return nil
	}

	state.Value = getValueWithQuorum(params.Membership, state.ValidatedMsgs.Msgs[granite.PROPOSE][state.Round])
	if state.Value != nil {
		//FINISHED
		decide(...)
		return nil
	}

	state.Value = getValue(maputil.GetValues(state.ValidatedMsgs.Msgs[granite.PREPARE][state.Round]))
	state.Round++
	state.Step = granite.CONVERGE

	var signature []byte
	signature, err := params.Crypto.Sign([][]byte{state.Value, params.InstanceUID, state.Round.Bytes()})
	if err != nil {
		return err
	}

	//Send CONVERGE if not finished
	transportpbdsl.SendMessage(m, mc.Net, granitepbmsgs.ConsensusMsg(mc.Self, granite.COMMIT, state.Round, state.Proposal, nil, signature), maputil.GetKeys(params.Membership.Nodes))
	// Set up a new timer for the next proposal.
	timeoutEvent := granitepbevents.ConvergeTimeout(
		mc.Self)

	eventpbdsl.TimerDelay(m,
		mc.Timer,
		[]*eventpbtypes.Event{timeoutEvent},
		timertypes.Duration(params.ConvergeDelay),
	)
	return nil
}

// ------------------ Helper functions ------------------

func findMode[T comparable](dataList []T) T {
	dataCountMap := make(map[T]int)

	// Iterate through each data and count occurrences
	for _, data := range dataList {
		dataCountMap[data]++
	}

	var modeData T
	maxCount := 0

	// Find the data that occurs most frequently
	for data, count := range dataCountMap {
		if count > maxCount {
			maxCount = count
			modeData = data
		}
	}

	return modeData
}

// getValueWithQuorum returns a value if it gathers at least a strong quorum.
// If no value has a strong quorum, it returns the default value.
func getValueWithQuorum(membership *trantorpbtypes.Membership, messages map[types.NodeID]*granitepbtypes.ConsensusMsg) []byte {

	// Store the occurrence counts of each string value and the nodes that vouched for it
	valueNodes := make(map[string][]types.NodeID)

	// Populate valueCounts and valueNodes
	for source, msg := range messages {
		data := string(msg.Data) // Assuming the field is named Data and is of type string
		if _, exists := valueNodes[data]; !exists {
			valueNodes[data] = []types.NodeID{}
		}
		valueNodes[data] = append(valueNodes[data], source) // Assuming each message has a NodeID field
	}

	// Check each string value to see if it meets the strong quorum requirement
	for data, _ := range valueNodes {
		if membutil.HaveStrongQuorum(membership, valueNodes[data]) {
			return []byte(data)
		}
	}

	// If no value met the requirement, return the default value
	return nil
}

// getValue returns one of the non-empty values at random.
func getValue(messages []*granitepbtypes.ConsensusMsg) []byte {
	nonEmptyValues := [][]byte{}

	// Populate nonEmptyValues with non-empty string values from messages
	for _, msg := range messages {
		data := msg.Data // Assuming the field is named Data and is of type string
		if data != nil {
			nonEmptyValues = append(nonEmptyValues, data)
		}
	}

	// If there are no non-empty values, return an empty string
	if len(nonEmptyValues) == 0 {
		return nil
	}

	// Seed the random number generator and select a value randomly
	rand.Seed(time.Now().UnixNano())
	return nonEmptyValues[rand.Intn(len(nonEmptyValues))]
}

// getValueMinimumTicket returns the value with the lowest ticket.
func getValueMinimumTicket(membership *trantorpbtypes.Membership, messages map[types.NodeID]*granitepbtypes.ConsensusMsg) []byte {
	//TODO implement (return nil if empty, although at least local value should be part of set)
	return nil
}