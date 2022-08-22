/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package iss contains the implementation of the ISS protocol, the new generation of Mir.
// For the details of the protocol, see (TODO).
// To use ISS, instantiate it by calling `iss.New` and use it as the Protocol module when instantiating a mir.Node.
// A default configuration (to pass, among other arguments, to `iss.New`)
// can be obtained from `iss.DefaultConfig`.
//
// Current status: This package is currently being implemented and is not yet functional.
package iss

import (
	"encoding/binary"
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/filecoin-project/mir/pkg/contextstore"
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/factorymodule"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/messagebuffer"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/availabilitypb"
	"github.com/filecoin-project/mir/pkg/pb/availabilitypb/mscpb"
	"github.com/filecoin-project/mir/pkg/pb/commonpb"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/factorymodulepb"
	"github.com/filecoin-project/mir/pkg/pb/isspb"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	"github.com/filecoin-project/mir/pkg/serializing"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
)

// ============================================================
// Auxiliary types
// ============================================================

// ModuleConfig contains the names of modules ISS depends on.
// The corresponding modules are expected by ISS to be stored under these keys by the Node.
type ModuleConfig struct {
	Self         t.ModuleID
	Net          t.ModuleID
	App          t.ModuleID
	Wal          t.ModuleID
	Hasher       t.ModuleID
	Crypto       t.ModuleID
	Timer        t.ModuleID
	Availability t.ModuleID
}

func DefaultModuleConfig() ModuleConfig {
	return ModuleConfig{
		Self:         "iss",
		Net:          "net",
		App:          "app",
		Wal:          "wal",
		Hasher:       "hasher",
		Crypto:       "crypto",
		Timer:        "timer",
		Availability: "availability",
	}
}

// The segment type represents an ISS segment.
// It is used to parametrize an orderer (i.e. the SB instance).
type segment struct {

	// The leader node of the orderer.
	Leader t.NodeID

	// List of all nodes executing the orderer implementation.
	Membership []t.NodeID

	// List of sequence numbers for which the orderer is responsible.
	// This is the actual "segment" of the commit log.
	SeqNrs []t.SeqNr
}

// The CommitLogEntry type represents an entry of the commit log, the final output of the ordering process.
// Whenever an orderer delivers an availability certificate (or a special abort value),
// it is inserted to the commit log in form of a commitLogEntry.
type CommitLogEntry struct {
	// Sequence number at which this entry has been ordered.
	Sn t.SeqNr

	// The delivered availability certificate data.
	// TODO: Replace by actual certificate when deterministic serialization of certificates is implemented.
	CertData []byte

	// The digest (hash) of the entry.
	Digest []byte

	// A flag indicating whether this entry is an actual certificate (false)
	// or whether the orderer delivered a special abort value (true).
	Aborted bool

	// In case Aborted is true, this field indicates the ID of the node
	// that is suspected to be the reason for the orderer aborting (usually the leader).
	// This information can be used by the leader selection policy at epoch transition.
	Suspect t.NodeID
}

// ============================================================
// ISS type and constructor
// ============================================================

// The ISS type represents the ISS protocol module to be used when instantiating a node.
// The type should not be instantiated directly, but only properly initialized values
// returned from the New() function should be used.
type ISS struct {

	// --------------------------------------------------------------------------------
	// These fields should only be set at initialization and remain static
	// --------------------------------------------------------------------------------

	// IDs of modules ISS interacts with.
	moduleConfig ModuleConfig

	// The ID of the node executing this instance of the protocol.
	ownID t.NodeID

	// Logger the ISS implementation uses to output log messages.
	// This is mostly for debugging - not to be confused with the commit log.
	logger logging.Logger

	// --------------------------------------------------------------------------------
	// These fields might change from epoch to epoch. Modified only by initEpoch()
	// --------------------------------------------------------------------------------

	// The ISS configuration parameters (e.g. segment length, proposal frequency etc...)
	// passed to New() when creating an ISS protocol instance.
	// TODO: Make it possible to change this dynamically.
	config *ModuleParams

	// The current epoch instance.
	epoch *epochInfo

	// Epoch instances.
	epochs map[t.EpochNr]*epochInfo

	// Highest epoch numbers indicated in Checkpoint messages from each node.
	nodeEpochMap map[t.NodeID]t.EpochNr

	// --------------------------------------------------------------------------------
	// These fields are modified throughout an epoch.
	// TODO: Move them into `epochInfo`?
	// --------------------------------------------------------------------------------

	// Buffers representing a backlog of messages destined to future epochs.
	// A node that already transitioned to a newer epoch might send messages,
	// while this node is slightly behind (still in an older epoch) and cannot process these messages yet.
	// Such messages end up in this buffer (if there is buffer space) for later processing.
	// The buffer is checked after each epoch transition.
	messageBuffers map[t.NodeID]*messagebuffer.MessageBuffer

	// The final log of committed availability certificates.
	// For each sequence number, it holds the committed certificate (or the special abort value).
	// Each Deliver event of an orderer translates to inserting an entry in the commitLog.
	// This, in turn, leads to delivering the certificate to the application,
	// as soon as all entries with lower sequence numbers have been delivered.
	// I.e., the entries are not necessarily inserted in order of their sequence numbers,
	// but they are delivered to the application in that order.
	commitLog map[t.SeqNr]*CommitLogEntry

	// CommitLogEntries for which the hash has been requested, but not yet computed.
	// When an orderer delivers a certificate, ISS creates a CommitLogEntry with a nil Hash, stores it here,
	// and creates a HashRequest. When the HashResult arrives, ISS removes the CommitLogEntry from unhashedLogEntries,
	// fills in the missing hash, and inserts it to the commitLog.
	unhashedLogEntries map[t.SeqNr]*CommitLogEntry

	// The first undelivered sequence number in the commitLog.
	// This field drives the in-order delivery of the log entries to the application.
	nextDeliveredSN t.SeqNr

	// The first sequence number to be delivered in the new epoch.
	newEpochSN t.SeqNr

	// Stores the stable checkpoint with the highest sequence number observed so far.
	// If no stable checkpoint has been observed yet, lastStableCheckpoint is initialized to a stable checkpoint value
	// corresponding to the initial state and associated with sequence number 0.
	lastStableCheckpoint *isspb.StableCheckpoint

	// TODO: Finish and document this.
	configurations []map[t.NodeID]t.NodeAddress

	// Stores the context of request-response type invocations of other modules.
	contextStore contextstore.ContextStore[any]
}

