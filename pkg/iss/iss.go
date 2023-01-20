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
	"github.com/filecoin-project/mir/pkg/events"
	factoryevents "github.com/filecoin-project/mir/pkg/factorymodule/events"
	issconfig "github.com/filecoin-project/mir/pkg/iss/config"
	lsp "github.com/filecoin-project/mir/pkg/iss/leaderselectionpolicy"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/orderers"
	"github.com/filecoin-project/mir/pkg/pb/availabilitypb"
	"github.com/filecoin-project/mir/pkg/pb/availabilitypb/mscpb"
	"github.com/filecoin-project/mir/pkg/pb/checkpointpb"
	"github.com/filecoin-project/mir/pkg/pb/commonpb"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/factorymodulepb"
	"github.com/filecoin-project/mir/pkg/pb/isspb"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
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
	epochs map[t.EpochNr]*epochInfo

	// Highest epoch numbers indicated in Checkpoint messages from each node.
	nodeEpochMap map[t.NodeID]t.EpochNr

	// The memberships for the current epoch and the params.ConfigOffset - 1 following epochs
	// (totalling params.ConfigOffset memberships).
	// E.g., if params.ConfigOffset 3 and the current epoch is 5, this field contains memberships for epoch 5, 6, and 7.
	memberships []map[t.NodeID]t.NodeAddress

	// The latest new membership obtained from the application.
	// To be included as last of the list of membership in the next new configuration.
	// The epoch number this membership corresponds to is the current epoch number + params.ConfigOffset.
	// At epoch initialization, this is set to nil. We then set it to the next membership announced by the application.
	nextNewMembership map[t.NodeID]t.NodeAddress

	// The final log of committed availability certificates.
	// For each sequence number, it holds the committed certificate (or the special abort value).
	// Each Deliver event of an orderer translates to inserting an entry in the commitLog.
	// This, in turn, leads to delivering the certificate to the application,
	// as soon as all entries with lower sequence numbers have been delivered.
	// I.e., the entries are not necessarily inserted in order of their sequence numbers,
	// but they are delivered to the application in that order.
	commitLog map[t.SeqNr]*CommitLogEntry

	// The first undelivered sequence number in the commitLog.
	// This field drives the in-order delivery of the log entries to the application.
	nextDeliveredSN t.SeqNr

	// The first sequence number to be delivered in the new epoch.
	newEpochSN t.SeqNr

	// Stores the stable checkpoint with the highest sequence number observed so far.
	// If no stable checkpoint has been observed yet, lastStableCheckpoint is initialized to a stable checkpoint value
	// corresponding to the initial state and associated with sequence number 0.
	lastStableCheckpoint *checkpoint.StableCheckpoint

	// The logic for selecting leader nodes in each epoch.
	// For details see the documentation of the LeaderSelectionPolicy type.
	// ATTENTION: The leader selection policy is stateful!
	// Must not be nil.
	LeaderPolicy lsp.LeaderSelectionPolicy
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

) (*ISS, error) {
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
		ownID:                ownID,
		moduleConfig:         moduleConfig,
		Params:               params,
		hashImpl:             hashImpl,
		chkpVerifier:         chkpVerifier,
		logger:               logger,
		epoch:                nil, // Initialized later.
		epochs:               make(map[t.EpochNr]*epochInfo),
		nodeEpochMap:         make(map[t.NodeID]t.EpochNr),
		memberships:          startingChkp.Memberships(),
		nextNewMembership:    nil,
		commitLog:            make(map[t.SeqNr]*CommitLogEntry),
		nextDeliveredSN:      startingChkp.SeqNr(),
		newEpochSN:           startingChkp.SeqNr(),
		lastStableCheckpoint: startingChkp,
		LeaderPolicy:         leaderPolicy,
		// TODO: Make sure that verification of the stable checkpoint certificate for epoch 0 is handled properly.
		//       (Probably "always valid", if the membership is right.) There is no epoch -1 with nodes to sign it.
	}

	// Initialize the first epoch.
	// The corresponding application state will be loaded upon init.
	// TODO: Make leader policy part of checkpoint.
	iss.logger.Log(logging.LevelInfo, "Initializing new epoch",
		"epochNr", startingChkp.Epoch(), "numNodes", len(startingChkp.Memberships()[0]))

	// Return the initialized protocol module.
	return iss, nil
}

