/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package orderers

import (
	"github.com/filecoin-project/mir/pkg/dsl"
	issconfig "github.com/filecoin-project/mir/pkg/iss/config"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/messagebuffer"
	"github.com/filecoin-project/mir/pkg/modules"
	common2 "github.com/filecoin-project/mir/pkg/orderers/common"
	"github.com/filecoin-project/mir/pkg/orderers/internal/common"
	"github.com/filecoin-project/mir/pkg/orderers/internal/parts/catchup"
	"github.com/filecoin-project/mir/pkg/orderers/internal/parts/goodcase"
	viewchange2 "github.com/filecoin-project/mir/pkg/orderers/internal/parts/viewchange"
	ot "github.com/filecoin-project/mir/pkg/orderers/types"
	factorypbtypes "github.com/filecoin-project/mir/pkg/pb/factorypb/types"
	ordererpbtypes "github.com/filecoin-project/mir/pkg/pb/ordererpb/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

// ============================================================
// PBFT State type and constructor
// ============================================================

// It takes the following parameters:
//   - moduleConfig
//   - ownID: The ID of this node.
//   - segment: The segment governing this SB instance,
//     specifying the leader, the set of sequence numbers, etc.
//   - config: PBFT-specific configuration parameters.
//   - externalValidator: Contains the application-specific code for validating incoming proposals.
//   - logger: Logger for outputting debugging messages.

func NewOrdererModule(
	moduleConfig common2.ModuleConfig,
	ownID t.NodeID,
	segment *common2.Segment,
	config *common.PBFTConfig,
	logger logging.Logger) modules.PassiveModule {

	// Set all the necessary fields of the new instance and return it.
	state := &common.State{
		Segment:           segment,
		Slots:             make(map[ot.ViewNr]map[tt.SeqNr]*common.PbftSlot),
		SegmentCheckpoint: common.NewPbftSegmentChkp(),
		Proposal: common.PbftProposalState{
			ProposalsMade:     0,
			CertRequested:     false,
			CertRequestedView: 0,
			ProposalTimeout:   0,
		},
		MessageBuffers: messagebuffer.NewBuffers(
			removeNodeID(config.Membership, ownID), // Create a message buffer for everyone except for myself.
			config.MsgBufCapacity,
			logging.Decorate(logger, "MsgBufs: "),
		),
		View:             0,
		InViewChange:     false,
		ViewChangeStates: make(map[ot.ViewNr]*common.PbftViewChangeState),
	}

	params := &common.ModuleParams{
		OwnID:  ownID,
		Config: config,
	}
	m := dsl.NewModule(moduleConfig.Self)

	goodcase.IncludeGoodCase(m, state, params, moduleConfig, logger)

	viewchange2.IncludeViewChange(m, state, params, moduleConfig, logger)

	catchup.IncludeSegmentCheckpoint(m, state, params, moduleConfig, logger)

	return m

}

func newOrdererConfig(issParams *issconfig.ModuleParams, membership []t.NodeID, epochNr tt.EpochNr) *common.PBFTConfig {

	// Return a new PBFT configuration with selected values from the ISS configuration.
	return &common.PBFTConfig{
		Membership:               membership,
		MaxProposeDelay:          issParams.MaxProposeDelay,
		MsgBufCapacity:           issParams.MsgBufCapacity,
		DoneResendPeriod:         issParams.PBFTDoneResendPeriod,
		CatchUpDelay:             issParams.PBFTCatchUpDelay,
		ViewChangeSNTimeout:      issParams.PBFTViewChangeSNTimeout,
		ViewChangeSegmentTimeout: issParams.PBFTViewChangeSegmentTimeout,
		ViewChangeResendPeriod:   issParams.PBFTViewChangeResendPeriod,
		EpochNr:                  epochNr,
	}
}

// removeNodeID removes a node ID from a list of node IDs.
// Takes a membership list and a Node ID and returns a new list of nodeIDs containing all IDs from the membership list,
// except for (if present) the specified nID.
// This is useful for obtaining the list of "other nodes" by removing the own ID from the membership.
func removeNodeID(membership []t.NodeID, nID t.NodeID) []t.NodeID {

	// Allocate the new node list.
	others := make([]t.NodeID, 0, len(membership))

	// Add all membership IDs except for the specified one.
	for _, nodeID := range membership {
		if nodeID != nID {
			others = append(others, nodeID)
		}
	}

	// Return the new list.
	return others
}

func InstanceParams(
	segment *common2.Segment,
	availabilityID t.ModuleID,
	epoch tt.EpochNr,
	PPVId t.ModuleID,
) *factorypbtypes.GeneratorParams {
	return &factorypbtypes.GeneratorParams{Type: &factorypbtypes.GeneratorParams_PbftModule{
		PbftModule: &ordererpbtypes.PBFTModule{
			Segment:        segment.PbType(),
			AvailabilityId: availabilityID.Pb(),
			Epoch:          epoch.Pb(),
			PpvModuleId:    string(PPVId),
		},
	}}
}