// New returns a new initialized instance of the ISS protocol module to be used when instantiating a mir.Node.
// Arguments:
//   - ownID:  the ID of the node being instantiated with ISS.
//   - config: ISS protocol-specific configuration (e.g. segment length, proposal frequency etc...).
//     see the documentation of the ModuleParams type for details.
//   - logger: Logger the ISS implementation uses to output log messages.
func New(ownID t.NodeID, moduleConfig ModuleConfig, params *ModuleParams, initialState *commonpb.StateSnapshot, logger logging.Logger) (*ISS, error) {

	// Check whether the passed configuration is valid.
	if err := CheckParams(params); err != nil {
		return nil, fmt.Errorf("invalid ISS configuration: %w", err)
	}

	memberships := make([]map[t.NodeID]t.NodeAddress, params.ConfigOffset+1)
	for i := 0; i < params.ConfigOffset+1; i++ {
		memberships[i] = t.Membership(initialState.Configuration.Memberships[i])
	}

	// Initialize a new ISS object.
	iss := &ISS{
		// Static fields
		ownID:        ownID,
		moduleConfig: moduleConfig,
		logger:       logger,

		// Fields modified only by initEpoch
		config:       params,
		epochs:       make(map[t.EpochNr]*epochInfo),
		nodeEpochMap: make(map[t.NodeID]t.EpochNr),

		// Fields modified throughout an epoch
		commitLog:          make(map[t.SeqNr]*CommitLogEntry),
		unhashedLogEntries: make(map[t.SeqNr]*CommitLogEntry),
		nextDeliveredSN:    0,
		newEpochSN:         0,
		messageBuffers: messagebuffer.NewBuffers(
			// Create a message buffer for everyone except for myself.
			removeNodeID(maputil.GetSortedKeys(params.InitialMembership), ownID),
			params.MsgBufCapacity,
			logging.Decorate(logger, "Msgbuf: "),
		),
		lastStableCheckpoint: &isspb.StableCheckpoint{
			Sn:       0,
			Snapshot: initialState,
			Cert:     map[string][]byte{},
			// TODO: When the storing of actual application state is implemented, some encoding of "initial state"
			//       will have to be set here. E.g., an empty byte slice could be defined as "initial state" and
			//       the application required to interpret it as such.

			// TODO: Make sure that verification of the stable checkpoint certificate for epoch 0 is handled properly.
			//       (Probably "always valid", if the membership is right.) There is no epoch -1 with nodes to sign it.
		},
		configurations: memberships,

		contextStore: contextstore.NewSequentialContextStore[any](),
	}

	// Initialize the first epoch (epoch 0).
	// If the starting epoch is different (e.g. because the node is restarting),
	// the corresponding state (including epoch number) must be loaded through applying Events read from the WAL.
	iss.initEpoch(0, maputil.GetSortedKeys(params.InitialMembership))

	// Return the initialized protocol module.
	return iss, nil
}