// InitialStateSnapshot creates and returns a default initial snapshot that can be used to instantiate ISS.
// The created snapshot corresponds to epoch 0, without any committed transactions,
// and with the initial membership (found in params and used for epoch 0)
// repeated for all the params.ConfigOffset following epochs.
func InitialStateSnapshot(
	appState []byte,
	params *issconfig.ModuleParams,
) (*commonpb.StateSnapshot, error) {

	// Create the first membership and all ConfigOffset following ones (by using the initial one).
	memberships := make([]map[t.NodeID]t.NodeAddress, params.ConfigOffset+1)
	for i := 0; i < params.ConfigOffset+1; i++ {
		memberships[i] = params.InitialMembership
	}

	// Create the initial leader selection policy.
	var leaderPolicyData []byte
	var err error
	switch params.LeaderSelectionPolicy {
	case lsp.Simple:
		leaderPolicyData, err = lsp.NewSimpleLeaderPolicy(
			maputil.GetSortedKeys(params.InitialMembership),
		).Bytes()
	case lsp.Blacklist:
		leaderPolicyData, err = lsp.NewBlackListLeaderPolicy(
			maputil.GetSortedKeys(params.InitialMembership),
			issconfig.StrongQuorum(len(params.InitialMembership)),
		).Bytes()
	default:
		return nil, fmt.Errorf("unknown leader selection policy type: %v", params.LeaderSelectionPolicy)
	}
	if err != nil {
		return nil, err
	}

	firstEpochLength := params.SegmentLength * len(params.InitialMembership)
	return &commonpb.StateSnapshot{
		AppData: appState,
		EpochData: &commonpb.EpochData{
			EpochConfig:        events.EpochConfig(0, 0, firstEpochLength, memberships),
			ClientProgress:     clientprogress.NewClientProgress(nil).Pb(),
			LeaderPolicy:       leaderPolicyData,
			PreviousMembership: nil,
		},
	}, nil
}

// ============================================================
// Protocol Interface implementation
// ============================================================

// ApplyEvents receives a list of events, processes them sequentially, and returns a list of resulting events.
func (iss *ISS) ApplyEvents(eventsIn *events.EventList) (*events.EventList, error) {
	return modules.ApplyEventsSequentially(eventsIn, iss.ApplyEvent)
}

// ApplyEvent receives one event and applies it to the ISS protocol state machine, potentially altering its state
// and producing a (potentially empty) list of more events to be applied to other modules.
func (iss *ISS) ApplyEvent(event *eventpb.Event) (*events.EventList, error) {

	switch e := event.Type.(type) {
	case *eventpb.Event_Init:
		return iss.applyInit()
	case *eventpb.Event_NewConfig:
		return iss.applyNewConfig(e.NewConfig)
	case *eventpb.Event_Checkpoint:
		switch e := e.Checkpoint.Type.(type) {
		case *checkpointpb.Event_EpochProgress:
			return iss.applyEpochProgress(e.EpochProgress)
		case *checkpointpb.Event_StableCheckpoint:
			return iss.applyStableCheckpoint((*checkpoint.StableCheckpoint)(e.StableCheckpoint))
		default:
			return nil, fmt.Errorf("unknown checkpoint event type: %T", e)
		}
	case *eventpb.Event_Iss: // The ISS event type wraps all ISS-specific events.
		switch issEvent := e.Iss.Type.(type) {
		case *isspb.ISSEvent_SbDeliver:
			return iss.applySBInstDeliver(issEvent.SbDeliver)
		case *isspb.ISSEvent_PushCheckpoint:
			return iss.applyPushCheckpoint()
		default:
			return nil, fmt.Errorf("unknown ISS event type: %T", issEvent)
		}
	case *eventpb.Event_MessageReceived:
		return iss.applyMessageReceived(e.MessageReceived)

	default:
		return nil, fmt.Errorf("unknown protocol (ISS) event type: %T", event.Type)
	}
}

