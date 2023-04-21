/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package iss contains the implementation of the ISS protocol, the new generation of Mir.
// For the details of the protocol, see README.md in this directory.
// To use ISS, instantiate it by calling `iss.New` and use it as the Protocol module when instantiating a mir.Node.
// A default configuration (to pass, among other arguments, to `iss.New`)
// can be obtained from `issutil.DefaultParams`.
package iss

import (
	"encoding/binary"
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/filecoin-project/mir/pkg/checkpoint"
	chkpprotos "github.com/filecoin-project/mir/pkg/checkpoint/protobufs"
	"github.com/filecoin-project/mir/pkg/clientprogress"
	"github.com/filecoin-project/mir/pkg/crypto"
	"github.com/filecoin-project/mir/pkg/dsl"
	issconfig "github.com/filecoin-project/mir/pkg/iss/config"
	lsp "github.com/filecoin-project/mir/pkg/iss/leaderselectionpolicy"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/orderers"
	apppbdsl "github.com/filecoin-project/mir/pkg/pb/apppb/dsl"
	"github.com/filecoin-project/mir/pkg/pb/availabilitypb"
	apbdsl "github.com/filecoin-project/mir/pkg/pb/availabilitypb/dsl"
	mscpbtypes "github.com/filecoin-project/mir/pkg/pb/availabilitypb/mscpb/types"
	apbtypes "github.com/filecoin-project/mir/pkg/pb/availabilitypb/types"
	chkppbdsl "github.com/filecoin-project/mir/pkg/pb/checkpointpb/dsl"
	chkppbmsgs "github.com/filecoin-project/mir/pkg/pb/checkpointpb/msgs"
	chkppbtypes "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	commonpbtypes "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	eventpbdsl "github.com/filecoin-project/mir/pkg/pb/eventpb/dsl"
	eventpbtypes "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	factorypbdsl "github.com/filecoin-project/mir/pkg/pb/factorypb/dsl"
	factorypbtypes "github.com/filecoin-project/mir/pkg/pb/factorypb/types"
	isspbdsl "github.com/filecoin-project/mir/pkg/pb/isspb/dsl"
	isspbevents "github.com/filecoin-project/mir/pkg/pb/isspb/events"
	transportpbdsl "github.com/filecoin-project/mir/pkg/pb/transportpb/dsl"
	"github.com/filecoin-project/mir/pkg/timer/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"
)

// The ISS type represents the ISS protocol module to be used when instantiating a node.
// The type should not be instantiated directly, but only properly initialized values
// returned from the New() function should be used.
type ISS struct {

	// The ID of the node executing this instance of the protocol.
	ownID t.NodeID

	// IDs of modules ISS interacts with.
	moduleConfig *ModuleConfig

	// The ISS configuration parameters (e.g. Segment length, proposal frequency etc...)
	// passed to New() when creating an ISS protocol instance.
	Params *issconfig.ModuleParams

	// Implementation of the hash function to use by ISS for computing all hashes.
	hashImpl crypto.HashImpl

	// Verifier of received stable checkpoints. This is most likely going to be the crypto module used by the protocol.
	chkpVerifier checkpoint.Verifier

	// Logger the ISS implementation uses to output log messages.
	// This is mostly for debugging - not to be confused with the commit log.
	logger logging.Logger

	// The current epoch instance.
	epoch *epochInfo

	// Epoch instances.
	epochs map[tt.EpochNr]*epochInfo

	// Highest epoch numbers indicated in Checkpoint messages from each node.
	nodeEpochMap map[t.NodeID]tt.EpochNr

	// The memberships for the current epoch and the params.ConfigOffset following epochs
	// (totalling params.ConfigOffset memberships).
	// E.g., if params.ConfigOffset 3 and the current epoch is 5, this field contains memberships for epoch 5, 6, 7 and 8.
	memberships []*commonpbtypes.Membership

	// The latest new membership obtained from the application.
	// To be included as last of the list of membership in the next new configuration.
	// The epoch number this membership corresponds to is the current epoch number + params.ConfigOffset.
	// At epoch initialization, this is set to nil. We then set it to the next membership announced by the application.
	nextNewMembership *commonpbtypes.Membership

	// The final log of committed availability certificates.
	// For each sequence number, it holds the committed certificate (or the special abort value).
	// Each Deliver event of an orderer translates to inserting an entry in the commitLog.
	// This, in turn, leads to delivering the certificate to the application,
	// as soon as all entries with lower sequence numbers have been delivered.
	// I.e., the entries are not necessarily inserted in order of their sequence numbers,
	// but they are delivered to the application in that order.
	commitLog map[tt.SeqNr]*CommitLogEntry

	// The first undelivered sequence number in the commitLog.
	// This field drives the in-order delivery of the log entries to the application.
	nextDeliveredSN tt.SeqNr

	// The first sequence number to be delivered in the new epoch.
	newEpochSN tt.SeqNr

	// Stores the stable checkpoint with the highest sequence number agreed-upon so far.
	// If no stable checkpoint has been observed yet, lastStableCheckpoint is initialized to a stable checkpoint value
	// corresponding to the initial state and associated with sequence number 0.
	lastStableCheckpoint *checkpoint.StableCheckpoint

	// Stores the highest stable checkpoint sequence number received from the checkpointing protocol so far.
	// This value is set immediately on reception of the checkpoint, without waiting for agreement on it.
	// This is only necessary if the checkpointing protocol announces a stable checkpoint multiple times,
	// in which case we use this value to only start a single agreement protocol for the same checkpoint.
	lastPendingCheckpointSN tt.SeqNr

	// The logic for selecting leader nodes in each epoch.
	// For details see the documentation of the LeaderSelectionPolicy type.
	// ATTENTION: The leader selection policy is stateful!
	// Must not be nil.
	LeaderPolicy lsp.LeaderSelectionPolicy

	// The DSL Module.
	m dsl.Module
}