func InitialStateSnapshot(
	appState []byte,
	params *ModuleParams,
) *commonpb.StateSnapshot {

	memberships := make([]*commonpb.Membership, params.ConfigOffset+1)
	for i := 0; i < params.ConfigOffset+1; i++ {
		memberships[i] = t.MembershipPb(params.InitialMembership)
	}

	return &commonpb.StateSnapshot{
		AppData: appState,
		Configuration: &commonpb.EpochConfig{
			EpochNr:     0,
			Memberships: memberships,
		},
	}
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

	// TODO: Make event handlers return error as a second argument, instead of hard-coding nil here.

	switch e := event.Type.(type) {
	case *eventpb.Event_Init:
		return iss.applyInit()
	case *eventpb.Event_HashResult:
		return iss.applyHashResult(e.HashResult)
	case *eventpb.Event_SignResult:
		return iss.applySignResult(e.SignResult)
	case *eventpb.Event_NodeSigsVerified:
		return iss.applyNodeSigsVerified(e.NodeSigsVerified)
	case *eventpb.Event_AppSnapshot:
		return iss.applyAppSnapshot(e.AppSnapshot)
	case *eventpb.Event_NewConfig:
		return iss.applyNewConfig(e.NewConfig)
	case *eventpb.Event_Iss: // The ISS event type wraps all ISS-specific events.
		switch issEvent := e.Iss.Type.(type) {
		case *isspb.ISSEvent_Sb:
			return iss.applySBEvent(issEvent.Sb)
		case *isspb.ISSEvent_StableCheckpoint:
			return iss.applyStableCheckpoint(issEvent.StableCheckpoint)
		case *isspb.ISSEvent_PersistCheckpoint, *isspb.ISSEvent_PersistStableCheckpoint:
			// TODO: Ignoring WAL loading for the moment.
			return events.EmptyList(), nil
		case *isspb.ISSEvent_PushCheckpoint:
			return iss.applyPushCheckpoint()
		default:
			return nil, fmt.Errorf("unknown ISS event type: %T", issEvent)
		}

	case *eventpb.Event_Availability:
		switch avEvent := e.Availability.Type.(type) {
		case *availabilitypb.Event_NewCert:
			return iss.applyNewCert(avEvent.NewCert)
		default:
			return nil, fmt.Errorf("unknown availability event type: %T", avEvent)
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
func (iss *ISS) applyInit() (*events.EventList, error) {
	eventsOut := events.EmptyList()

	// Initialize the availability modules.
	for epochNr, configuration := range iss.configurations {
		eventsOut.PushBack(iss.initAvailabilityModule(t.EpochNr(epochNr), t.MembershipPb(configuration)))
	}

	// Trigger an Init event at all orderers.
	eventsOut.PushBackList(iss.initOrderers())

	return eventsOut, nil
}

// applyHashResult applies the HashResult event to the state of the ISS protocol state machine.
// A HashResult event means that a hash requested by this module has been computed by the Hasher module
// and is now ready to be processed.
func (iss *ISS) applyHashResult(result *eventpb.HashResult) (*events.EventList, error) {
	// We know already that the hash origin is the HashOrigin is of type ISS, since ISS produces no other types
	// and all HashResults with a different origin would not have been routed here.
	issOrigin := result.Origin.Type.(*eventpb.HashOrigin_Iss).Iss

	// TODO: Make event handlers return error as a second argument, instead of hard-coding nil here.

	// Further inspect the origin of the initial hash request and decide what to do with it.
	switch origin := issOrigin.Type.(type) {
	case *isspb.ISSHashOrigin_Sb:
		// If the original hash request has been produced by an SB instance,
		// create an appropriate SBEvent and apply it.
		// Note: Note that this is quite inefficient, allocating unnecessary boilerplate objects inside SBEvent().
		//       instead of calling iss.ApplyEvent(), one could directly call iss.applySBEvent()
		//       with a manually created isspb.SBEvent. The inefficient approach is chosen for code readability.
		epoch := t.EpochNr(origin.Sb.Epoch)
		instance := t.SBInstanceNr(origin.Sb.Instance)
		return iss.ApplyEvent(SBEvent(
			iss.moduleConfig.Self,
			epoch,
			instance,
			SBHashResultEvent(result.Digests, origin.Sb.Origin),
		))
	case *isspb.ISSHashOrigin_LogEntrySn:
		// Hash originates from delivering a CommitLogEntry.
		return iss.applyLogEntryHashResult(result.Digests[0], t.SeqNr(origin.LogEntrySn))
	case *isspb.ISSHashOrigin_StateSnapshotEpoch:
		// Hash originates from delivering an event of the application creating a state snapshot
		return iss.applyStateSnapshotHashResult(result.Digests[0], t.EpochNr(origin.StateSnapshotEpoch))
	default:
		return nil, fmt.Errorf("unknown origin of hash result: %T", origin)
	}
}

// applySignResult applies the SignResult event to the state of the ISS protocol state machine.
// A SignResult event means that a signature requested by this module has been computed by the Crypto module
// and is now ready to be processed.
func (iss *ISS) applySignResult(result *eventpb.SignResult) (*events.EventList, error) {
	// We know already that the SignOrigin is of type ISS, since ISS produces no other types
	// and all SignResults with a different origin would not have been routed here.
	issOrigin := result.Origin.Type.(*eventpb.SignOrigin_Iss).Iss

	// Further inspect the origin of the initial signing request and decide what to do with it.
	switch origin := issOrigin.Type.(type) {
	case *isspb.ISSSignOrigin_Sb:
		// If the original sign request has been produced by an SB instance,
		// create an appropriate SBEvent and apply it.
		// Note: Note that this is quite inefficient, allocating unnecessary boilerplate objects inside SBEvent().
		//       instead of calling iss.ApplyEvent(), one could directly call iss.applySBEvent()
		//       with a manually created isspb.SBEvent. The inefficient approach is chosen for code readability.
		epoch := t.EpochNr(origin.Sb.Epoch)
		instance := t.SBInstanceNr(origin.Sb.Instance)
		return iss.ApplyEvent(SBEvent(
			iss.moduleConfig.Self,
			epoch,
			instance,
			SBSignResultEvent(result.Signature, origin.Sb.Origin),
		))
	case *isspb.ISSSignOrigin_CheckpointEpoch:
		return iss.applyCheckpointSignResult(result.Signature, t.EpochNr(origin.CheckpointEpoch)), nil
	default:
		return nil, fmt.Errorf("unknown origin of sign result: %T", origin)
	}
}

// applyNodeSigVerified applies the NodeSigVerified event to the state of the ISS protocol state machine.
// Such an event means that a signature verification requested by this module has been completed by the Crypto module
// and is now ready to be processed.
func (iss *ISS) applyNodeSigsVerified(result *eventpb.NodeSigsVerified) (*events.EventList, error) {
	// We know already that the SigVerOrigin is of type ISS, since ISS produces no other types
	// and all NodeSigVerified events with a different origin would not have been routed here.
	issOrigin := result.Origin.Type.(*eventpb.SigVerOrigin_Iss).Iss

	// Further inspect the origin of the initial signature verification request and decide what to do with it.
	switch origin := issOrigin.Type.(type) {
	case *isspb.ISSSigVerOrigin_Sb:
		// If the original verification request has been produced by an SB instance,
		// create an appropriate SBEvent and apply it.
		// Note: Note that this is quite inefficient, allocating unnecessary boilerplate objects inside SBEvent().
		//       instead of calling iss.ApplyEvent(), one could directly call iss.applySBEvent()
		//       with a manually created isspb.SBEvent. The inefficient approach is chosen for code readability.
		epoch := t.EpochNr(origin.Sb.Epoch)
		instance := t.SBInstanceNr(origin.Sb.Instance)
		return iss.ApplyEvent(SBEvent(iss.moduleConfig.Self, epoch, instance, SBNodeSigsVerifiedEvent(
			result.Valid,
			result.Errors,
			t.NodeIDSlice(result.NodeIds),
			origin.Sb.Origin,
			result.AllOk,
		)))
	case *isspb.ISSSigVerOrigin_CheckpointEpoch:
		// A checkpoint only has one signature and thus each slice of the result only contains one element.
		return iss.applyCheckpointSigVerResult(
			result.Valid[0],
			result.Errors[0],
			t.NodeID(result.NodeIds[0]),
			t.EpochNr(origin.CheckpointEpoch),
		), nil
	case *isspb.ISSSigVerOrigin_StableCheckpoint:
		return iss.applyStableCheckpointSigVerResult(result.AllOk, origin.StableCheckpoint), nil
	default:
		return nil, fmt.Errorf("unknown origin of sign result: %T", origin)
	}
}

// applyAppSnapshot applies the event of the application creating a state snapshot.
// It passes the snapshot to the appropriate CheckpointTracker (identified by the event's associated epoch number).
func (iss *ISS) applyAppSnapshot(appSnapshot *eventpb.AppSnapshot) (*events.EventList, error) {

	//events.StateSnapshot(, events.EpochConfig(chat.currentEpoch, chat.memberships)),

	snapshotEpoch := iss.contextStore.RecoverAndDispose(contextstore.ItemID(appSnapshot.Origin.ItemID)).(t.EpochNr)

	if epoch, ok := iss.epochs[snapshotEpoch]; ok {
		return epoch.Checkpoint.ProcessStateSnapshot(events.StateSnapshot(
			appSnapshot.AppData,
			events.EpochConfig(epoch.Nr, iss.configurations[:int(snapshotEpoch)+iss.config.ConfigOffset+1]),
			// TODO: Make sure that all the necessary configurations have been received through the NewConfig event.
		)), nil
	}

	iss.logger.Log(logging.LevelWarn, "Snapshot epoch mismatch.",
		"curEpoch", iss.epoch.Nr, "snapshotEpoch", snapshotEpoch)
	return events.EmptyList(), nil
}

func (iss *ISS) applyNewConfig(config *eventpb.NewConfig) (*events.EventList, error) {
	eventsOut := events.EmptyList()

	// Convenience variable.
	newMembership := t.Membership(config.Membership)

	// Initialize the new availability module
	eventsOut.PushBack(iss.initAvailabilityModule(t.EpochNr(len(iss.configurations)), config.Membership))

	iss.logger.Log(logging.LevelDebug, "Adding configuration",
		"forEpoch", len(iss.configurations),
		"currentEpoch", iss.epoch.Nr,
		"newConfigNodes", maputil.GetSortedKeys(newMembership))

	// Save the new configuration.
	iss.configurations = append(iss.configurations, newMembership)

	// Advance to the next epoch if this configuration was the last missing bit.
	if iss.epochFinished() {
		eventsOut.PushBackList(iss.advanceEpoch())
	}

	return eventsOut, nil
}

// applyLogEntryHashResult applies the event of receiving the digest of a delivered CommitLogEntry.
// It attaches the digest to the entry and inserts the entry to the commit log.
// Based on the state of the commitLog, it may trigger delivering certificates to the application.
func (iss *ISS) applyLogEntryHashResult(digest []byte, logEntrySN t.SeqNr) (*events.EventList, error) {

	// Remove pending CommitLogEntry from the "waiting room"
	logEntry := iss.unhashedLogEntries[logEntrySN]
	delete(iss.unhashedLogEntries, logEntrySN)

	// Attach digest to entry.
	logEntry.Digest = digest

	// Insert the entry in the commitLog.
	iss.commitLog[logEntry.Sn] = logEntry

	// Deliver commitLog entries to the application in sequence number order.
	// This is relevant in the case when the sequence number of the currently SB-delivered certificate
	// is the first sequence number not yet delivered to the application.
	return iss.processCommitted()

}

// applyStateSnapshotHashResult applies the event of receiving the digest of a delivered event of the application creating a state snapshot.
// It passes the snapshot hash to the appropriate CheckpointTracker (identified by the event's associated epoch number).
func (iss *ISS) applyStateSnapshotHashResult(digest []byte, epoch t.EpochNr) (*events.EventList, error) {
	if iss.epoch.Nr != epoch {
		return events.EmptyList(), nil
	}
	return iss.epoch.Checkpoint.ProcessStateSnapshotHash(digest)
}

// applyCheckpointSignResult applies the event of receiving the Checkpoint message signature.
// It passes the signature to the appropriate CheckpointTracker (identified by the event's associated epoch number).
func (iss *ISS) applyCheckpointSignResult(signature []byte, epoch t.EpochNr) *events.EventList {
	if iss.epoch.Nr != epoch {
		return events.EmptyList()
	}
	return iss.epoch.Checkpoint.ProcessCheckpointSignResult(signature)
}

// It passes the signature verification result to the appropriate CheckpointTracker (identified by the event's associated epoch number).
func (iss *ISS) applyCheckpointSigVerResult(valid bool, err string, node t.NodeID, epoch t.EpochNr) *events.EventList {
	if iss.epoch.Nr != epoch {
		return events.EmptyList()
	}
	return iss.epoch.Checkpoint.ProcessSigVerified(valid, err, node)
}

// applySBEvent applies an event triggered by or addressed to an orderer (i.e., instance of Sequenced Broadcast),
// if that event belongs to the current epoch.
// TODO: Update this comment when the TODO below is addressed.
func (iss *ISS) applySBEvent(event *isspb.SBEvent) (*events.EventList, error) {

	// Ignore persist events (their handling is not yet implemented).
	// TODO: Deal with this when proper handlers are implemented.
	switch event.Event.Type.(type) {
	case *isspb.SBInstanceEvent_PbftPersistPreprepare,
		*isspb.SBInstanceEvent_PbftPersistPrepare,
		*isspb.SBInstanceEvent_PbftPersistCommit,
		*isspb.SBInstanceEvent_PbftPersistSignedViewChange,
		*isspb.SBInstanceEvent_PbftPersistNewView:
		return events.EmptyList(), nil
	}

	epochNr := t.EpochNr(event.Epoch)
	if epochNr > iss.epoch.Nr {
		// Events coming from future epochs should never occur (as, unlike messages, events are all generated locally.)
		return nil, fmt.Errorf("ISS event (type %T, instance %d) from future epoch: %d",
			event.Event.Type, event.Instance, event.Epoch)
	} else if epoch, ok := iss.epochs[epochNr]; ok {
		// If the event is from a known epoch, apply it in relation to the corresponding orderer instance.
		// Since the event is locally generated, we do not need to check whether the orderer exists in the epoch.
		return iss.applySBInstanceEvent(event.Event, epoch.Orderers[event.Instance]), nil
	} else {
		// If the event is from an old epoch, ignore it.
		// This might potentially happen if the epoch advanced while the event has been waiting in some buffer.
		iss.logger.Log(logging.LevelDebug, "Ignoring old event.",
			"epoch", epochNr, "type", fmt.Sprintf("%T", event.Event.Type))
		return events.EmptyList(), nil
	}
}

func (iss *ISS) applyStableCheckpoint(stableCheckpoint *isspb.StableCheckpoint) (*events.EventList, error) {
	eventsOut := events.EmptyList()

	if stableCheckpoint.Sn > iss.lastStableCheckpoint.Sn {
		// If this is the most recent checkpoint observed, save it.
		iss.logger.Log(logging.LevelInfo, "New stable checkpoint.",
			"epoch", stableCheckpoint.Snapshot.Configuration.EpochNr,
			"sn", stableCheckpoint.Sn,
			"replacingEpoch", iss.lastStableCheckpoint.Snapshot.Configuration.EpochNr,
			"replacingSn", iss.lastStableCheckpoint.Sn,
			"numMemberships", len(stableCheckpoint.Snapshot.Configuration.Memberships),
		)
		iss.lastStableCheckpoint = stableCheckpoint

		// Prune old entries from WAL, old periodic timers, and ISS state pertaining to old epochs.
		// The state to prune is determined according to the retention index
		// which is derived from the epoch number the new
		// stable checkpoint is associated with.
		pruneIndex := int(stableCheckpoint.Snapshot.Configuration.EpochNr) - iss.config.RetainedEpochs
		if pruneIndex > 0 { // "> 0" and not ">= 0", since only entries strictly smaller than the index are pruned.

			// Prune WAL and Timer
			eventsOut.PushBack(events.WALTruncate(iss.moduleConfig.Wal, t.RetentionIndex(pruneIndex)))
			eventsOut.PushBack(events.TimerGarbageCollect(iss.moduleConfig.Timer, t.RetentionIndex(pruneIndex)))

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
				t.TimeDuration(iss.config.CatchUpTimerPeriod),

				// Note that we are not using the current epoch number here, because it is not relevant for checkpoints.
				// Using pruneIndex makes sure that the re-transmission is stopped
				// on every stable checkpoint (when another one is started).
				t.RetentionIndex(pruneIndex),
			))

		}
	} else {
		iss.logger.Log(logging.LevelDebug, "Ignoring outdated stable checkpoint.", "sn", stableCheckpoint.Sn)
	}

	return eventsOut, nil
}

func (iss *ISS) applyPushCheckpoint() (*events.EventList, error) {

	// Send the latest stable checkpoint to potentially
	// delayed nodes. The set of nodes to send the latest
	// stable checkpoint to is determined based on the
	// highest epoch number in the messages received from
	// the node so far. If the highest epoch number we
	// have heard from the node is more than config.RetainedEpochs
	// behind, then that node is likely to be left behind
	// and needs the stable checkpoint in order to start
	// catching up with state transfer.
	var delayed []t.NodeID
	lastStableCheckpointEpoch := t.EpochNr(iss.lastStableCheckpoint.Snapshot.Configuration.EpochNr)
	for _, n := range maputil.GetKeys(iss.configurations[iss.epoch.Nr]) {
		if lastStableCheckpointEpoch > iss.nodeEpochMap[n]+t.EpochNr(iss.config.RetainedEpochs) {
			delayed = append(delayed, n)
		}
	}

	iss.logger.Log(logging.LevelDebug, "Pushing state to nodes.",
		"delayed", delayed, "numNodes", len(iss.configurations[iss.epoch.Nr]), "nodeEpochMap", iss.nodeEpochMap)

	m := StableCheckpointMessage(iss.lastStableCheckpoint)
	return events.ListOf(events.SendMessage(iss.moduleConfig.Net, m, delayed)), nil
}

// applyMessageReceived applies a message received over the network.
// Note that this is not the only place messages are applied.
// Messages received "ahead of time" that have been buffered are applied in applyBufferedMessages.
func (iss *ISS) applyMessageReceived(messageReceived *eventpb.MessageReceived) (*events.EventList, error) {

	// Convenience variables used for readability.
	message := messageReceived.Msg
	from := t.NodeID(messageReceived.From)

	// ISS only accepts ISS messages. If another message is applied, the next line panics.
	switch msg := message.Type.(*messagepb.Message_Iss).Iss.Type.(type) {
	case *isspb.ISSMessage_Checkpoint:
		return iss.applyCheckpointMessage(msg.Checkpoint, from), nil
	case *isspb.ISSMessage_StableCheckpoint:
		return iss.applyStableCheckpointMessage(msg.StableCheckpoint, from)
	case *isspb.ISSMessage_Sb:
		return iss.applySBMessage(msg.Sb, from), nil
	default:
		return nil, fmt.Errorf("unknown ISS message type: %T", msg)
	}
}

// applyCheckpointMessage relays a Checkpoint message received over the network to the appropriate CheckpointTracker.
func (iss *ISS) applyCheckpointMessage(message *isspb.Checkpoint, source t.NodeID) *events.EventList {

	// Remember the highest epoch number for each node to detect
	// later if the remote node is delayed too much and requires
	// assistance in order to catch up through state transfer.
	epochNr := t.EpochNr(message.Epoch)
	if iss.nodeEpochMap[source] < epochNr {
		iss.nodeEpochMap[source] = epochNr
	}

	if epochNr > iss.epoch.Nr {
		// If the message is for a future epoch,
		// it might have been sent by a node that already transitioned to a newer epoch,
		// but this node is slightly behind (still in an older epoch) and cannot process the message yet.
		// In such case, save the message in a backlog (if there is buffer space) for later processing.
		iss.messageBuffers[source].Store(message)

		return events.EmptyList()
	}

	// If the message is for a known epoch, check its validity and
	// apply it to the corresponding checkpoint tracker instance.
	if epoch, ok := iss.epochs[epochNr]; ok {
		return epoch.Checkpoint.applyMessage(message, source)
	}

	// Otherwise, ignore message (in this case it must be from a garbage-collectd epoch).
	return events.EmptyList()
}

// applyStableCheckpointMessage processes a received StableCheckpoint message
// by creating a request for verifying the signatures in the included checkpoint certificate.
// The actual processing then happens in applyStableCheckpointSigVerResult.
func (iss *ISS) applyStableCheckpointMessage(chkp *isspb.StableCheckpoint, _ t.NodeID) (*events.EventList, error) {

	// Extract signatures and the signing node IDs from the received message.
	// TODO: Using underlying protobuf type for node ID explicitly here (nodeID string).
	//       Modify the code to only use the abstract type.
	nodeIDs := make([]t.NodeID, 0)
	signatures := make([][]byte, 0)
	maputil.IterateSorted(chkp.Cert, func(nodeID string, signature []byte) bool {
		nodeIDs = append(nodeIDs, t.NodeID(nodeID))
		signatures = append(signatures, signature)
		return true
	})

	// Request verification of signatures in the checkpoint certificate
	return events.ListOf(events.VerifyNodeSigs(
		"crypto",

		// TODO: !!!! This is wrong !!!!
		//       The snapshot first has to be hashed and only then the signatures can be verified.
		//       Using dummy value []byte{0} for state snapshot hash for now and skipping verification.

		[][][]byte{
			serializing.CheckpointForSig(t.EpochNr(chkp.Snapshot.Configuration.EpochNr), t.SeqNr(chkp.Sn), []byte{0}),
		},
		signatures,
		nodeIDs,
		StableCheckpointSigVerOrigin(iss.moduleConfig.Self, chkp),
	)), nil
}

// applyStableCheckpointSigVerResult applies a StableCheckpoint message
// the signature of which signature has been verified.
// It checks the message and decides whether to install the state snapshot from the message.
func (iss *ISS) applyStableCheckpointSigVerResult(signaturesOK bool, chkp *isspb.StableCheckpoint) *events.EventList {
	eventsOut := events.EmptyList()

	// TODO: !!!! This is wrong !!!!
	//       The snapshot first has to be hashed and only then the signatures can be verified.
	//       Using dummy value []byte{0} for state snapshot hash for now and skipping verification.

	//// Ignore checkpoint with in valid or insufficiently many signatures.
	//// TODO: Make sure this code still works when reconfiguration is implemented.
	//if !signaturesOK || len(chkp.Cert) < weakQuorum(len(iss.config.Membership)) {
	//	iss.logger.Log(logging.LevelWarn, "Ignoring invalid stable checkpoint message.", "epoch", chkp.Epoch)
	//	return eventsOut
	//}

	// Convenience variable
	chkpEpoch := t.EpochNr(chkp.Snapshot.Configuration.EpochNr)

	// Check how far is the received stable checkpoint ahead of
	// the local node's state.
	if chkpEpoch <= iss.epoch.Nr+1 {
		// Ignore stable checkpoints that are not far enough
		// ahead of the current state of the local node.
		return events.EmptyList()
	}

	iss.logger.Log(logging.LevelDebug, "Installing state snapshot.", "epoch", chkpEpoch)

	// Clean up global ISS state that belongs to the current epoch
	// instance that local replica got stuck with.
	iss.epochs = make(map[t.EpochNr]*epochInfo)
	// iss.epoch = nil // This will be overwritten by initEpoch anyway.
	iss.commitLog = make(map[t.SeqNr]*CommitLogEntry)
	iss.unhashedLogEntries = make(map[t.SeqNr]*CommitLogEntry)
	iss.nextDeliveredSN = t.SeqNr(chkp.Sn)
	iss.newEpochSN = iss.nextDeliveredSN

	// Initialize a new ISS epoch instance for the new stable
	// checkpoint to continue participating in the protocol
	// starting with that epoch after installing the state
	// snapshot from the new stable checkpoint.

	// Save the configurations obtained in the checkpoint
	iss.configurations = make([]map[t.NodeID]t.NodeAddress, len(chkp.Snapshot.Configuration.Memberships))
	for i, m := range chkp.Snapshot.Configuration.Memberships {
		iss.configurations[i] = t.Membership(m)
		eventsOut.PushBack(iss.initAvailabilityModule(t.EpochNr(i), m))
	}

	// TODO: Check if all the configurations are present in the checkpoint.
	// TODO: Properly serialize and deserialize the leader selection policy and pass it here.
	iss.initEpoch(chkpEpoch, maputil.GetSortedKeys(iss.configurations[chkpEpoch]))

	// Update the last stable checkpoint stored in the global ISS structure.
	iss.lastStableCheckpoint = chkp

	// Create an event to request the application module for
	// restoring its state from the snapshot received in the new
	// stable checkpoint message.
	eventsOut.PushBack(events.AppRestoreState(iss.moduleConfig.App, chkp.Snapshot))

	// Activate SB instances of the new epoch which will deliver
	// availability certificates after the application module has restored the state
	// from the snapshot.
	eventsOut.PushBackList(iss.initOrderers())

	// Apply any message buffered for the new epoch and append any
	// emitted event to the returns event list.
	eventsOut.PushBackList(iss.applyBufferedMessages())

	return eventsOut
}

func (iss *ISS) initAvailabilityModule(epochNr t.EpochNr, membership *commonpb.Membership) *eventpb.Event {
	return factorymodule.FactoryNewModule(
		iss.moduleConfig.Availability,
		iss.moduleConfig.Availability.Then(t.ModuleID(fmt.Sprintf("%v", epochNr))),
		t.RetentionIndex(epochNr),
		&factorymodulepb.GeneratorParams{Type: &factorymodulepb.GeneratorParams_MultisigCollector{
			MultisigCollector: &mscpb.InstanceParams{Membership: membership},
		}},
	)
}

// applySBMessage applies a message destined for an orderer (i.e. a Sequenced Broadcast implementation).
func (iss *ISS) applySBMessage(message *isspb.SBMessage, from t.NodeID) *events.EventList {

	epochNr := t.EpochNr(message.Epoch)
	if epochNr > iss.epoch.Nr {
		// If the message is for a future epoch,
		// it might have been sent by a node that already transitioned to a newer epoch,
		// but this node is slightly behind (still in an older epoch) and cannot process the message yet.
		// In such case, save the message in a backlog (if there is buffer space) for later processing.
		iss.messageBuffers[from].Store(message)
		return events.EmptyList()
	} else if epoch, ok := iss.epochs[epochNr]; ok {
		// If the message is for the current epoch, check its validity and
		// apply it to the corresponding orderer in form of an SBMessageReceived event.
		if err := epoch.validateSBMessage(message, from); err != nil {
			iss.logger.Log(logging.LevelWarn, "Ignoring invalid SB message.",
				"type", fmt.Sprintf("%T", message.Msg.Type), "from", from, "error", err, "epoch", epochNr)
			return events.EmptyList()
		}

		return iss.applySBInstanceEvent(SBMessageReceivedEvent(message.Msg, from), epoch.Orderers[message.Instance])
	} else {
		// Ignore old messages
		iss.logger.Log(logging.LevelDebug, "Ignoring message from old epoch.",
			"from", from, "epoch", epochNr, "type", fmt.Sprintf("%T", message.Msg.Type))
		return events.EmptyList()
	}
}

// ============================================================
// Additional protocol logic
// ============================================================

// initEpoch initializes a new ISS epoch with the given epoch number.
func (iss *ISS) initEpoch(newEpoch t.EpochNr, membership []t.NodeID) {

	var leaderPolicy LeaderSelectionPolicy
	if newEpoch == 0 {
		leaderPolicy = iss.config.LeaderPolicy
	} else {
		leaderPolicy = iss.epoch.leaderPolicy.Reconfigure(membership)
	}

	iss.logger.Log(logging.LevelInfo, "Initializing new epoch", "epochNr", newEpoch, "numNodes", len(membership))

	// Set the new epoch number and re-initialize list of orderers.
	epoch := newEpochInfo(
		newEpoch,
		membership,
		newCheckpointTracker(
			iss.moduleConfig,
			iss.ownID,
			iss.nextDeliveredSN,
			newEpoch,
			t.TimeDuration(iss.config.CheckpointResendPeriod),
			logging.Decorate(iss.logger, "CT: ", "epoch", newEpoch),
		),
		leaderPolicy,
	)

	iss.epochs[newEpoch] = &epoch
	iss.epoch = &epoch

	// TODO: DIRTY TRICK to get some capacity of any message buffer and copy it
	//       Implement proper adapting of buffer capacities when changing the number of message buffers.
	var mbCapacity int
	b, ok := maputil.Any(iss.messageBuffers)
	if !ok {
		mbCapacity = iss.config.MsgBufCapacity
	} else {
		mbCapacity = b.Capacity()
	}

	// Add new message buffers if not already present.
	// TODO: Implement garbage collection of old ones.
	//       Also, this is very dirty code. Fix!
	for _, nodeID := range membership {
		if _, ok := iss.messageBuffers[nodeID]; nodeID != iss.ownID && !ok {
			iss.messageBuffers[nodeID] = messagebuffer.New(
				nodeID,
				mbCapacity,
				logging.Decorate(logging.Decorate(iss.logger, "Msgbuf: "), "", "source", nodeID),
			)
		}
	}

	// Compute the set of leaders for the new epoch.
	// Note that leader policy is stateful, choosing leaders deterministically based on the state of the system.
	// Its state must be consistent across all nodes when calling Leaders() on it.
	leaders := iss.epoch.leaderPolicy.Leaders(newEpoch)

	// Create new segments of the commit log, one per leader selected by the leader selection policy.
	// Instantiate one orderer (SB instance) for each segment.
	for i, leader := range leaders {

		// Create segment.
		seg := &segment{
			Leader:     leader,
			Membership: membership,
			SeqNrs: sequenceNumbers(
				iss.nextDeliveredSN+t.SeqNr(i),
				t.SeqNr(len(leaders)),
				iss.config.SegmentLength),
		}
		iss.newEpochSN += t.SeqNr(len(seg.SeqNrs))

		// Instantiate a new PBFT orderer.
		// TODO: When more protocols are implemented, make this configurable, so other orderer types can be chosen.
		sbInst := newPbftInstance(
			iss.ownID,
			seg,
			newPBFTConfig(
				iss.config,
				membership,
				iss.moduleConfig.Availability.Then(t.ModuleID(fmt.Sprintf("%v", epoch.Nr))),
			),
			&sbEventService{moduleConfig: iss.moduleConfig, epoch: newEpoch, instance: t.SBInstanceNr(i)},
			logging.Decorate(iss.logger, "PBFT: ", "epoch", newEpoch, "instance", i))

		// Add the orderer to the list of orderers.
		iss.epoch.Orderers = append(iss.epoch.Orderers, sbInst)

	}
}

// initOrderers sends the SBInit event to all orderers in the current epoch.
func (iss *ISS) initOrderers() *events.EventList {
	eventsOut := events.EmptyList()

	sbInit := SBInitEvent()
	for _, orderer := range iss.epoch.Orderers {
		eventsOut.PushBackList(orderer.ApplyEvent(sbInit))
	}

	return eventsOut
}

// epochFinished returns true when all the sequence numbers of the current epochs have been committed, otherwise false.
func (iss *ISS) epochFinished() bool {
	return iss.nextDeliveredSN == iss.newEpochSN
}

// processCommitted delivers entries from the commitLog in order of their sequence numbers.
// Whenever a new entry is inserted in the commitLog, this function must be called
// to create Deliver events for all the certificates that can be delivered to the application.
// processCommitted also triggers other internal Events like epoch transitions and state checkpointing.
func (iss *ISS) processCommitted() (*events.EventList, error) {
	eventsOut := events.EmptyList()

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

		// Create a new Deliver event.
		eventsOut.PushBack(events.Deliver(iss.moduleConfig.App, iss.nextDeliveredSN, &cert))

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
		eventsOut.PushBackList(iss.advanceEpoch())
	}

	return eventsOut, nil
}