// The ImplementsModule method only serves the purpose of indicating that this is a Module and must not be called.
func (iss *ISS) ImplementsModule() {}

// ============================================================
// Event application
// ============================================================

// applyInit initializes the ISS protocol.
// This event is only expected to be applied once at startup,
// after all the events stored in the WAL have been applied and before any other event has been applied.
// (At this time, the WAL is not used. TODO: Update this when wal is implemented.)
func (iss *ISS) applyInit() (*events.EventList, error) {
	eventsOut := events.EmptyList()

	// Initialize application state according to the initial checkpoint.
	eventsOut.PushBack(events.AppRestoreState(iss.moduleConfig.App, iss.lastStableCheckpoint.Pb()))

	// Start the first epoch (not necessarily epoch 0, depending on the starting checkpoint).
	eventsOut.PushBackList(iss.startEpoch(iss.lastStableCheckpoint.Epoch()))

	return eventsOut, nil
}

// applySBInstDeliver processes the event of an SB instance delivering a certificate (or the special abort value)
// for a sequence number. It creates a corresponding commitLog entry and requests the computation of its hash.
// Note that applySBInstDeliver does not yet insert the entry to the commitLog. This will be done later.
// Operation continues on reception of the HashResult event.
func (iss *ISS) applySBInstDeliver(deliver *isspb.SBDeliver) (*events.EventList, error) {

	// Create a new preliminary log entry based on the delivered certificate and hash it.
	// Note that, although tempting, the hash used internally by the SB implementation cannot be re-used.
	// Apart from making the SB abstraction extremely leaky (reason enough not to do it), it would also be incorrect.
	// E.g., in PBFT, if the digest of the corresponding Preprepare message was used, the hashes at different nodes
	// might mismatch, if they commit in different PBFT views (and thus using different Preprepares).
	logEntry := &CommitLogEntry{
		Sn:       t.SeqNr(deliver.Sn),
		CertData: deliver.CertData,
		Digest:   nil,
		Aborted:  deliver.Aborted,
		Suspect:  t.NodeID(deliver.Leader),
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

func (iss *ISS) applyNewConfig(config *eventpb.NewConfig) (*events.EventList, error) {
	eventsOut := events.EmptyList()

	iss.logger.Log(logging.LevelDebug, "Received new configuration.",
		"numNodes", len(config.Membership.Membership), "epochNr", config.EpochNr, "currentEpoch", iss.epoch.Nr())

	// Check whether this event is not outdated.
	// This can (and did) happen in a corner case where the state gets restored from a checkpoint
	// while a NewConfig event is already in the pipeline.
	// Note that config.EpochNr is NOT the epoch where this new configuration is used,
	// but it is the epoch in which this new configuration is supposed to be received
	// (it will be used iss.params.ConfigOffset epochs later).
	if t.EpochNr(config.EpochNr) < iss.epoch.Nr() {
		iss.logger.Log(logging.LevelWarn, "Ignoring outdated membership.",
			"notificationEpoch", config.EpochNr, "currentEpoch", iss.epoch.Nr())
		return eventsOut, nil
	}

	// Sanity check.
	if iss.nextNewMembership != nil {
		return nil, fmt.Errorf("already have a new membership for epoch %v", iss.epoch.Nr())
	}

	// Convenience variable.
	newMembershipEpoch := iss.epoch.Nr() + t.EpochNr(len(iss.memberships))

	// Save the new configuration.
	iss.nextNewMembership = t.Membership(config.Membership)

	iss.logger.Log(logging.LevelDebug, "Adding configuration",
		"forEpoch", newMembershipEpoch,
		"currentEpoch", iss.epoch.Nr(),
		"newConfigNodes", maputil.GetSortedKeys(iss.nextNewMembership))

	// Advance to the next epoch if this configuration was the last missing bit.
	if iss.epochFinished() {
		l, err := iss.advanceEpoch()
		if err != nil {
			return nil, err
		}
		eventsOut.PushBackList(l)
	}

	return eventsOut, nil
}

// applyEpochProgress handles the event informing ISS that a node reached a certain epoch.
// This event can be, for example, emitted by the checkpoint protocol
// when it detects a node having reached a checkpoint for a certain epoch.
func (iss *ISS) applyEpochProgress(epochProgress *checkpointpb.EpochProgress) (*events.EventList, error) {

	// Remember the highest epoch number for each node to detect
	// later if the remote node is delayed too much and requires
	// assistance in order to catch up through state transfer.
	epochNr := t.EpochNr(epochProgress.Epoch)
	nodeID := t.NodeID(epochProgress.NodeId)
	if iss.nodeEpochMap[nodeID] < epochNr {
		iss.nodeEpochMap[nodeID] = epochNr
	}

	return events.EmptyList(), nil
}

// applyStableCheckpoint handles a new stable checkpoint produced by the checkpoint protocol.
// It performs the necessary cleanup, mostly garbage-collecting state subsumed by the stable checkpoint.
func (iss *ISS) applyStableCheckpoint(stableCheckpoint *checkpoint.StableCheckpoint) (*events.EventList, error) {
	eventsOut := events.EmptyList()

	if stableCheckpoint.SeqNr() > iss.lastStableCheckpoint.SeqNr() {
		// If this is the most recent checkpoint observed, save it.
		iss.logger.Log(logging.LevelInfo, "New stable checkpoint.",
			"epoch", stableCheckpoint.Epoch(),
			"sn", stableCheckpoint.SeqNr(),
			"replacingEpoch", iss.lastStableCheckpoint.Epoch(),
			"replacingSn", iss.lastStableCheckpoint.SeqNr(),
			"numMemberships", len(stableCheckpoint.Memberships()),
		)
		iss.lastStableCheckpoint = stableCheckpoint

		// Deliver the stable checkpoint (and potential batches committed in the meantime,
		// but blocked from being delivered due to this missing checkpoint) to the application.
		eventsOut.PushBack(chkpprotos.StableCheckpointEvent(iss.moduleConfig.App, stableCheckpoint.Pb()))
		pcResult, err := iss.processCommitted()
		if err != nil {
			return nil, err
		}
		eventsOut.PushBackList(pcResult)

		// Prune the state of all related modules.
		// The state to prune is determined according to the retention index
		// which is derived from the epoch number the new
		// stable checkpoint is associated with.
		pruneIndex := int(stableCheckpoint.Epoch()) - iss.Params.RetainedEpochs
		if pruneIndex > 0 { // "> 0" and not ">= 0", since only entries strictly smaller than the index are pruned.

			// Prune timer, checkpointing, availability, and orderers.
			eventsOut.PushBackSlice([]*eventpb.Event{
				events.TimerGarbageCollect(iss.moduleConfig.Timer, t.RetentionIndex(pruneIndex)),
				factoryevents.GarbageCollect(iss.moduleConfig.Checkpoint, t.RetentionIndex(pruneIndex)),
				factoryevents.GarbageCollect(iss.moduleConfig.Availability, t.RetentionIndex(pruneIndex)),
				factoryevents.GarbageCollect(iss.moduleConfig.Ordering, t.RetentionIndex(pruneIndex)),
			})
			// TODO: Make EventList.PushBack accept a variable number of arguments and use it here.

			// Prune epoch state.
			for epoch := range iss.epochs {
				if epoch < t.EpochNr(pruneIndex) {
					delete(iss.epochs, epoch)
				}
			}

			// Start state catch-up.
			// Using a periodic PushCheckpoint event instead of directly starting a periodic re-transmission
			// of StableCheckpoint messages makes it possible to stop sending checkpoints to nodes that caught up
			// before the re-transmission is garbage-collected.
			eventsOut.PushBack(events.TimerRepeat(
				iss.moduleConfig.Timer,
				[]*eventpb.Event{PushCheckpoint(iss.moduleConfig.Self)},
				t.TimeDuration(iss.Params.CatchUpTimerPeriod),

				// Note that we are not using the current epoch number here, because it is not relevant for checkpoints.
				// Using pruneIndex makes sure that the re-transmission is stopped
				// on every stable checkpoint (when another one is started).
				t.RetentionIndex(pruneIndex),
			))

		}
	} else {
		iss.logger.Log(logging.LevelDebug, "Ignoring outdated stable checkpoint.", "sn", stableCheckpoint.SeqNr())
	}

	return eventsOut, nil
}

// applyPushCheckpoint handles the periodic event of pushing the latest stable checkpoint
// to nodes that seem to have fallen behind.
// This event is produced periodically by the timer (with a configurable period).
// Note that (in the good case), no node seems to have fallen behind and this handler does nothing.
func (iss *ISS) applyPushCheckpoint() (*events.EventList, error) {

	// Send the latest stable checkpoint to potentially
	// delayed nodes. The set of nodes to send the latest
	// stable checkpoint to is determined based on the
	// highest epoch number in the messages received from
	// the node so far. If the highest epoch number we
	// have heard from the node is more than params.RetainedEpochs
	// behind, then that node is likely to be left behind
	// and needs the stable checkpoint in order to start
	// catching up with state transfer.
	var delayed []t.NodeID
	lastStableCheckpointEpoch := iss.lastStableCheckpoint.Epoch()
	for n := range iss.epoch.nodeIDs {
		if lastStableCheckpointEpoch > iss.nodeEpochMap[n]+t.EpochNr(iss.Params.RetainedEpochs) {
			delayed = append(delayed, n)
		}
	}

	if len(delayed) > 0 {
		iss.logger.Log(logging.LevelDebug, "Pushing state to nodes.",
			"delayed", delayed, "numNodes", len(iss.epoch.nodeIDs), "nodeEpochMap", iss.nodeEpochMap)
	}

	m := StableCheckpointMessage(iss.lastStableCheckpoint)
	return events.ListOf(events.SendMessage(iss.moduleConfig.Net, m, delayed)), nil
}

// applyMessageReceived applies a message received over the network.
// Currently, only the stable checkpoint message (used for catching up) is used.
func (iss *ISS) applyMessageReceived(messageReceived *eventpb.MessageReceived) (*events.EventList, error) {

	// Convenience variables used for readability.
	message := messageReceived.Msg
	from := t.NodeID(messageReceived.From)

	switch msg := message.Type.(type) {
	case *messagepb.Message_Iss:
		switch msg := msg.Iss.Type.(type) {
		case *isspb.ISSMessage_StableCheckpoint:
			return iss.applyStableCheckpointMessage(msg.StableCheckpoint, from)
		default:
			return nil, fmt.Errorf("unknown ISS message type: %T", msg)
		}
	default:
		return nil, fmt.Errorf("unknown message type: %T", msg)
	}
}

// applyStableCheckpointMessage processes a received StableCheckpoint message
// by creating a request for verifying the signatures in the included checkpoint certificate.
// The actual processing then happens in applyStableCheckpointSigVerResult.
func (iss *ISS) applyStableCheckpointMessage(chkpPb *checkpointpb.StableCheckpoint, _ t.NodeID) (*events.EventList, error) {

	eventsOut := events.EmptyList()

	chkp := checkpoint.StableCheckpointFromPb(chkpPb)

	// TODO: Technically this is wrong.
	//       The memberhips information in the checkpint is used to verify the checkpoint itself.
	//       This makes it possible to construct an arbitrary valid checkpoint.
	//       Use an independent local source of memberhip information instead.
	if err := chkp.VerifyCert(iss.hashImpl, iss.chkpVerifier, chkp.PreviousMembership()); err != nil {
		iss.logger.Log(logging.LevelWarn, "Ignoring stable checkpoint. Certificate don walid.",
			"localEpoch", iss.epoch.Nr(),
			"chkpEpoch", chkp.Epoch(),
		)
		return eventsOut, nil
	}

	// Check if checkpoint contains the configured number of configurations.
	if len(chkp.Memberships()) != iss.Params.ConfigOffset+1 {
		iss.logger.Log(logging.LevelWarn, "Ignoring stable checkpoint. Membership configuration mismatch.",
			"expectedNum", iss.Params.ConfigOffset+1,
			"receivedNum", len(chkp.Memberships()))
		return eventsOut, nil
	}

	// Check how far the received stable checkpoint is ahead of the local node's state.
	if chkp.Epoch() <= iss.epoch.Nr()+1 {
		// Ignore stable checkpoints that are not far enough
		// ahead of the current state of the local node.
		return events.EmptyList(), nil
	}

	// Deserialize received leader selection policy. If deserialization fails, ignore the whole message.
	result, err := lsp.LeaderPolicyFromBytes(chkp.Snapshot.EpochData.LeaderPolicy)
	if err != nil {
		iss.logger.Log(logging.LevelWarn, "Error deserializing leader selection policy from checkpoint", err)
		return events.EmptyList(), nil
	}
	iss.LeaderPolicy = result

	iss.logger.Log(logging.LevelWarn, "Installing state snapshot.",
		"epoch", chkp.Epoch(), "nodes", chkp.Memberships(), "leaders", iss.LeaderPolicy.Leaders())

	// Clean up global ISS state that belongs to the current epoch
	// instance that local replica got stuck with.
	iss.epochs = make(map[t.EpochNr]*epochInfo)
	// iss.epoch = nil // This will be overwritten by initEpoch anyway.
	iss.commitLog = make(map[t.SeqNr]*CommitLogEntry)
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
	eventsOut.PushBack(events.AppRestoreState(iss.moduleConfig.App, chkp.Pb()))

	// Start executing the current epoch (the one the checkpoint corresponds to).
	// This must happen after the state is restored,
	// so the application has the correct state for returning the next configuration.
	eventsOut.PushBackList(iss.startEpoch(chkp.Epoch()))

	// Prune the old state of all related modules.
	// Since we are loading the complete state from a checkpoint,
	// we prune all state related to anything before that checkpoint.
	eventsOut.PushBackSlice([]*eventpb.Event{
		events.TimerGarbageCollect(iss.moduleConfig.Timer, t.RetentionIndex(chkp.Epoch())),
		factoryevents.GarbageCollect(iss.moduleConfig.Checkpoint, t.RetentionIndex(chkp.Epoch())),
		factoryevents.GarbageCollect(iss.moduleConfig.Availability, t.RetentionIndex(chkp.Epoch())),
		factoryevents.GarbageCollect(iss.moduleConfig.Ordering, t.RetentionIndex(chkp.Epoch())),
	})

	return eventsOut, nil
}

// ============================================================
// Additional protocol logic
// ============================================================

// startEpoch emits the events necessary for a new epoch to start operating.
// This includes informing the application about the new epoch and initializing all the necessary external modules
// such as availability and orderers.
func (iss *ISS) startEpoch(epochNr t.EpochNr) *events.EventList {
	eventsOut := events.EmptyList()

	// Initialize the internal data structures for the new epoch.
	nodeIDs := maputil.GetSortedKeys(iss.memberships[0])
	epoch := newEpochInfo(epochNr, iss.newEpochSN, nodeIDs, iss.LeaderPolicy)
	iss.epochs[epochNr] = &epoch
	iss.epoch = &epoch
	iss.logger.Log(logging.LevelInfo, "Initializing new epoch", "epochNr", epochNr, "nodes", nodeIDs, "leaders", iss.LeaderPolicy.Leaders())

	// Signal the new epoch to the application.
	eventsOut.PushBack(events.NewEpoch(iss.moduleConfig.App, iss.epoch.Nr()))

	// Initialize the new availability module.
	eventsOut.PushBack(iss.initAvailability())

	// Initialize the orderer modules for the current epoch.
	eventsOut.PushBackList(iss.initOrderers())

	return eventsOut
}

// initAvailability emits an event for the availability module to create a new submodule
// corresponding to the current ISS epoch.
func (iss *ISS) initAvailability() *eventpb.Event {
	return factoryevents.NewModule(
		iss.moduleConfig.Availability,
		iss.moduleConfig.Availability.Then(t.ModuleID(fmt.Sprintf("%v", iss.epoch.Nr()))),
		t.RetentionIndex(iss.epoch.Nr()),
		&factorymodulepb.GeneratorParams{Type: &factorymodulepb.GeneratorParams_MultisigCollector{
			MultisigCollector: &mscpb.InstanceParams{Membership: t.MembershipPb(iss.memberships[0])},
		}},
	)
}

// initOrderers sends the SBInit event to all orderers in the current epoch.
func (iss *ISS) initOrderers() *events.EventList {

	eventsOut := events.EmptyList()

	leaders := iss.epoch.leaderPolicy.Leaders()
	for i, leader := range leaders {

		// Create segment.
		seg := &orderers.Segment{
			Leader:     leader,
			Membership: maputil.GetSortedKeys(iss.epoch.nodeIDs),
			SeqNrs: sequenceNumbers(
				iss.nextDeliveredSN+t.SeqNr(i),
				t.SeqNr(len(leaders)),
				iss.Params.SegmentLength),
		}
		iss.newEpochSN += t.SeqNr(len(seg.SeqNrs))

		// Instantiate a new PBFT orderer.
		eventsOut.PushBack(factoryevents.NewModule(
			iss.moduleConfig.Ordering,
			iss.moduleConfig.Ordering.Then(t.ModuleID(fmt.Sprintf("%v", iss.epoch.Nr()))).Then(t.ModuleID(fmt.Sprintf("%v", i))),
			t.RetentionIndex(iss.epoch.Nr()),
			orderers.InstanceParams(
				seg,
				iss.moduleConfig.Availability.Then(t.ModuleID(fmt.Sprintf("%v", iss.epoch.Nr()))),
				iss.epoch.Nr(),
			),
		))

		//Add the segment to the list of segments.
		iss.epoch.Segments = append(iss.epoch.Segments, seg)

	}

	return eventsOut
}

// hasEpochCheckpoint returns true if the current epoch's checkpoint protocol has already produced a stable checkpoint.
func (iss *ISS) hasEpochCheckpoint() bool {
	return t.SeqNr(iss.lastStableCheckpoint.Sn) == iss.epoch.FirstSN()
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
func (iss *ISS) processCommitted() (*events.EventList, error) {
	eventsOut := events.EmptyList()

	// Only deliver certificates if the current epoch's stable checkpoint has already been established.
	// We require this, since stable checkpoints are also delivered to the application in addition to the certificates.
	// The application may rely on the fact that each epoch starts by a stable checkpoint
	// delivered before the epoch's batches.
	if !iss.hasEpochCheckpoint() {
		return eventsOut, nil
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
				return nil, fmt.Errorf("cannot unmarshal availability certificate: %w", err)
			}
		}

		if iss.commitLog[iss.nextDeliveredSN].Aborted {
			iss.LeaderPolicy.Suspect(iss.epoch.Nr(), iss.commitLog[iss.nextDeliveredSN].Suspect)
		}
		// Create a new DeliverCert event.
		eventsOut.PushBack(events.DeliverCert(iss.moduleConfig.App, iss.nextDeliveredSN, &cert))

		// Output debugging information.
		iss.logger.Log(logging.LevelDebug, "Delivering entry.", "sn", iss.nextDeliveredSN)

		// Remove just delivered certificate from the temporary
		// store of certificates that were agreed upon out-of-order.
		delete(iss.commitLog, iss.nextDeliveredSN)

		// Increment the sequence number of the next certificate to deliver.
		iss.nextDeliveredSN++
	}

	// If the epoch is finished, transition to the next epoch.
	if iss.epochFinished() {
		l, err := iss.advanceEpoch()
		if err != nil {
			return nil, err
		}
		eventsOut.PushBackList(l)
	}

	return eventsOut, nil
}

// advanceEpoch transitions from the current epoch to the next one.
// It must only be called when transitioning to the next epoch is possible,
// i.e., when the current epoch is completely finished.
func (iss *ISS) advanceEpoch() (*events.EventList, error) {
	eventsOut := events.EmptyList()

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
	iss.LeaderPolicy = iss.LeaderPolicy.Reconfigure(maputil.GetSortedKeys(iss.memberships[0]))

	// Start executing the new epoch.
	// This must happen before starting the checkpoint protocol, since the application
	// must already be in the new epoch when processing the state snapshot request
	// emitted by the checkpoint sub-protocol
	// (startEpoch emits an event for the application making it transition to the new epoch).
	eventsOut.PushBackList(iss.startEpoch(newEpochNr))

	// Serialize current leader selection policy.
	leaderPolicyData, err := iss.LeaderPolicy.Bytes()
	if err != nil {
		return nil, fmt.Errorf("could not serialize leader selection policy: %w", err)
	}

	// Create a new checkpoint tracker to start the checkpointing protocol.
	// This must happen after initialization of the new epoch,
	// as the sequence number the checkpoint will be associated with (iss.nextDeliveredSN)
	// is already part of the new epoch.
	// iss.nextDeliveredSN is the first sequence number *not* included in the checkpoint,
	// i.e., as sequence numbers start at 0, the checkpoint includes the first iss.nextDeliveredSN sequence numbers.
	// The membership used for the checkpoint tracker still must be the old membership.
	chkpModuleID := iss.moduleConfig.Checkpoint.Then(t.ModuleID(fmt.Sprintf("%v", newEpochNr)))
	eventsOut.PushBack(factoryevents.NewModule(
		iss.moduleConfig.Checkpoint,
		chkpModuleID,
		t.RetentionIndex(newEpochNr),
		chkpprotos.InstanceParams(
			oldMembership,
			checkpoint.DefaultResendPeriod,
			leaderPolicyData,
			events.EpochConfig(iss.epoch.Nr(), iss.epoch.FirstSN(), iss.epoch.Len(), iss.memberships),
		),
	))

	// Ask the application for a state snapshot and have it send the result directly to the checkpoint module.
	// Note that the new instance of the checkpoint protocol is not yet created at this moment,
	// but it is guaranteed to be created before the application's response.
	// This is because the NewModule event will already be enqueued for the checkpoint factory
	// when the application receives the snapshot request.
	eventsOut.PushBack(events.AppSnapshotRequest(iss.moduleConfig.App, chkpModuleID))

	return eventsOut, nil
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

// sequenceNumbers returns a list of sequence numbers of length `length`,
// starting with sequence number `start`, with the difference between two consecutive sequence number being `step`.
// This function is used to compute the sequence numbers of a Segment.
// When there is `step` segments, their interleaving creates a consecutive block of sequence numbers
// that constitutes an epoch.
func sequenceNumbers(start t.SeqNr, step t.SeqNr, length int) []t.SeqNr {
	seqNrs := make([]t.SeqNr, length)
	for i, nextSn := 0, start; i < length; i, nextSn = i+1, nextSn+step {
		seqNrs[i] = nextSn
	}
	return seqNrs
}

// reqStrKey takes a request reference and transforms it to a string for using as a map key.
func reqStrKey(req *requestpb.HashedRequest) string {
	return fmt.Sprintf("%v-%d.%v", req.Req.ClientId, req.Req.ReqNo, req.Digest)
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