// New returns a new initialized instance of the ISS protocol module to be used when instantiating a mir.Node.
func New(

	// ID of the node being instantiated with ISS.
	ownID t.NodeID,

	// IDs of the modules ISS interacts with.
	moduleConfig *ModuleConfig,

	// ISS protocol-specific configuration (e.g. segment length, proposal frequency etc...).
	// See the documentation of the issutil.ModuleParams type for details.
	params *issconfig.ModuleParams,

	// Stable checkpoint defining the initial state of the protocol.
	startingChkp *checkpoint.StableCheckpoint,

	// Hash implementation to use when computing hashes.
	hashImpl crypto.HashImpl,

	// Verifier of received stable checkpoints.
	// This is most likely going to be the crypto module used by the protocol.
	chkpVerifier checkpoint.Verifier,

	// Logger the ISS implementation uses to output log messages.
	logger logging.Logger,

) (modules.PassiveModule, error) {
	if logger == nil {
		logger = logging.ConsoleErrorLogger
	}

	// Check whether the passed configuration is valid.
	if err := issconfig.CheckParams(params); err != nil {
		return nil, fmt.Errorf("invalid ISS configuration: %w", err)
	}

	// TODO: Make sure that startingChkp is consistent with params.
	leaderPolicy, err := lsp.LeaderPolicyFromBytes(startingChkp.Snapshot.EpochData.LeaderPolicy)
	if err != nil {
		return nil, fmt.Errorf("invalid leader policy in starting checkpoint: %w", err)
	}

	// Initialize a new ISS object.
	iss := &ISS{
		ownID:                   ownID,
		moduleConfig:            moduleConfig,
		Params:                  params,
		hashImpl:                hashImpl,
		chkpVerifier:            chkpVerifier,
		logger:                  logger,
		epoch:                   nil, // Initialized later.
		epochs:                  make(map[tt.EpochNr]*epochInfo),
		nodeEpochMap:            make(map[t.NodeID]tt.EpochNr),
		memberships:             startingChkp.Memberships(),
		nextNewMembership:       nil,
		commitLog:               make(map[tt.SeqNr]*CommitLogEntry),
		nextDeliveredSN:         startingChkp.SeqNr(),
		newEpochSN:              startingChkp.SeqNr(),
		lastStableCheckpoint:    startingChkp,
		lastPendingCheckpointSN: 0,
		LeaderPolicy:            leaderPolicy,
		// TODO: Make sure that verification of the stable checkpoint certificate for epoch 0 is handled properly.
		//       (Probably "always valid", if the membership is right.) There is no epoch -1 with nodes to sign it.
	}

	// ============================================================
	// Event application
	// ============================================================

	iss.m = dsl.NewModule(moduleConfig.Self)

	// Upon Init event, initialize the ISS protocol.
	// This event is only expected to be applied once at startup,
	// after all the events stored in the WAL have been applied and before any other event has been applied.
	// (At this time, the WAL is not used. TODO: Update this when wal is implemented.)
	eventpbdsl.UponInit(iss.m, func() error {

		// Initialize application state according to the initial checkpoint.
		apppbdsl.RestoreState(
			iss.m,
			iss.moduleConfig.App,
			chkppbtypes.StableCheckpointFromPb(iss.lastStableCheckpoint.Pb()),
		)

		// Start the first epoch (not necessarily epoch 0, depending on the starting checkpoint).
		iss.startEpoch(iss.lastStableCheckpoint.Epoch())

		return nil
	})

	// Upon SBDeliver event, process the event of an SB instance delivering data.
	// It invokes the appropriate handler depending on whether this data is
	// an availability certificate (or the special abort value) or a common checkpoint.
	isspbdsl.UponSBDeliver(iss.m, func(sn tt.SeqNr, data []uint8, aborted bool, leader t.NodeID, instanceId t.ModuleID) error {
		// If checkpoint ordering instance, deliver without verification
		if instanceId.Sub().Sub() == "chkp" {
			return iss.deliverCommonCheckpoint(data)
		}

		// If decided empty certificate, deliver already
		if len(data) == 0 {
			return iss.deliverCert(sn, data, aborted, leader)
		}

		// verify the certificate before deciding it
		return iss.verifyCert(sn, data, aborted, leader)
	})

	apbdsl.UponCertVerified(iss.m, func(valid bool, err string, context *verifyCertContext) error {
		if !valid {
			// decide empty cert
			context.data = []byte{}
			// suspect leader
			iss.LeaderPolicy.Suspect(iss.epoch.Nr(), context.leader)
		}
		return iss.deliverCert(context.sn, context.data, context.aborted, context.leader)
	})

	isspbdsl.UponNewConfig(iss.m, func(epochNr tt.EpochNr, membership *commonpbtypes.Membership) error {
		iss.logger.Log(logging.LevelDebug, "Received new configuration.",
			"numNodes", len(membership.Nodes), "epochNr", epochNr, "currentEpoch", iss.epoch.Nr())

		// Check whether this event is not outdated.
		// This can (and did) happen in a corner case where the state gets restored from a checkpoint
		// while a NewConfig event is already in the pipeline.
		// Note that config.EpochNr is NOT the epoch where this new configuration is used,
		// but it is the epoch in which this new configuration is supposed to be received
		// (it will be used iss.params.ConfigOffset epochs later).
		if epochNr < iss.epoch.Nr() {
			iss.logger.Log(logging.LevelWarn, "Ignoring outdated membership.",
				"notificationEpoch", epochNr, "currentEpoch", iss.epoch.Nr())
			return nil
		}
		// Sanity check.
		if iss.nextNewMembership != nil {
			return fmt.Errorf("already have a new membership for epoch %v", iss.epoch.Nr())
		}
		// Convenience variable.
		newMembershipEpoch := iss.epoch.Nr() + tt.EpochNr(len(iss.memberships))

		// Save the new configuration.
		iss.nextNewMembership = membership

		iss.logger.Log(logging.LevelDebug, "Adding configuration",
			"forEpoch", newMembershipEpoch,
			"currentEpoch", iss.epoch.Nr(),
			"newConfigNodes", maputil.GetSortedKeys(iss.nextNewMembership.Nodes))

		// Advance to the next epoch if this configuration was the last missing bit.
		if iss.epochFinished() {
			err := iss.advanceEpoch()
			if err != nil {
				return err
			}
		}
		return nil
	})

	// Upon EpochProgress handle the event informing ISS that a node reached a certain epoch.
	// This event can be, for example, emitted by the checkpoint protocol
	// when it detects a node having reached a checkpoint for a certain epoch.
	chkppbdsl.UponEpochProgress(iss.m, func(nodeID t.NodeID, epochNr tt.EpochNr) error {
		// Remember the highest epoch number for each node to detect
		// later if the remote node is delayed too much and requires
		// assistance in order to catch up through state transfer.
		if iss.nodeEpochMap[nodeID] < epochNr {
			iss.nodeEpochMap[nodeID] = epochNr
		}
		return nil
	})

	// applyStableCheckpoint handles a new stable checkpoint produced by the checkpoint protocol.
	// It serializes and submits the checkpoint for agreement.
	chkppbdsl.UponStableCheckpoint(iss.m, func(sn tt.SeqNr, snapshot *commonpbtypes.StateSnapshot, cert map[t.NodeID][]byte) error {
		// Ignore old checkpoints.
		if sn <= iss.lastPendingCheckpointSN {
			iss.logger.Log(logging.LevelDebug, "Ignoring outdated stable checkpoint.", "sn", sn)
			return nil
		}
		iss.lastPendingCheckpointSN = sn

		// Convenience variables
		membership := snapshot.EpochData.PreviousMembership
		epoch := snapshot.EpochData.EpochConfig.EpochNr

		// If this is the most recent checkpoint observed, save it.
		iss.logger.Log(logging.LevelInfo, "New stable checkpoint.",
			"currentEpoch", iss.epoch.Nr(),
			"chkpEpoch", epoch,
			"sn", sn,
			"replacingEpoch", iss.lastStableCheckpoint.Epoch(),
			"replacingSn", iss.lastStableCheckpoint.SeqNr(),
			"numMemberships", len(snapshot.EpochData.EpochConfig.Memberships),
		)

		// The remainder of this function creates a new orderer instance to agree on a common checkpoint.
		// This is only necessary because different nodes might have different checkpoint certificates
		// (as they consist of quorums of collected signatures, which might differ at different nodes).
		// Since we want all nodes to deliver the same checkpoint certificate to the application,
		// we first need to agree on one (any of them).

		// Choose a leader for the new orderer instance.
		// TODO: Use the corresponding epoch's leader set to pick a leader, instead of just selecting one from all nodes.
		leader := maputil.GetSortedKeys(membership.Nodes)[int(epoch)%len(membership.Nodes)]

		// Serialize checkpoint, so it can be proposed as a value.
		stableCheckpoint := chkppbtypes.StableCheckpoint{
			Sn:       sn,
			Snapshot: snapshot,
			Cert:     cert,
		}
		chkpData, err := checkpoint.StableCheckpointFromPb(stableCheckpoint.Pb()).Serialize()
		if err != nil {
			return fmt.Errorf("failed serializing stable checkpoint: %w", err)
		}

		// Instantiate a new PBFT orderer.
		factorypbdsl.NewModule(iss.m,
			iss.moduleConfig.Ordering,
			iss.moduleConfig.Ordering.Then(t.ModuleID(fmt.Sprintf("%v", epoch))).Then("chkp"),
			tt.RetentionIndex(epoch),
			orderers.InstanceParams(
				orderers.NewSegment(leader, membership, map[tt.SeqNr][]byte{0: chkpData}),
				"", // The checkpoint orderer should never talk to the availability module, as it has a set proposal.
				epoch,
				orderers.CheckpointValidityChecker,
			))

		return nil
	})

	isspbdsl.UponPushCheckpoint(iss.m, func() error {
		// Send the latest stable checkpoint to potentially
		// delayed nodes. The set of nodes to send the latest
		// stable checkpoint to is determined based on the
		// highest epoch number in the checkpoint messages received from
		// the node so far. If the highest epoch number we
		// have heard from the node is more than params.RetainedEpochs
		// behind, then that node is likely to be left behind
		// and needs the stable checkpoint in order to start
		// catching up with state transfer.

		// Convenience variables
		epoch := iss.lastStableCheckpoint.Epoch()
		membership := iss.lastStableCheckpoint.Memberships()[0] // Membership of the epoch started by th checkpoint

		// We loop through the checkpoint's first epoch's membership
		// and not through the membership this replica is currently in.
		// (Which might differ if the replica already advanced to a new epoch,
		// but hasn't obtained its starting checkpoint yet.)
		// This is required to avoid sending old checkpoints to replicas
		// that are not yet part of the system for those checkpoints.
		var delayed []t.NodeID
		for n := range membership.Nodes {
			if epoch > iss.nodeEpochMap[n]+tt.EpochNr(iss.Params.RetainedEpochs) {
				delayed = append(delayed, n)
			}
		}

		if len(delayed) > 0 {
			iss.logger.Log(logging.LevelDebug, "Pushing state to nodes.",
				"delayed", delayed, "numNodes", len(iss.epoch.Membership.Nodes), "nodeEpochMap", iss.nodeEpochMap)
		}

		transportpbdsl.SendMessage(iss.m,
			iss.moduleConfig.Net,
			chkppbmsgs.StableCheckpoint(iss.moduleConfig.Self, iss.lastStableCheckpoint.SeqNr(),
				iss.lastStableCheckpoint.StateSnapshot(),
				iss.lastStableCheckpoint.Certificate()),
			delayed)

		return nil
	})

	chkppbdsl.UponStableCheckpointReceived(iss.m, func(sender t.NodeID, sn tt.SeqNr, snapshot *commonpbtypes.StateSnapshot, cert map[t.NodeID][]byte) error {
		_chkp := chkppbtypes.StableCheckpoint{
			Sn:       sn,
			Snapshot: snapshot,
			Cert:     cert,
		}
		chkp := checkpoint.StableCheckpointFromPb(_chkp.Pb())

		// Check how far the received stable checkpoint is ahead of the local node's state.
		chkpMembershipOffset := int(chkp.Epoch()) - 1 - int(iss.epoch.Nr())
		if chkpMembershipOffset <= 0 {
			// Ignore stable checkpoints that are not far enough
			// ahead of the current state of the local node.
			return nil
		}

		// Ignore checkpoint if we are not part of its membership
		// (more precisely, membership of the epoch the checkpoint is at the start of).
		// Correct nodes should never send such checkpoints, but faulty ones could.
		if _, ok := chkp.Memberships()[0].Nodes[iss.ownID]; !ok {
			iss.logger.Log(logging.LevelWarn, "Ignoring checkpoint. Not in membership.",
				"sender", sender, "memberships", chkp.Memberships())
			return nil
		}

		chkpMembership := chkp.PreviousMembership() // TODO this is wrong and it is a vulnerability, come back to fix after discussion (issue #384)
		if chkpMembershipOffset > iss.Params.ConfigOffset {
			// cannot verify checkpoint signatures, too far ahead
			// TODO here we should externalize verification/decision to dedicated module (issue #402)
			iss.logger.Log(logging.LevelWarn, "-----------------------------------------------------\n",
				"ATTENTION: cannot verify membership of checkpoint, too far ahead, proceed with caution\n",
				"-----------------------------------------------------\n",
				"localEpoch", iss.epoch.Nr(),
				"chkpEpoch", chkp.Epoch(),
				"configOffset", iss.Params.ConfigOffset,
			)
		} else {
			chkpMembership = iss.memberships[chkpMembershipOffset]
		}
		if err := chkp.VerifyCert(iss.hashImpl, iss.chkpVerifier, chkpMembership); err != nil {
			iss.logger.Log(logging.LevelWarn, "Ignoring stable checkpoint. Certificate not valid.",
				"localEpoch", iss.epoch.Nr(),
				"chkpEpoch", chkp.Epoch(),
			)
			return nil
		}

		// Check if checkpoint contains the configured number of configurations.
		if len(chkp.Memberships()) != iss.Params.ConfigOffset+1 {
			iss.logger.Log(logging.LevelWarn, "Ignoring stable checkpoint. Membership configuration mismatch.",
				"expectedNum", iss.Params.ConfigOffset+1,
				"receivedNum", len(chkp.Memberships()))
			return nil
		}

		// Deserialize received leader selection policy. If deserialization fails, ignore the whole message.
		result, err := lsp.LeaderPolicyFromBytes(chkp.Snapshot.EpochData.LeaderPolicy)
		if err != nil {
			iss.logger.Log(logging.LevelWarn, "Error deserializing leader selection policy from checkpoint", err)
			return nil
		}
		iss.LeaderPolicy = result

		iss.logger.Log(logging.LevelWarn, "Installing state snapshot.",
			"epoch", chkp.Epoch(), "nodes", chkp.Memberships(), "leaders", iss.LeaderPolicy.Leaders())

		// Clean up global ISS state that belongs to the current epoch
		// instance that local replica got stuck with.
		iss.epochs = make(map[tt.EpochNr]*epochInfo)
		// iss.epoch = nil // This will be overwritten by initEpoch anyway.
		iss.commitLog = make(map[tt.SeqNr]*CommitLogEntry)
		iss.nextDeliveredSN = chkp.SeqNr()
		iss.newEpochSN = iss.nextDeliveredSN

		// Save the configurations obtained in the checkpoint
		// and initialize the corresponding availability submodules.
		iss.memberships = chkp.Memberships()
		iss.nextNewMembership = nil
		// Update the last stable checkpoint stored in the global ISS structure.
		iss.lastStableCheckpoint = chkp
		// Create an event to request the application module for
		// restoring its state from the snapshot received in the new
		// stable checkpoint message.
		apppbdsl.RestoreState(iss.m, iss.moduleConfig.App, chkppbtypes.StableCheckpointFromPb(chkp.Pb()))

		// Start executing the current epoch (the one the checkpoint corresponds to).
		// This must happen after the state is restored,
		// so the application has the correct state for returning the next configuration.
		iss.startEpoch(chkp.Epoch())

		// Deliver the checkpoint to the application in the proper epoch.
		// Technically, this could be made redundant, since it is the same checkpoint the application is restoring from.
		// However, it helps preserve the pattern of calls to the application where a checkpoint is explicitly delivered
		// in each epoch.
		// Note: It is important that this line goes after the call to iss.startEpoch(), as iss.startEpoch() also interacts
		//       with the application, notifying it about the new epoch.
		chkppbdsl.StableCheckpoint(iss.m,
			iss.moduleConfig.App,
			chkp.SeqNr(),
			chkp.StateSnapshot(),
			chkp.Certificate(),
		)

		// Prune the old state of all related modules.
		// Since we are loading the complete state from a checkpoint,
		// we prune all state related to anything before that checkpoint.
		pruneIndex := chkp.Epoch()
		eventpbdsl.TimerGarbageCollect(iss.m, iss.moduleConfig.Timer, tt.RetentionIndex(pruneIndex))
		factorypbdsl.GarbageCollect(iss.m, iss.moduleConfig.Checkpoint, tt.RetentionIndex(pruneIndex))
		factorypbdsl.GarbageCollect(iss.m, iss.moduleConfig.Availability, tt.RetentionIndex(pruneIndex))
		factorypbdsl.GarbageCollect(iss.m, iss.moduleConfig.Ordering, tt.RetentionIndex(pruneIndex))

		return nil
	})
	return iss.m, nil
}