func (iss *ISS) advanceEpoch() *events.EventList {
	eventsOut := events.EmptyList()

	// Convenience variables
	oldEpochNr := iss.epoch.Nr
	newEpochNr := iss.epoch.Nr + 1

	iss.logger.Log(logging.LevelDebug, "Advancing epoch.",
		"epoch", oldEpochNr,
		"nextSN", iss.nextDeliveredSN,
		"numConfigs", len(iss.configurations),
	)

	if !(len(iss.configurations) > int(newEpochNr)) {
		iss.logger.Log(logging.LevelWarn, "Cannot advance to epoch yet. Waiting for configuration.",
			"newEpoch", newEpochNr)
		return eventsOut
	}

	// Initialize the internal data structures for the new epoch.
	iss.initEpoch(newEpochNr, maputil.GetSortedKeys(iss.configurations[newEpochNr]))

	// Signal the new epoch to the application.
	// This must happen before starting the checkpoint protocol, since the application
	// must already be in the new epoch when processing the state snapshot request
	// emitted by the checkpoint sub-protocol.
	eventsOut.PushBack(events.NewEpoch(iss.moduleConfig.App, newEpochNr))

	// Look up a (or create a new) checkpoint tracker and start the checkpointing protocol.
	// This must happen after initialization of the new epoch,
	// as the sequence number the checkpoint will be associated with (iss.nextDeliveredSN)
	// is already part of the new epoch.
	// The checkpoint tracker might already exist if a corresponding message has been already received.
	// iss.nextDeliveredSN is the first sequence number *not* included in the checkpoint,
	// i.e., as sequence numbers start at 0, the checkpoint includes the first iss.nextDeliveredSN sequence numbers.
	// The membership used for the checkpoint tracker still must be the old membership.
	eventsOut.PushBackList(iss.epoch.Checkpoint.Start(
		maputil.GetSortedKeys(iss.configurations[oldEpochNr]),
		iss.contextStore.Store(newEpochNr),
	))

	// Give the init signals to the newly instantiated orderers.
	eventsOut.PushBackList(iss.initOrderers())

	// Process backlog of buffered SB messages.
	eventsOut.PushBackList(iss.applyBufferedMessages())

	return eventsOut
}

