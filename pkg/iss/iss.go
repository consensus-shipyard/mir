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
	"sort"

	"golang.org/x/exp/constraints"

	"google.golang.org/protobuf/proto"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/messagebuffer"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/isspb"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	"github.com/filecoin-project/mir/pkg/pb/statuspb"
	"github.com/filecoin-project/mir/pkg/serializing"
	t "github.com/filecoin-project/mir/pkg/types"
)

// ============================================================
// Auxiliary types
// ============================================================

// The missingRequestInfo type represents information about the requests that the node is waiting to receive
// before accepting a proposal for a sequence number.
// A request counts as "received" by the node when the RequestReady event for that request is applied.
// For example, with the PBFT orderer, after the PBFT implementation of some SB instance receives a preprepare message,
// it will wait until ISS confirms that all requests referenced in the preprepare message have been received.
type missingRequestInfo struct {

	// A globally unique identifier of this struct. It consists of the orderer ID
	// and the orderer's internal reference of the proposal for which requests are missing.
	// When notifying the orderer about the received requests, this reference is passed back to the orderer,
	// so it can pair the notification with the corresponding SBWaitForRequests Event.
	Ref missingRequestInfoRef

	// Set of all missing requests, represented as map of request references
	// indexed by the string representation of those request reference.
	Requests map[string]*requestpb.RequestRef

	// SB instance to be notified (via the RequestsReady event) when all missing requests have been received.
	Orderer sbInstance

	// TODO: implement NAcks after a timeout, using iss.demandRequestRetransmission(missingRequests)
}

// missingRequestInfoRef uniquely identifies a missingRequestInfo across all orderer instances
// (the contained Ref field is only guaranteed to be unique within one SB instance, but not globally).
// The separate type is necessary, as it is used as a map key.
type missingRequestInfoRef struct {

	// The orderer that is waiting for missing requests
	Orderer t.SBInstanceNr

	// Orderer's (i.e., SB instance's) internal reference to be passed back to the orderer
	// when announcing that the missing requests have been received.
	SBRef *isspb.SBReqWaitReference
}

// The segment type represents an ISS segment.
// It is use to parametrize an orderer (i.e. the SB instance).
type segment struct {

	// The leader node of the orderer.
	Leader t.NodeID

	// List of all nodes executing the orderer implementation.
	Membership []t.NodeID

	// List of sequence numbers for which the orderer is responsible.
	// This is the actual "segment" of the commit log.
	SeqNrs []t.SeqNr

	// List of IDs of buckets from which the orderer will draw the requests it proposes.
	BucketIDs []int
}

// epochInfo holds epoch-specific information that becomes irrelevant on advancing to the next epoch.
type epochInfo struct {

	// Epoch number.
	Nr t.EpochNr

	// Orderers associated with the epoch.
	Orderers []sbInstance

	// Checkpoint sub-protocol state.
	Checkpoint *checkpointTracker
}

// The CommitLogEntry type represents an entry of the commit log, the final output of the ordering process.
// Whenever an orderer delivers a batch (or a special abort value),
// it is inserted to the commit log in form of a commitLogEntry.
type CommitLogEntry struct {
	// Sequence number at which this entry has been ordered.
	Sn t.SeqNr

	// The delivered request batch.
	Batch *requestpb.Batch

	// The digest (hash) of the batch.
	Digest []byte

	// A flag indicating whether this entry is an actual request batch (false)
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

	// The ID of the node executing this instance of the protocol.
	ownID t.NodeID

	// The buckets holding incoming requests.
	// In each epoch, these buckets are re-distributed to the epoch's orderers,
	// each ordering one segment of the commit log.
	buckets *bucketGroup

	// Logger the ISS implementation uses to output log messages.
	// This is mostly for debugging - not to be confused with the commit log.
	logger logging.Logger

	// --------------------------------------------------------------------------------
	// These fields might change from epoch to epoch. Modified only by initEpoch()
	// --------------------------------------------------------------------------------

	// The ISS configuration parameters (e.g. number of buckets, batch size, etc...)
	// passed to New() when creating an ISS protocol instance.
	// TODO: Make it possible to change this dynamically.
	config *Config

	// The current epoch instance.
	epoch *epochInfo

	// Epoch instances.
	epochs map[t.EpochNr]*epochInfo

	// Highest epoch numbers indicated in Checkpoint messages from each node.
	nodeEpochMap map[t.NodeID]t.EpochNr

	// Index of orderers based on the buckets they are assigned.
	// For each bucket ID, this map stores the orderer to which the bucket is assigned in the current epoch.
	bucketOrderers map[int]sbInstance

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

	// The final log of committed batches.
	// For each sequence number, it holds the committed batch (or the special abort value).
	// Each Deliver event of an orderer translates to inserting an entry in the commitLog.
	// This, in turn, leads to delivering the batch to the application,
	// as soon as all entries with lower sequence numbers have been delivered.
	// I.e., the entries are not necessarily inserted in order of their sequence numbers,
	// but they are delivered to the application in that order.
	commitLog map[t.SeqNr]*CommitLogEntry

	// CommitLogEntries for which the hash has been requested, but not yet computed.
	// When an orderer delivers a Batch, ISS creates a CommitLogEntry with a nil Hash, stores it here,
	// and creates a HashRequest. When the HashResult arrives, ISS removes the CommitLogEntry from unhashedLogEntries,
	// fills in the missing hash, and inserts it to the commitLog.
	unhashedLogEntries map[t.SeqNr]*CommitLogEntry

	// The first undelivered sequence number in the commitLog.
	// This field drives the in-order delivery of the log entries to the application.
	nextDeliveredSN t.SeqNr

	// The first sequence number to be delivered in the new epoch.
	newEpochSN t.SeqNr

	// This field holds information about the requests that the node is waiting to receive
	// before accepting a proposal for that sequence number.
	// Each proposal has a unique reference *within its SB instance*.
	// Thus, the map is indexed a tuple of SB instance ID and that instance's reference
	// (packed in a missingRequestInfoRef struct).
	// A request counts as "received" by the node when the RequestReady event for that request is applied.
	// As soon as all requests for a proposal have been received, the corresponding entry is deleted from this map.
	missingRequests map[missingRequestInfoRef]*missingRequestInfo

	// For each request that is missing
	// (i.e. for which the proposal has been received by some orderer, but the request itself has not yet arrived),
	// this field holds a pointer to the object tracking missing requests for the whole proposal.
	// As soon as a request has been received the corresponding entry is deleted from this map.
	missingRequestIndex map[string]*missingRequestInfo

	// Stores the stable checkpoint with the highest sequence number observed so far.
	// If no stable checkpoint has been observed yet, lastStableCheckpoint is initialized to a stable checkpoint value
	// corresponding to the initial state and associated with sequence number 0.
	lastStableCheckpoint *isspb.StableCheckpoint
}