// InitialStateSnapshot creates and returns a default initial snapshot that can be used to instantiate ISS.
// The created snapshot corresponds to epoch 0, without any committed transactions,
// and with the initial membership (found in params and used for epoch 0)
// repeated for all the params.ConfigOffset following epochs.
func InitialStateSnapshot(
	appState []byte,
	params *issconfig.ModuleParams,
) (*commonpbtypes.StateSnapshot, error) {

	// Create the first membership and all ConfigOffset following ones (by using the initial one).
	memberships := sliceutil.Repeat(params.InitialMembership, params.ConfigOffset+1)

	// Create the initial leader selection policy.
	var leaderPolicyData []byte
	var err error
	switch params.LeaderSelectionPolicy {
	case lsp.Simple:
		leaderPolicyData, err = lsp.NewSimpleLeaderPolicy(
			maputil.GetSortedKeys(params.InitialMembership.Nodes),
		).Bytes()
	case lsp.Blacklist:
		leaderPolicyData, err = lsp.NewBlackListLeaderPolicy(
			maputil.GetSortedKeys(params.InitialMembership.Nodes),
			issconfig.StrongQuorum(len(params.InitialMembership.Nodes)),
		).Bytes()
	default:
		return nil, fmt.Errorf("unknown leader selection policy type: %v", params.LeaderSelectionPolicy)
	}
	if err != nil {
		return nil, err
	}

	firstEpochLength := uint64(params.SegmentLength * len(params.InitialMembership.Nodes))
	return &commonpbtypes.StateSnapshot{
		AppData: appState,
		EpochData: &commonpbtypes.EpochData{
			EpochConfig: &commonpbtypes.EpochConfig{
				EpochNr:     0,
				FirstSn:     0,
				Length:      firstEpochLength,
				Memberships: memberships,
			},
			ClientProgress: commonpbtypes.ClientProgressFromPb(clientprogress.NewClientProgress(nil).Pb()),
			LeaderPolicy:   leaderPolicyData,
			// TODO: Revisit this when nil values are properly supported in generated types.
			PreviousMembership: &commonpbtypes.Membership{Nodes: make(map[t.NodeID]*commonpbtypes.NodeIdentity)},
		},
	}, nil
}