// applyBufferedMessages applies all SB messages destined to the current epoch
// that have been buffered during past epochs.
// This function is always called directly after initializing a new epoch, except for epoch 0.
func (iss *ISS) applyBufferedMessages() *events.EventList {
	eventsOut := events.EmptyList()

	// Iterate over the all messages in all buffers, selecting those that can be applied.
	for _, buffer := range iss.messageBuffers {
		buffer.Iterate(iss.bufferedMessageFilter, func(source t.NodeID, msg proto.Message) {

			// Apply all messages selected by the filter.
			switch m := msg.(type) {
			case *isspb.SBMessage:
				eventsOut.PushBackList(iss.applySBMessage(m, source))
			case *isspb.Checkpoint:
				eventsOut.PushBackList(iss.applyCheckpointMessage(m, source))
			}

		})
	}

	return eventsOut
}

// bufferedMessageFilter decides, given a message, whether it is appropriate to apply the message, discard it,
// or keep it in the buffer, returning the appropriate value of type messagebuffer.Applicable.
func (iss *ISS) bufferedMessageFilter(_ t.NodeID, message proto.Message) messagebuffer.Applicable {
	switch msg := message.(type) {
	case *isspb.SBMessage:
		// For SB messages, filter based on the epoch number of the message.
		switch e := t.EpochNr(msg.Epoch); {
		case e < iss.epoch.Nr:
			return messagebuffer.Past
		case e == iss.epoch.Nr:
			return messagebuffer.Current
		default: // e > iss.epoch
			return messagebuffer.Future
		}
	case *isspb.Checkpoint:
		// For Checkpoint messages, messages with current or older epoch number are considered,
		// if they correspond to a more recent checkpoint than the last stable checkpoint stored locally at this node.
		switch {
		case msg.Sn <= iss.lastStableCheckpoint.Sn:
			return messagebuffer.Past
		case t.EpochNr(msg.Epoch) <= iss.epoch.Nr:
			return messagebuffer.Current
		default: // message from future epoch
			return messagebuffer.Future
		}
	default:
		panic(fmt.Errorf("cannot extract epoch from message type: %T", message))
	}
	// Note that validation is not performed here, as it is performed anyway when applying the message.
	// Thus, the messagebuffer.Invalid option is not used.
}