// New returns a new initialized instance of the ISS protocol module to be used when instantiating a mir.Node.
// Arguments:
// - ownID:  the ID of the node being instantiated with ISS.
// - config: ISS protocol-specific configuration (e.g. number of buckets, batch size, etc...).
//           see the documentation of the Config type for details.
// - logger: Logger the ISS implementation uses to output log messages.
func New(ownID t.NodeID, config *Config, logger logging.Logger) (*ISS, error) {

	// Check whether the passed configuration is valid.
	if err := CheckConfig(config); err != nil {
		return nil, fmt.Errorf("invalid ISS configuration: %w", err)
	}

	// Initialize a new ISS object.
	iss := &ISS{
		// Static fields
		ownID:   ownID,
		buckets: newBuckets(config.NumBuckets, logger),
		logger:  logger,

		// Fields modified only by initEpoch
		config:         config,
		epochs:         make(map[t.EpochNr]*epochInfo),
		nodeEpochMap:   make(map[t.NodeID]t.EpochNr),
		bucketOrderers: nil,

		// Fields modified throughout an epoch
		commitLog:           make(map[t.SeqNr]*CommitLogEntry),
		unhashedLogEntries:  make(map[t.SeqNr]*CommitLogEntry),
		nextDeliveredSN:     0,
		newEpochSN:          0,
		missingRequests:     nil, // allocated in initEpoch()
		missingRequestIndex: nil, // allocated in initEpoch()
		messageBuffers: messagebuffer.NewBuffers(
			removeNodeID(config.Membership, ownID), // Create a message buffer for everyone except for myself.
			config.MsgBufCapacity,
			logging.Decorate(logger, "Msgbuf: "),
		),
		lastStableCheckpoint: &isspb.StableCheckpoint{
			Epoch: 0,
			Sn:    0,
			// TODO: When the storing of actual application state is implemented, some encoding of "initial state"
			//       will have to be set here. E.g., an empty byte slice could be defined as "initial state" and
			//       the application required to interpret it as such.
		},
	}

	// Initialize the first epoch (epoch 0).
	// If the starting epoch is different (e.g. because the node is restarting),
	// the corresponding state (including epoch number) must be loaded through applying Events read from the WAL.
	iss.initEpoch(0)

	// Return the initialized protocol module.
	return iss, nil
}

// ============================================================
// Protocol Interface implementation
// ============================================================

// ApplyEvent receives one event and applies it to the ISS protocol state machine, potentially altering its state
// and producing a (potentially empty) list of more events to be applied to other modules.
func (iss *ISS) ApplyEvent(event *eventpb.Event) *events.EventList {
	switch e := event.Type.(type) {
	case *eventpb.Event_Init:
		return iss.applyInit(e.Init)
	case *eventpb.Event_HashResult:
		return iss.applyHashResult(e.HashResult)
	case *eventpb.Event_SignResult:
		return iss.applySignResult(e.SignResult)
	case *eventpb.Event_NodeSigsVerified:
		return iss.applyNodeSigsVerified(e.NodeSigsVerified)
	case *eventpb.Event_RequestReady:
		return iss.applyRequestReady(e.RequestReady)
	case *eventpb.Event_AppSnapshot:
		return iss.applyAppSnapshot(e.AppSnapshot)
	case *eventpb.Event_Iss: // The ISS event type wraps all ISS-specific events.
		switch issEvent := e.Iss.Type.(type) {
		case *isspb.ISSEvent_Sb:
			return iss.applySBEvent(issEvent.Sb)
		case *isspb.ISSEvent_StableCheckpoint:
			return iss.applyStableCheckpoint(issEvent.StableCheckpoint)
		case *isspb.ISSEvent_PersistCheckpoint, *isspb.ISSEvent_PersistStableCheckpoint:
			// TODO: Ignoring WAL loading for the moment.
			return &events.EventList{}
		default:
			panic(fmt.Sprintf("unknown ISS event type: %T", issEvent))
		}
	case *eventpb.Event_MessageReceived:
		return iss.applyMessageReceived(e.MessageReceived)
	default:
		panic(fmt.Sprintf("unknown protocol (ISS) event type: %T", event.Type))
	}
}