// ============================================================
// Additional protocol logic
// ============================================================

// startEpoch emits the events necessary for a new epoch to start operating.
// This includes informing the application about the new epoch and initializing all the necessary external modules
// such as availability and orderers.
func (iss *ISS) startEpoch(epochNr tt.EpochNr) {

	// Initialize the internal data structures for the new epoch.
	epoch := newEpochInfo(epochNr, iss.newEpochSN, iss.memberships[0], iss.LeaderPolicy)
	iss.epochs[epochNr] = &epoch
	iss.epoch = &epoch
	iss.logger.Log(logging.LevelInfo, "Initializing new epoch",
		"epochNr", epochNr,
		"nodes", maputil.GetSortedKeys(iss.memberships[0].Nodes),
		"leaders", iss.LeaderPolicy.Leaders(),
	)

	// Signal the new epoch to the application.
	apppbdsl.NewEpoch(iss.m, iss.moduleConfig.App, iss.epoch.Nr())

	// Initialize the new availability module.
	iss.initAvailability()

	// Initialize the orderer modules for the current epoch.
	iss.initOrderers()
}

// initAvailability emits an event for the availability module to create a new submodule
// corresponding to the current ISS epoch.
func (iss *ISS) initAvailability() {
	availabilityID := iss.moduleConfig.Availability.Then(t.ModuleID(fmt.Sprintf("%v", iss.epoch.Nr())))
	//events := make([]*eventpb.Event, 0)

	factorypbdsl.NewModule(
		iss.m,
		iss.moduleConfig.Availability,
		availabilityID,
		tt.RetentionIndex(iss.epoch.Nr()),
		&factorypbtypes.GeneratorParams{
			Type: &factorypbtypes.GeneratorParams_MultisigCollector{
				MultisigCollector: &mscpbtypes.InstanceParams{
					Membership:  iss.memberships[0],
					Limit:       5, // hardcoded right now
					MaxRequests: uint64(iss.Params.SegmentLength)},
			},
		},
	)

	apbdsl.ComputeCert(iss.m, availabilityID)
}