// ============================================================
// Auxiliary functions
// ============================================================

// sequenceNumbers returns a list of sequence numbers of length `length`,
// starting with sequence number `start`, with the difference between two consecutive sequence number being `step`.
// This function is used to compute the sequence numbers of a segment.
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

// membershipSet takes a list of node IDs and returns a map of empty structs with an entry for each node ID in the list.
// The returned map is effectively a set representation of the given list,
// useful for testing whether any given node ID is in the set.
func membershipSet(membership []t.NodeID) map[t.NodeID]struct{} {

	// Allocate a new map representing a set of node IDs
	set := make(map[t.NodeID]struct{})

	// Add an empty struct for each node ID in the list.
	for _, nodeID := range membership {
		set[nodeID] = struct{}{}
	}

	// Return the resulting set of node IDs.
	return set
}

// Returns a configuration of a new PBFT instance based on the current ISS configuration.
func newPBFTConfig(issParams *ModuleParams, membership []t.NodeID, availabilityModuleID t.ModuleID) *PBFTConfig {

	// Return a new PBFT configuration with selected values from the ISS configuration.
	return &PBFTConfig{
		Membership:               membership,
		AvailabilityModuleID:     availabilityModuleID,
		MaxProposeDelay:          issParams.MaxProposeDelay,
		MsgBufCapacity:           issParams.MsgBufCapacity,
		DoneResendPeriod:         issParams.PBFTDoneResendPeriod,
		CatchUpDelay:             issParams.PBFTCatchUpDelay,
		ViewChangeSNTimeout:      issParams.PBFTViewChangeSNTimeout,
		ViewChangeSegmentTimeout: issParams.PBFTViewChangeSegmentTimeout,
		ViewChangeResendPeriod:   issParams.PBFTViewChangeResendPeriod,
	}
}

// removeNodeID emoves a node ID from a list of node IDs.
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

func maxFaulty(n int) int {
	// assuming n > 3f:
	//   return max f
	return (n - 1) / 3
}

func strongQuorum(n int) int {
	// assuming n > 3f:
	//   return min q: 2q > n+f
	f := maxFaulty(n)
	return (n+f)/2 + 1
}

func weakQuorum(n int) int {
	// assuming n > 3f:
	//   return min q: q > f
	return maxFaulty(n) + 1
}