// Status returns a protobuf representation of the current protocol state that can be later printed (TODO: Say how).
// This functionality is meant mostly for debugging and is *not* meant to provide an interface for
// serializing and deserializing the whole protocol state.
func (iss *ISS) Status() (s *statuspb.ProtocolStatus, err error) {
	// TODO: Implement this.
	return nil, nil
}

// ============================================================
// Event application
// ============================================================

// applyInit initializes the ISS protocol.
// This event is only expected to be applied once at startup,
// after all the events stored in the WAL have been applied and before any other event has been applied.
func (iss *ISS) applyInit(init *eventpb.Init) *events.EventList {

	// Trigger an Init event at all orderers.
	return iss.initOrderers()
}

// applyHashResult applies the HashResult event to the state of the ISS protocol state machine.
// A HashResult event means that a hash requested by this module has been computed by the Hasher module
// and is now ready to be processed.
func (iss *ISS) applyHashResult(result *eventpb.HashResult) *events.EventList {
	// We know already that the hash origin is the HashOrigin is of type ISS, since ISS produces no other types
	// and all HashResults with a different origin would not have been routed here.
	issOrigin := result.Origin.Type.(*eventpb.HashOrigin_Iss).Iss

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
		return iss.ApplyEvent(SBEvent(epoch, instance, SBHashResultEvent(result.Digests, origin.Sb.Origin)))
	case *isspb.ISSHashOrigin_LogEntrySn:
		// Hash originates from delivering a CommitLogEntry.
		return iss.applyLogEntryHashResult(result.Digests[0], t.SeqNr(origin.LogEntrySn))
	case *isspb.ISSHashOrigin_AppSnapshotEpoch:
		// Hash originates from delivering an event of the application creating a state snapshot
		return iss.applyAppSnapshotHashResult(result.Digests[0], t.EpochNr(origin.AppSnapshotEpoch))
	default:
		panic(fmt.Sprintf("unknown origin of hash result: %T", origin))
	}
}

// applySignResult applies the SignResult event to the state of the ISS protocol state machine.
// A SignResult event means that a signature requested by this module has been computed by the Crypto module
// and is now ready to be processed.
func (iss *ISS) applySignResult(result *eventpb.SignResult) *events.EventList {
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
		return iss.ApplyEvent(SBEvent(epoch, instance, SBSignResultEvent(result.Signature, origin.Sb.Origin)))
	case *isspb.ISSSignOrigin_CheckpointEpoch:
		return iss.applyCheckpointSignResult(result.Signature, t.EpochNr(origin.CheckpointEpoch))
	default:
		panic(fmt.Sprintf("unknown origin of sign result: %T", origin))
	}
}