// initOrderers sends the SBInit event to all orderers in the current epoch.
func (iss *ISS) initOrderers() {

	leaders := iss.epoch.leaderPolicy.Leaders()
	for i, leader := range leaders {

		// Create segment.
		// The sequence proposals are all set to nil, so that the orderer proposes new availability certificates.
		proposals := freeProposals(iss.nextDeliveredSN+tt.SeqNr(i), tt.SeqNr(len(leaders)), iss.Params.SegmentLength)
		seg := orderers.NewSegment(leader, iss.epoch.Membership, proposals)
		iss.newEpochSN += tt.SeqNr(seg.Len())

		// Instantiate a new PBFT orderer.
		factorypbdsl.NewModule(iss.m, iss.moduleConfig.Ordering,
			iss.moduleConfig.Ordering.Then(t.ModuleID(fmt.Sprintf("%v", iss.epoch.Nr()))).Then(t.ModuleID(fmt.Sprintf("%v", i))),
			tt.RetentionIndex(iss.epoch.Nr()),
			orderers.InstanceParams(
				seg,
				iss.moduleConfig.Availability.Then(t.ModuleID(fmt.Sprintf("%v", iss.epoch.Nr()))),
				iss.epoch.Nr(),
				orderers.PermissiveValidityChecker,
			))

		//Add the segment to the list of segments.
		iss.epoch.Segments = append(iss.epoch.Segments, seg)

	}
}

// hasEpochCheckpoint returns true if the current epoch's checkpoint protocol has already produced a stable checkpoint.
func (iss *ISS) hasEpochCheckpoint() bool {
	return tt.SeqNr(iss.lastStableCheckpoint.Sn) == iss.epoch.FirstSN()
}

// epochFinished returns true when all the sequence numbers of the current epochs have been committed,
// the starting checkpoint of the epoch is stable, and the next new membership has been announced.
// Otherwise, returns false.
func (iss *ISS) epochFinished() bool {
	return iss.nextDeliveredSN == iss.newEpochSN && iss.hasEpochCheckpoint() && iss.nextNewMembership != nil
}

// processCommitted delivers entries from the commitLog in order of their sequence numbers.
// Whenever a new entry is inserted in the commitLog, this function must be called
// to create Deliver events for all the certificates that can be delivered to the application.
// processCommitted also triggers other internal Events like epoch transitions and state checkpointing.
func (iss *ISS) processCommitted() error {

	// Only deliver certificates if the current epoch's stable checkpoint has already been established.
	// We require this, since stable checkpoints are also delivered to the application in addition to the certificates.
	// The application may rely on the fact that each epoch starts by a stable checkpoint
	// delivered before the epoch's batches.
	if !iss.hasEpochCheckpoint() {
		return nil
	}

	// The iss.nextDeliveredSN variable always contains the lowest sequence number
	// for which no certificate has been delivered yet.
	// As long as there is an entry in the commitLog with that sequence number,
	// deliver the corresponding certificate and advance to the next sequence number.
	for iss.commitLog[iss.nextDeliveredSN] != nil {

		certData := iss.commitLog[iss.nextDeliveredSN].CertData

		var cert availabilitypb.Cert
		if len(certData) > 0 {
			if err := proto.Unmarshal(certData, &cert); err != nil {
				return fmt.Errorf("cannot unmarshal availability certificate: %w", err)
			}
		}

		if iss.commitLog[iss.nextDeliveredSN].Aborted {
			iss.LeaderPolicy.Suspect(iss.epoch.Nr(), iss.commitLog[iss.nextDeliveredSN].Suspect)
		}

		// Create a new DeliverCert event.
		var _cert *apbtypes.Cert
		if cert.Type == nil {
			_cert = &apbtypes.Cert{Type: &apbtypes.Cert_Mscs{Mscs: &mscpbtypes.Certs{}}}
		} else {
			_cert = apbtypes.CertFromPb(&cert)
		}
		isspbdsl.DeliverCert(iss.m, iss.moduleConfig.App, iss.nextDeliveredSN, _cert)
		iss.logger.Log(logging.LevelDebug, "Delivering entry.", "sn", iss.nextDeliveredSN)

		// Remove just delivered certificate from the temporary
		// store of certificates that were agreed upon out-of-order.
		delete(iss.commitLog, iss.nextDeliveredSN)

		// Increment the sequence number of the next certificate to deliver.
		iss.nextDeliveredSN++
	}

	// If the epoch is finished, transition to the next epoch.
	if iss.epochFinished() {
		return iss.advanceEpoch()
	}

	return nil
}