// applyNodeSigVerified applies the NodeSigVerified event to the state of the ISS protocol state machine.
// Such an event means that a signature verification requested by this module has been completed by the Crypto module
// and is now ready to be processed.
func (iss *ISS) applyNodeSigsVerified(result *eventpb.NodeSigsVerified) *events.EventList {
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
		return iss.ApplyEvent(SBEvent(epoch, instance, SBNodeSigsVerifiedEvent(
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
		)
	default:
		panic(fmt.Sprintf("unknown origin of sign result: %T", origin))
	}
}

// applyRequestReady applies the RequestReady event to the state of the ISS protocol state machine.
// A RequestReady event means that a request is considered valid and authentic by the node and can be processed.
func (iss *ISS) applyRequestReady(requestReady *eventpb.RequestReady) *events.EventList {
	eventsOut := &events.EventList{}

	// Get request reference.
	ref := requestReady.RequestRef

	// Get bucket to which the new request maps.
	bucket := iss.buckets.RequestBucket(ref)

	// Add request to its bucket if it has not been added yet.
	if !bucket.Add(ref) {
		// If the request already has been added, do nothing and return, as if the event did not exist.
		// Returning here is important, because the rest of this function
		// must only be executed once for a request in an epoch (to prevent request duplication).
		return &events.EventList{}
	}

	// If necessary, notify the orderer responsible for this request to continue processing it.
	// (This is necessary in case the request has been "missing", i.e., proposed but not yet received.)
	eventsOut.PushBackList(iss.notifyOrderer(ref))

	// Count number of requests in all the buckets assigned to the same instance as the bucket of the received request.
	// These are all the requests pending to be proposed by the instance.
	// TODO: This is extremely inefficient and on the critical path
	//       (done on every request reception and TotalRequests loops over all buckets in the segment).
	//       Maintain a counter of requests in assigned buckets instead.
	pendingRequests := iss.buckets.Select(iss.bucketOrderers[bucket.ID].Segment().BucketIDs).TotalRequests()

	// If there are enough pending requests to fill a batch,
	// announce the total number of pending requests to the corresponding orderer.
	// Note that this deprives the orderer from the information about the number of pending requests,
	// as long as there are fewer of them than MaxBatchSize, but the orderer (so far) should not need this information.
	if pendingRequests >= iss.config.MaxBatchSize {
		eventsOut.PushBackList(iss.bucketOrderers[bucket.ID].ApplyEvent(SBPendingRequestsEvent(pendingRequests)))
	}

	return eventsOut
}

// applyAppSnapshot applies the event of the application creating a state snapshot.
// It passes the snapshot to the appropriate CheckpointTracker (identified by the event's associated epoch number).
func (iss *ISS) applyAppSnapshot(snapshot *eventpb.AppSnapshot) *events.EventList {
	if iss.epoch.Nr != t.EpochNr(snapshot.Epoch) {
		return &events.EventList{}
	}
	return iss.epoch.Checkpoint.ProcessAppSnapshot(snapshot.Data)
}

// applyLogEntryHashResult applies the event of receiving the digest of a delivered CommitLogEntry.
// It attaches the digest to the entry and inserts the entry to the commit log.
// Based on the state of the commitLog, it may trigger delivering batches to the application.
func (iss *ISS) applyLogEntryHashResult(digest []byte, logEntrySN t.SeqNr) *events.EventList {

	// Remove pending CommitLogEntry from the "waiting room"
	logEntry := iss.unhashedLogEntries[logEntrySN]
	delete(iss.unhashedLogEntries, logEntrySN)

	// Attach digest to entry.
	logEntry.Digest = digest

	// Insert the entry in the commitLog.
	iss.commitLog[logEntry.Sn] = logEntry

	// Deliver commitLog entries to the application in sequence number order.
	// This is relevant in the case when the sequence number of the currently SB-delivered batch
	// is the first sequence number not yet delivered to the application.
	return iss.processCommitted()

}

// applyAppSnapshotHashResult applies the event of receiving the digest of a delivered event of the application creating a state snapshot.
// It passes the snapshot hash to the appropriate CheckpointTracker (identified by the event's associated epoch number).
func (iss *ISS) applyAppSnapshotHashResult(digest []byte, epoch t.EpochNr) *events.EventList {
	if iss.epoch.Nr != epoch {
		return &events.EventList{}
	}
	return iss.epoch.Checkpoint.ProcessAppSnapshotHash(digest)
}

// applyCheckpointSignResult applies the event of receiving the Checkpoint message signature.
// It passes the signature to the appropriate CheckpointTracker (identified by the event's associated epoch number).
func (iss *ISS) applyCheckpointSignResult(signature []byte, epoch t.EpochNr) *events.EventList {
	if iss.epoch.Nr != epoch {
		return &events.EventList{}
	}
	return iss.epoch.Checkpoint.ProcessCheckpointSignResult(signature)
}

// It passes the signature verification result to the appropriate CheckpointTracker (identified by the event's associated epoch number).
func (iss *ISS) applyCheckpointSigVerResult(valid bool, err string, node t.NodeID, epoch t.EpochNr) *events.EventList {
	if iss.epoch.Nr != epoch {
		return &events.EventList{}
	}
	return iss.epoch.Checkpoint.ProcessSigVerified(valid, err, node)
}

// applySBEvent applies an event triggered by or addressed to an orderer (i.e., instance of Sequenced Broadcast),
// if that event belongs to the current epoch.
// TODO: Update this comment when the TODO below is addressed.
func (iss *ISS) applySBEvent(event *isspb.SBEvent) *events.EventList {

	// Ignore persist events (their handling is not yet implemented).
	// TODO: Deal with this when proper handlers are implemented.
	switch event.Event.Type.(type) {
	case *isspb.SBInstanceEvent_PbftPersistPreprepare,
		*isspb.SBInstanceEvent_PbftPersistPrepare,
		*isspb.SBInstanceEvent_PbftPersistCommit,
		*isspb.SBInstanceEvent_PbftPersistSignedViewChange,
		*isspb.SBInstanceEvent_PbftPersistNewView:
		return &events.EventList{}
	}

	switch epoch := t.EpochNr(event.Epoch); {
	case epoch > iss.epoch.Nr:
		// Events coming from future epochs should never occur (as, unlike messages, events are all generated locally.)
		panic(fmt.Sprintf("trying to handle ISS event (type %T, instance %d) from future epoch: %d",
			event.Event.Type, event.Instance, event.Epoch))

	case epoch == iss.epoch.Nr:
		// If the event is from the current epoch, apply it in relation to the corresponding orderer instance.
		return iss.applySBInstanceEvent(event.Event, t.SBInstanceNr(event.Instance))

	default: // epoch < iss.epoch:
		// If the event is from an old epoch, ignore it.
		// This might potentially happen if the epoch advanced while the event has been waiting in some buffer.
		// TODO: Is this really correct? Is it possible that orderers from old epochs
		//       (that have a reason to not yet be garbage-collected) still need to process events?
		iss.logger.Log(logging.LevelDebug, "Ignoring old event.", "epoch", epoch)
		return &events.EventList{}
	}
}

func (iss *ISS) applyStableCheckpoint(stableCheckpoint *isspb.StableCheckpoint) *events.EventList {
	eventsOut := &events.EventList{}

	if stableCheckpoint.Sn > iss.lastStableCheckpoint.Sn {
		// If this is the most recent checkpoint observed, save it.
		iss.logger.Log(logging.LevelInfo, "New stable checkpoint.",
			"epoch", stableCheckpoint.Epoch,
			"sn", stableCheckpoint.Sn,
			"replacingEpoch", iss.lastStableCheckpoint.Epoch,
			"replacingSn", iss.lastStableCheckpoint.Sn)
		iss.lastStableCheckpoint = stableCheckpoint

		// Send the new stable checkpoint to potentially
		// delayed nodes. The set of nodes to send the new
		// stable checkpoint to is determined based on the
		// highest epoch number in the messages received from
		// the node so far. If the highest epoch number we
		// have heard from the node is more than one epoch
		// behind then that node is likely to be left behind
		// and needs the stable checkpoint in order to start
		// catching up with state transfer.
		var delayed []t.NodeID
		for _, n := range iss.config.Membership {
			if t.EpochNr(stableCheckpoint.Epoch) > iss.nodeEpochMap[n]+1 {
				delayed = append(delayed, n)
			}
		}
		m := StableCheckpointMessage(stableCheckpoint)
		eventsOut.PushBack(events.SendMessage(m, delayed))

		// Prune old entries from WAL. The entries to prune
		// are determined according to the retention index
		// which is derived from the epoch number the new
		// stable checkpoint is associated with.
		eventsOut.PushBack(events.WALTruncate("wal", t.WALRetIndex(stableCheckpoint.Epoch)))

		// Clean up the global ISS state from all the epoch
		// instances that are associated with epoch numbers
		// less than the epoch number of the new stable
		// checkpoint associated with.
		for epoch := range iss.epochs {
			if epoch < t.EpochNr(stableCheckpoint.Epoch) {
				delete(iss.epochs, epoch)
			}
		}
	} else {
		iss.logger.Log(logging.LevelDebug, "Ignoring outdated stable checkpoint.", "sn", stableCheckpoint.Sn)
	}

	return eventsOut
}

// applyMessageReceived applies a message received over the network.
// Note that this is not the only place messages are applied.
// Messages received "ahead of time" that have been buffered are applied in applyBufferedMessages.
func (iss *ISS) applyMessageReceived(messageReceived *eventpb.MessageReceived) *events.EventList {

	// Convenience variables used for readability.
	message := messageReceived.Msg
	from := t.NodeID(messageReceived.From)

	// ISS only accepts ISS messages. If another message is applied, the next line panics.
	switch msg := message.Type.(*messagepb.Message_Iss).Iss.Type.(type) {
	case *isspb.ISSMessage_Checkpoint:
		return iss.applyCheckpointMessage(msg.Checkpoint, from)
	case *isspb.ISSMessage_StableCheckpoint:
		return iss.applyStableCheckpointMessage(msg.StableCheckpoint, from)
	case *isspb.ISSMessage_Sb:
		return iss.applySBMessage(msg.Sb, from)
	case *isspb.ISSMessage_RetransmitRequests:
		return iss.applyRetransmitRequestsMessage(msg.RetransmitRequests, from)
	default:
		panic(fmt.Errorf("unknown ISS message type: %T", msg))
	}
}

// applyCheckpointMessage relays a Checkpoint message received over the network to the appropriate CheckpointTracker.
func (iss *ISS) applyCheckpointMessage(message *isspb.Checkpoint, source t.NodeID) *events.EventList {

	// Remember the highest epoch number for each node to detect
	// later if the remote node is delayed too much and requires
	// assistance in order to catch up through state transfer.
	epoch := t.EpochNr(message.Epoch)
	if iss.nodeEpochMap[source] < epoch {
		iss.nodeEpochMap[source] = epoch
	}

	switch {
	case epoch > iss.epoch.Nr:
		// If the message is for a future epoch,
		// it might have been sent by a node that already transitioned to a newer epoch,
		// but this node is slightly behind (still in an older epoch) and cannot process the message yet.
		// In such case, save the message in a backlog (if there is buffer space) for later processing.
		iss.messageBuffers[source].Store(message)

		return &events.EventList{}

	case epoch == iss.epoch.Nr:
		// If the message is for the current epoch, check its validity and
		// apply it to the corresponding checkpoint tracker instance.
		return iss.epoch.Checkpoint.applyMessage(message, source)

	default: // epoch < iss.epoch.Nr:
		// Ignore old messages
		return &events.EventList{}
	}
}

// applyStableCheckpointMessage applies a StableCheckpoint message
// received over the network. It verifies the message and decides
// whether to install the state snapshot from the message.
func (iss *ISS) applyStableCheckpointMessage(m *isspb.StableCheckpoint, source t.NodeID) *events.EventList {
	eventsOut := &events.EventList{}

	// Check how far is the received stable checkpoint ahead of
	// the local node's state.
	if t.EpochNr(m.Epoch) <= iss.epoch.Nr+1 {
		// Ignore stable checkpoints that are not far enough
		// ahead of the current state of the local node.
		return &events.EventList{}
	}

	iss.logger.Log(logging.LevelDebug, "Installing state snapshot.", "epoch", m.Epoch)

	// Clean up global ISS state that belongs to the current epoch
	// instance that local replica got stuck with.
	iss.epochs = make(map[t.EpochNr]*epochInfo)
	iss.epoch = nil
	iss.commitLog = make(map[t.SeqNr]*CommitLogEntry)
	iss.unhashedLogEntries = make(map[t.SeqNr]*CommitLogEntry)
	iss.nextDeliveredSN = t.SeqNr(m.Sn)
	iss.newEpochSN = iss.nextDeliveredSN

	// Initialize a new ISS epoch instance for the new stable
	// checkpoint to continue participating in the protocol
	// starting with that epoch after installing the state
	// snapshot from the new stable checkpoint.
	iss.initEpoch(t.EpochNr(m.Epoch))

	// Update the last stable checkpoint stored in the global ISS structure.
	iss.lastStableCheckpoint = m

	// Create an event to request the application module for
	// restoring its state from the snapshot received in the new
	// stable checkpoint message.
	eventsOut.PushBack(events.AppRestoreState("app", m.AppSnapshot))

	// Activate SB instances of the new epoch which will deliver
	// batches after the application module has restored the state
	// from the snapshot.
	eventsOut.PushBackList(iss.initOrderers())

	// Apply any message buffered for the new epoch and append any
	// emitted event to the returns event list.
	eventsOut.PushBackList(iss.applyBufferedMessages())

	return eventsOut
}

// applySBMessage applies a message destined for an orderer (i.e. a Sequenced Broadcast implementation).
func (iss *ISS) applySBMessage(message *isspb.SBMessage, from t.NodeID) *events.EventList {

	switch epoch := t.EpochNr(message.Epoch); {
	case epoch > iss.epoch.Nr:
		// If the message is for a future epoch,
		// it might have been sent by a node that already transitioned to a newer epoch,
		// but this node is slightly behind (still in an older epoch) and cannot process the message yet.
		// In such case, save the message in a backlog (if there is buffer space) for later processing.
		iss.messageBuffers[from].Store(message)
		return &events.EventList{}

	case epoch == iss.epoch.Nr:
		// If the message is for the current epoch, check its validity and
		// apply it to the corresponding orderer in form of an SBMessageReceived event.
		if err := iss.validateSBMessage(message, from); err != nil {
			iss.logger.Log(logging.LevelWarn, "Ignoring invalid SB message.",
				"type", fmt.Sprintf("%T", message.Msg.Type), "from", from, "error", err)
			return &events.EventList{}
		}

		return iss.applySBInstanceEvent(SBMessageReceivedEvent(message.Msg, from), t.SBInstanceNr(message.Instance))
	default: // epoch < iss.epoch:
		// Ignore old messages
		// TODO: In case old SB instances from previous epoch still need to linger around,
		//       they might need to receive messages... Instead of simply checking the message epoch
		//       against the current epoch, we might need to remember which epochs have been already garbage-collected
		//       and use that information to decide what to do with messages.
		return &events.EventList{}
	}
}

// applyRetransmitRequestsMessage applies a message demanding request retransmission to a node
// that received a proposal containing some requests, but was not yet able to authenticate those requests.
// TODO: Implement this function. See demandRequestRetransmission for more comments.
func (iss *ISS) applyRetransmitRequestsMessage(req *isspb.RetransmitRequests, from t.NodeID) *events.EventList {
	iss.logger.Log(logging.LevelWarn, "UNIMPLEMENTED: Ignoring request retransmission request.",
		"from", from, "numReqs", len(req.Requests))
	return &events.EventList{}
}

// ============================================================
// Additional protocol logic
// ============================================================

// initEpoch initializes a new ISS epoch with the given epoch number.
func (iss *ISS) initEpoch(newEpoch t.EpochNr) {

	iss.logger.Log(logging.LevelInfo, "New epoch", "epochNr", newEpoch)

	// Set the new epoch number and re-initialize list of orderers.
	epoch := &epochInfo{
		Nr: newEpoch,
		Checkpoint: newCheckpointTracker(iss.ownID, iss.nextDeliveredSN, newEpoch,
			logging.Decorate(iss.logger, "CT: ", "epoch", newEpoch)),
	}
	iss.epochs[newEpoch] = epoch
	iss.epoch = epoch

	// Compute the set of leaders for the new epoch.
	// Note that leader policy is stateful, choosing leaders deterministically based on the state of the system.
	// Its state must be consistent across all nodes when calling Leaders() on it.
	leaders := iss.config.LeaderPolicy.Leaders(newEpoch)

	// Compute the assignment of buckets to orderers (each leader will correspond to one orderer).
	leaderBuckets := iss.buckets.Distribute(leaders, newEpoch)

	// Reset missing request tracking information.
	// These fields could still contain some data if any preprepared batches were not committed on the "good path",
	// e.g., committed using state transfer or not committed at all.
	iss.missingRequests = make(map[missingRequestInfoRef]*missingRequestInfo)
	iss.missingRequestIndex = make(map[string]*missingRequestInfo)

	// Initialize index of orderers based on the buckets they are assigned.
	// Given a bucket, this index helps locate the orderer to which the bucket is assigned.
	iss.bucketOrderers = make(map[int]sbInstance)

	// Create new segments of the commit log, one per leader selected by the leader selection policy.
	// Instantiate one orderer (SB instance) for each segment.
	for i, leader := range leaders {

		// Create segment.
		seg := &segment{
			Leader:     leader,
			Membership: iss.config.Membership,
			SeqNrs: sequenceNumbers(
				iss.nextDeliveredSN+t.SeqNr(i),
				t.SeqNr(len(leaders)),
				iss.config.SegmentLength),
			BucketIDs: leaderBuckets[leader],
		}
		iss.newEpochSN += t.SeqNr(len(seg.SeqNrs))

		// Instantiate a new PBFT orderer.
		// TODO: When more protocols are implemented, make this configurable, so other orderer types can be chosen.
		sbInst := newPbftInstance(
			iss.ownID,
			seg,
			iss.buckets.Select(seg.BucketIDs).TotalRequests(),
			newPBFTConfig(iss.config),
			&sbEventService{epoch: newEpoch, instance: t.SBInstanceNr(i)},
			logging.Decorate(iss.logger, "PBFT: ", "epoch", newEpoch, "instance", i))

		// Add the orderer to the list of orderers.
		iss.epoch.Orderers = append(iss.epoch.Orderers, sbInst)

		// Populate index of orderers based on the buckets they are assigned.
		for _, bID := range seg.BucketIDs {
			iss.bucketOrderers[bID] = sbInst
		}
	}
}

// initOrderers sends the SBInit event to all orderers in the current epoch.
func (iss *ISS) initOrderers() *events.EventList {
	eventsOut := &events.EventList{}

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

// validateSBMessage checks whether an SBMessage is valid in the current epoch.
// Returns nil if validation succeeds.
// If validation fails, returns the reason for which the message is considered invalid.
func (iss *ISS) validateSBMessage(message *isspb.SBMessage, from t.NodeID) error {

	// Message must be destined for the current epoch.
	if t.EpochNr(message.Epoch) != iss.epoch.Nr {
		return fmt.Errorf("invalid epoch: %v (current epoch is %v)", message.Instance, iss.epoch)
	}

	// Message must refer to a valid SB instance.
	if int(message.Instance) > len(iss.epoch.Orderers) {
		return fmt.Errorf("invalid SB instance number: %d", message.Instance)
	}

	// Message must be sent by a node in the current membership.
	// TODO: This lookup is extremely inefficient, computing the membership set on each message validation.
	//       Cache the output of membershipSet() throughout the epoch.
	if _, ok := membershipSet(iss.config.Membership)[from]; !ok {
		return fmt.Errorf("sender of SB message not in the membership: %v", from)
	}

	return nil
}

// notifyOrderer checks whether a request is the last one missing from a proposal
// and, if so, notifies the corresponding orderer.
// If some orderer instance is waiting for a request (i.e., the request has been "missing" - proposed but not received),
// notifyOrderer marks the request as not missing any more and (if no more requests are missing) notifies that orderer
// to continue processing the corresponding proposal.
func (iss *ISS) notifyOrderer(reqRef *requestpb.RequestRef) *events.EventList {

	// Calculate the map key (string representation) of the request reference.
	reqKey := reqStrKey(reqRef)

	// If the request has been missing
	if missingRequests := iss.missingRequestIndex[reqKey]; missingRequests != nil {

		// Remove the request from the missing request index.
		delete(missingRequests.Requests, reqKey)

		// Remove the request from the list of missing requests of the corresponding proposal.
		delete(iss.missingRequestIndex, reqKey)

		// If this was the last missing request of the proposal
		if len(missingRequests.Requests) == 0 {

			// Remove the proposal from the set of proposals with missing requests
			// and notify the corresponding orderer that no more requests from this proposal are missing.
			delete(iss.missingRequests, missingRequests.Ref)
			return missingRequests.Orderer.ApplyEvent(SBRequestsReady(missingRequests.Ref.SBRef))
		}
	}

	// If the orderer need not be notified yet, do nothing.
	return &events.EventList{}
}

// demandRequestRetransmission asks for the retransmission of missing requests.
// The result of this call should ultimately be a RequestReady event occurring for each of the missing requests.
// TODO: Explain that this can be done by asking the leader for a verifiable request authenticator
//       (e.g. a client signature) or contacting multiple nodes to confirm the reception of this request,
//       or even by asking the client directly to retransmit the request to this node.
// TODO: Implement at least one of these options.
// TODO: Remove the //nolint:unused comment when used again.
func (iss *ISS) demandRequestRetransmission(reqInfo *missingRequestInfo) *events.EventList { //nolint:unused

	// Create a slice of requests for which to demand retransmission.
	// This is only necessary because they are stored in a map.
	requests := make([]*requestpb.RequestRef, 0, len(reqInfo.Requests))
	// TODO: iterate in a deterministic order!
	for _, reqRef := range reqInfo.Requests {
		requests = append(requests, reqRef)
	}

	// Send a message to the leader that made the proposal for which requests are still missing.
	return (&events.EventList{}).PushBack(events.SendMessage(
		RetransmitRequestsMessage(requests), []t.NodeID{reqInfo.Orderer.Segment().Leader},
	))
}

// processCommitted delivers entries from the commitLog in order of their sequence numbers.
// Whenever a new entry is inserted in the commitLog, this function must be called
// to create Deliver events for all the batches that can be delivered to the application.
// processCommitted also triggers other internal Events like epoch transitions and state checkpointing.
func (iss *ISS) processCommitted() *events.EventList {
	eventsOut := &events.EventList{}

	// The iss.nextDeliveredSN variable always contains the lowest sequence number
	// for which no batch has been delivered yet.
	// As long as there is an entry in the commitLog with that sequence number,
	// deliver the corresponding batch and advance to the next sequence number.
	for iss.commitLog[iss.nextDeliveredSN] != nil {

		// TODO: Once system configuration requests are introduced, apply them here.

		// Create a new Deliver event.
		eventsOut.PushBack(events.Deliver("app", iss.nextDeliveredSN, iss.commitLog[iss.nextDeliveredSN].Batch))

		// Output debugging information.
		iss.logger.Log(logging.LevelDebug, "Delivering entry.",
			"sn", iss.nextDeliveredSN, "nReq", len(iss.commitLog[iss.nextDeliveredSN].Batch.Requests))

		// Remove just delivered batch from the temporary
		// store of batches that were agreed upon out-of-order.
		delete(iss.commitLog, iss.nextDeliveredSN)

		// Increment the sequence number of the next batch to deliver.
		iss.nextDeliveredSN++
	}

	// If the epoch is finished, transition to the next epoch.
	if iss.epochFinished() {

		// Initialize the internal data structures for the new epoch.
		iss.initEpoch(iss.epoch.Nr + 1)

		// Look up a (or create a new) checkpoint tracker and start the checkpointing protocol.
		// This must happen after initialization of the new epoch,
		// as the sequence number the checkpoint will be associated with (iss.nextDeliveredSN)
		// is already part of the new epoch.
		// The checkpoint tracker might already exist if a corresponding message has been already received.
		// iss.nextDeliveredSN is the first sequence number *not* included in the checkpoint,
		// i.e., as sequence numbers start at 0, the checkpoint includes the first iss.nextDeliveredSN sequence numbers.
		eventsOut.PushBackList(iss.epoch.Checkpoint.Start(iss.config.Membership))

		// Give the init signals to the newly instantiated orderers.
		// TODO: Currently this probably sends the Init event to old orderers as well.
		//       That should not happen! Investigate and fix.
		eventsOut.PushBackList(iss.initOrderers())

		// Process backlog of buffered SB messages.
		eventsOut.PushBackList(iss.applyBufferedMessages())
	}

	return eventsOut
}

// applyBufferedMessages applies all SB messages destined to the current epoch
// that have been buffered during past epochs.
// This function is always called directly after initializing a new epoch, except for epoch 0.
func (iss *ISS) applyBufferedMessages() *events.EventList {
	eventsOut := &events.EventList{}

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

// removeFromBuckets removes the given requests from their corresponding buckets.
// This happens when a batch is committed.
// TODO: Implement marking requests as "in flight"/proposed, so we don't accept proposals with duplicates.
//       This could maybe be done on the WaitForRequests/RequestsReady path...
func (iss *ISS) removeFromBuckets(requests []*requestpb.RequestRef) {

	// Remove each request from its bucket.
	for _, reqRef := range requests {
		iss.buckets.RequestBucket(reqRef).Remove(reqRef)
	}
}

// bufferedMessageFilter decides, given a message, whether it is appropriate to apply the message, discard it,
// or keep it in the buffer, returning the appropriate value of type messagebuffer.Applicable.
func (iss *ISS) bufferedMessageFilter(source t.NodeID, message proto.Message) messagebuffer.Applicable {
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
func reqStrKey(reqRef *requestpb.RequestRef) string {
	return fmt.Sprintf("%v-%d.%v", reqRef.ClientId, reqRef.ReqNo, reqRef.Digest)
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
func newPBFTConfig(issConfig *Config) *PBFTConfig {

	// Make a copy of the current membership.
	pbftMembership := make([]t.NodeID, len(issConfig.Membership))
	copy(pbftMembership, issConfig.Membership)

	// Return a new PBFT configuration with selected values from the ISS configuration.
	return &PBFTConfig{
		Membership:               issConfig.Membership,
		MaxProposeDelay:          issConfig.MaxProposeDelay,
		MsgBufCapacity:           issConfig.MsgBufCapacity,
		MaxBatchSize:             issConfig.MaxBatchSize,
		ViewChangeBatchTimeout:   issConfig.PBFTViewChangeBatchTimeout,
		ViewChangeSegmentTimeout: issConfig.PBFTViewChangeSegmentTimeout,
		ViewChangeResendPeriod:   issConfig.PBFTViewChangeResendPeriod,
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

	// Encode the batch content.
	batchData := serializing.BatchForHash(entry.Batch)

	// Put everything together in a slice and return it.
	data := make([][]byte, 0, len(entry.Batch.Requests)+3)
	data = append(data, snBuf, suspectBuf, []byte{aborted})
	data = append(data, batchData...)
	return data
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

func iterateSorted[K constraints.Ordered, V any](m map[K]V, f func(key K, value V) (cont bool)) {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	for _, k := range keys {
		if !f(k, m[k]) {
			break
		}
	}
}