// advanceEpoch transitions from the current epoch to the next one.
// It must only be called when transitioning to the next epoch is possible,
// i.e., when the current epoch is completely finished.
func (iss *ISS) advanceEpoch() error {
	// Convenience variables
	oldEpochNr := iss.epoch.Nr()
	newEpochNr := oldEpochNr + 1

	iss.logger.Log(logging.LevelDebug, "Advancing epoch.",
		"epoch", oldEpochNr,
		"nextSN", iss.nextDeliveredSN,
		"numConfigs", len(iss.memberships),
	)

	// Advance the membership pipeline
	oldMembership := iss.memberships[0]
	iss.memberships = append(iss.memberships[1:], iss.nextNewMembership)
	iss.nextNewMembership = nil
	iss.LeaderPolicy = iss.LeaderPolicy.Reconfigure(maputil.GetSortedKeys(iss.memberships[0].Nodes))

	// Serialize current leader selection policy.
	leaderPolicyData, err := iss.LeaderPolicy.Bytes()
	if err != nil {
		return fmt.Errorf("could not serialize leader selection policy: %w", err)
	}

	// Start executing the new epoch.
	// This must happen before starting the checkpoint protocol, since the application
	// must already be in the new epoch when processing the state snapshot request
	// emitted by the checkpoint sub-protocol
	// (startEpoch emits an event for the application making it transition to the new epoch).
	iss.startEpoch(newEpochNr)

	// Create a new checkpoint tracker to start the checkpointing protocol.
	// This must happen after initialization of the new epoch,
	// as the sequence number the checkpoint will be associated with (iss.nextDeliveredSN)
	// is already part of the new epoch.
	// iss.nextDeliveredSN is the first sequence number *not* included in the checkpoint,
	// i.e., as sequence numbers start at 0, the checkpoint includes the first iss.nextDeliveredSN sequence numbers.
	// The membership used for the checkpoint tracker still must be the old membership.
	chkpModuleID := iss.moduleConfig.Checkpoint.Then(t.ModuleID(fmt.Sprintf("%v", newEpochNr)))
	factorypbdsl.NewModule(iss.m,
		iss.moduleConfig.Checkpoint,
		chkpModuleID,
		tt.RetentionIndex(newEpochNr),
		chkpprotos.InstanceParams(
			oldMembership,
			checkpoint.DefaultResendPeriod,
			leaderPolicyData,
			&commonpbtypes.EpochConfig{iss.epoch.Nr(), iss.epoch.FirstSN(), uint64(iss.epoch.Len()), iss.memberships},
		),
	)

	// Ask the application for a state snapshot and have it send the result directly to the checkpoint module.
	// Note that the new instance of the checkpoint protocol is not yet created at this moment,
	// but it is guaranteed to be created before the application's response.
	// This is because the NewModule event will already be enqueued for the checkpoint factory
	// when the application receives the snapshot request.
	apppbdsl.SnapshotRequest(iss.m, iss.moduleConfig.App, chkpModuleID)

	return nil
}

// verifyCert requests the availability module to verify the certificate from the preprepare message
func (iss *ISS) verifyCert(sn tt.SeqNr, data []uint8, aborted bool, leader t.NodeID) error {
	cert := &availabilitypb.Cert{}

	if err := proto.Unmarshal(data, cert); err != nil { // wrong certificate,
		iss.logger.Log(logging.LevelWarn, "failed to unmarshal cert", "err", err)
		// suspect leader
		iss.LeaderPolicy.Suspect(iss.epoch.Nr(), leader)
		// decide empty certificate
		data = []byte{}
		return iss.deliverCert(sn, data, aborted, leader)
	}

	apbdsl.VerifyCert(iss.m, iss.moduleConfig.Availability.Then(t.ModuleID(fmt.Sprintf("%v", iss.epoch.Nr()))),
		apbtypes.CertFromPb(cert), &verifyCertContext{
			sn:      sn,
			data:    data,
			aborted: aborted,
			leader:  leader,
		})
	return nil
}

// deliverCert creates a commitLog entry from an availability certificate delivered by an orderer
// and requests the computation of its hash.
// Note that applySBInstDeliver does not yet insert the entry to the commitLog. This will be done later.
// Operation continues on reception of the HashResult event.
func (iss *ISS) deliverCert(sn tt.SeqNr, data []uint8, aborted bool, leader t.NodeID) error {

	// Create a new preliminary log entry based on the delivered certificate and hash it.
	// Note that, although tempting, the hash used internally by the SB implementation cannot be re-used.
	// Apart from making the SB abstraction extremely leaky (reason enough not to do it), it would also be incorrect.
	// E.g., in PBFT, if the digest of the corresponding Preprepare message was used, the hashes at different nodes
	// might mismatch, if they commit in different PBFT views (and thus using different Preprepares).
	logEntry := &CommitLogEntry{
		Sn:       sn,
		CertData: data,
		Digest:   nil,
		Aborted:  aborted,
		Suspect:  leader,
	}

	digest := iss.computeHash(serializeLogEntryForHashing(logEntry))

	// Attach digest to entry.
	logEntry.Digest = digest

	// Insert the entry in the commitLog.
	iss.commitLog[logEntry.Sn] = logEntry

	// Deliver commitLog entries to the application in sequence number order.
	// This is relevant in the case when the sequence number of the currently SB-delivered certificate
	// is the first sequence number not yet delivered to the application.
	return iss.processCommitted()
}

func (iss *ISS) deliverCommonCheckpoint(chkpData []byte) error {
	chkp := &checkpoint.StableCheckpoint{}
	if err := chkp.Deserialize(chkpData); err != nil {
		return fmt.Errorf("could not deserialize common checkpoint: %w", err)
	}

	// Ignore old checkpoints.
	if chkp.SeqNr() <= iss.lastStableCheckpoint.SeqNr() {
		iss.logger.Log(logging.LevelDebug, "Ignoring outdated stable checkpoint.", "sn", chkp.SeqNr())
		return nil
	}

	// If this is the most recent checkpoint observed, save it.
	iss.logger.Log(logging.LevelInfo, "Agreed on new common checkpoint.",
		"currentEpoch", iss.epoch.Nr(),
		"chkpEpoch", chkp.Epoch(),
		"sn", chkp.SeqNr(),
		"replacingEpoch", iss.lastStableCheckpoint.Epoch(),
		"replacingSn", iss.lastStableCheckpoint.SeqNr(),
		"numMemberships", len(chkp.Memberships()),
	)
	iss.lastStableCheckpoint = chkp

	// Deliver the stable checkpoint (and potential batches committed in the meantime,
	// but blocked from being delivered due to this missing checkpoint) to the application.
	chkppbdsl.StableCheckpoint(iss.m,
		iss.moduleConfig.App,
		chkp.SeqNr(),
		chkp.Snapshot,
		chkp.Certificate(),
	)

	err := iss.processCommitted()
	if err != nil {
		return err
	}

	// Prune the state of all related modules.
	// The state to prune is determined according to the retention index
	// which is derived from the epoch number the new
	// stable checkpoint is associated with.
	pruneIndex := int(chkp.Epoch()) - iss.Params.RetainedEpochs
	if pruneIndex > 0 { // "> 0" and not ">= 0", since only entries strictly smaller than the index are pruned.

		// Prune timer, checkpointing, availability, and orderers.
		eventpbdsl.TimerGarbageCollect(iss.m, iss.moduleConfig.Timer, tt.RetentionIndex(pruneIndex))
		factorypbdsl.GarbageCollect(iss.m, iss.moduleConfig.Checkpoint, tt.RetentionIndex(pruneIndex))
		factorypbdsl.GarbageCollect(iss.m, iss.moduleConfig.Availability, tt.RetentionIndex(pruneIndex))
		factorypbdsl.GarbageCollect(iss.m, iss.moduleConfig.Ordering, tt.RetentionIndex(pruneIndex))

		// Prune epoch state.
		for epoch := range iss.epochs {
			if epoch < tt.EpochNr(pruneIndex) {
				delete(iss.epochs, epoch)
			}
		}

		// Start state catch-up.
		// Using a periodic PushCheckpoint event instead of directly starting a periodic re-transmission
		// of StableCheckpoint messages makes it possible to stop sending checkpoints to nodes that caught up
		// before the re-transmission is garbage-collected.
		eventpbdsl.TimerRepeat(iss.m, iss.moduleConfig.Timer,
			[]*eventpbtypes.Event{isspbevents.PushCheckpoint(iss.moduleConfig.Self)},
			types.Duration(iss.Params.CatchUpTimerPeriod),
			// Note that we are not using the current epoch number here, because it is not relevant for checkpoints.
			// Using pruneIndex makes sure that the re-transmission is stopped
			// on every stable checkpoint (when another one is started).
			tt.RetentionIndex(pruneIndex),
		)

	}

	return nil
}

func (iss *ISS) computeHash(data [][]byte) []byte {

	// One data item consists of potentially multiple byte slices.
	// Add each of them to the hash function.
	h := iss.hashImpl.New()
	for _, d := range data {
		h.Write(d)
	}

	// Return resulting digest.
	return h.Sum(nil)

}

// ============================================================
// Auxiliary functions
// ============================================================

// freeProposals returns a map of sequence numbers with associated "free" proposals (nil values).
// The returned map has length `length`, starting with sequence number `start`,
// with the difference between two consecutive sequence number being `step`.
// This function is used to compute the sequence numbers of a Segment.
// When there is `step` segments, their interleaving creates a consecutive block of sequence numbers
// that constitutes an epoch.
func freeProposals(start tt.SeqNr, step tt.SeqNr, length int) map[tt.SeqNr][]byte {
	seqNrs := make(map[tt.SeqNr][]byte, length)
	for i, nextSn := 0, start; i < length; i, nextSn = i+1, nextSn+step {
		seqNrs[nextSn] = nil
	}
	return seqNrs
}

func serializeLogEntryForHashing(entry *CommitLogEntry) [][]byte {

	// Encode integer fields.
	suspectBuf := []byte(entry.Suspect.Pb())
	snBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(snBuf, entry.Sn.Pb())

	// Encode boolean Aborted field as one byte.
	aborted := byte(0)
	if entry.Aborted {
		aborted = 1
	}

	// Put everything together in a slice and return it.
	return [][]byte{snBuf, suspectBuf, {aborted}, entry.CertData}
}

// verifyCertContext saves the context of requesting a certificate to verify from the availability layer.
type verifyCertContext struct {
	sn      tt.SeqNr
	data    []uint8
	aborted bool
	leader  t.NodeID
}
