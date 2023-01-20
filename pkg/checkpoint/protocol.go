// TODO: Finish writing proper comments in this file.

package checkpoint

import (
	"bytes"
	"time"

	"github.com/pkg/errors"

	"github.com/filecoin-project/mir/pkg/checkpoint/protobufs"
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/batchfetcherpb"
	"github.com/filecoin-project/mir/pkg/pb/checkpointpb"
	"github.com/filecoin-project/mir/pkg/pb/commonpb"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	"github.com/filecoin-project/mir/pkg/serializing"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
)

const (
	DefaultResendPeriod = t.TimeDuration(time.Second)
)

// Protocol represents the state associated with a single instance of the checkpoint protocol
// (establishing a single stable checkpoint).
type Protocol struct {
	logging.Logger

	// IDs of modules the checkpoint tracker interacts with.
	// TODO: Eventually put the checkpoint tracker in a separate package and create its own ModuleConfig type.
	moduleConfig *ModuleConfig

	// The ID of the node executing this instance of the protocol.
	ownID t.NodeID

	// Epoch to which this checkpoint belongs.
	// It is always the epoch the checkpoint's associated sequence number (seqNr) is part of.
	epoch t.EpochNr

	// Sequence number associated with this checkpoint protocol instance.
	// This checkpoint encompasses seqNr sequence numbers,
	// i.e., seqNr is the first sequence number *not* encompassed by this checkpoint.
	// One can imagine that the checkpoint represents the state of the system just before seqNr,
	// i.e., "between" seqNr-1 and seqNr.
	seqNr t.SeqNr

	// The IDs of nodes to execute this instance of the checkpoint protocol.
	// Note that it is the membership of epoch e-1 that constructs the membership for epoch e.
	// (As the starting checkpoint for e is the "finishing" checkpoint for e-1.)
	membership []t.NodeID

	// State snapshot associated with this checkpoint.
	stateSnapshot *commonpb.StateSnapshot

	// Hash of the state snapshot data associated with this checkpoint.
	stateSnapshotHash []byte

	// Set of (potentially invalid) nodes' signatures.
	signatures map[t.NodeID][]byte

	// Set of nodes from which a valid Checkpoint messages has been received.
	confirmations map[t.NodeID]struct{}

	// Set of Checkpoint messages that were received ahead of time.
	pendingMessages map[t.NodeID]*checkpointpb.Checkpoint

	// Time interval for repeated retransmission of checkpoint messages.
	resendPeriod t.TimeDuration
}

// NewProtocol allocates and returns a new instance of the Protocol associated with sequence number sn.
func NewProtocol(
	moduleConfig *ModuleConfig,
	ownID t.NodeID,
	membership map[t.NodeID]t.NodeAddress,
	epochConfig *commonpb.EpochConfig,
	leaderPolicyData []byte,
	resendPeriod t.TimeDuration,
	logger logging.Logger,
) *Protocol {
	return &Protocol{
		Logger:          logger,
		moduleConfig:    moduleConfig,
		ownID:           ownID,
		seqNr:           t.SeqNr(epochConfig.FirstSn),
		epoch:           t.EpochNr(epochConfig.EpochNr),
		resendPeriod:    resendPeriod,
		signatures:      make(map[t.NodeID][]byte),
		confirmations:   make(map[t.NodeID]struct{}),
		pendingMessages: make(map[t.NodeID]*checkpointpb.Checkpoint),
		membership:      maputil.GetSortedKeys(membership),
		stateSnapshot: &commonpb.StateSnapshot{
			AppData: nil,
			EpochData: &commonpb.EpochData{
				EpochConfig:        epochConfig,
				ClientProgress:     nil, // This will be filled by a separate event.
				LeaderPolicy:       leaderPolicyData,
				PreviousMembership: t.MembershipPb(membership),
			},
		},
	}
}

func (p *Protocol) ImplementsModule() {}

func (p *Protocol) ApplyEvents(evts *events.EventList) (*events.EventList, error) {
	return modules.ApplyEventsSequentially(evts, p.applyEvent)
}

func (p *Protocol) applyEvent(event *eventpb.Event) (*events.EventList, error) {
	switch e := event.Type.(type) {
	case *eventpb.Event_Init:
		return events.EmptyList(), nil // Nothing to initialize.
	case *eventpb.Event_AppSnapshot:
		return p.applyAppSnapshot(e.AppSnapshot)
	case *eventpb.Event_HashResult:
		return p.applyHashResult(e.HashResult)
	case *eventpb.Event_SignResult:
		return p.applySignResult(e.SignResult)
	case *eventpb.Event_NodeSigsVerified:
		return p.applyNodeSigsVerified(e.NodeSigsVerified)
	case *eventpb.Event_BatchFetcher:
		switch e := e.BatchFetcher.Type.(type) {
		case *batchfetcherpb.Event_ClientProgress:
			return p.applyClientProgress(e.ClientProgress)
		default:
			return nil, errors.Errorf("unexpected batch fetcher event type: %T", e)
		}
	case *eventpb.Event_MessageReceived:
		switch msg := e.MessageReceived.Msg.Type.(type) {
		case *messagepb.Message_Checkpoint:
			switch m := msg.Checkpoint.Type.(type) {
			case *checkpointpb.Message_Checkpoint:
				return p.applyMessage(m.Checkpoint, t.NodeID(e.MessageReceived.From)), nil
			default:
				return nil, errors.Errorf("unexpected checkpoint message type: %T", m)
			}
		default:
			return nil, errors.Errorf("unexpected message type: %T", e.MessageReceived.Msg.Type)
		}
	default:
		return nil, errors.Errorf("unknown event type: %T", e)
	}
}

func (p *Protocol) applyAppSnapshot(appSnapshot *eventpb.AppSnapshot) (*events.EventList, error) {

	// Treat nil data as an empty byte slice.
	var appData []byte
	if appSnapshot.AppData != nil {
		appData = appSnapshot.AppData
	} else {
		appData = []byte{}
	}

	// Save the received app snapshot if there is none yet.
	if p.stateSnapshot.AppData == nil {
		p.stateSnapshot.AppData = appData
		if p.snapshotReady() {
			return p.processStateSnapshot()
		}
	}
	return events.EmptyList(), nil
}

func (p *Protocol) applyClientProgress(clientProgress *commonpb.ClientProgress) (*events.EventList, error) {

	// Save the received client progress if there is none yet.
	if p.stateSnapshot.EpochData.ClientProgress == nil {
		p.stateSnapshot.EpochData.ClientProgress = clientProgress
		if p.snapshotReady() {
			return p.processStateSnapshot()
		}
	}
	return events.EmptyList(), nil
}

func (p *Protocol) snapshotReady() bool {
	return p.stateSnapshot.AppData != nil &&
		p.stateSnapshot.EpochData.ClientProgress != nil
}

func (p *Protocol) processStateSnapshot() (*events.EventList, error) {

	// Initiate computing the hash of the snapshot.
	return events.ListOf(events.HashRequest(
		p.moduleConfig.Hasher,
		[][][]byte{serializing.SnapshotForHash(p.stateSnapshot)},
		protobufs.HashOrigin(p.moduleConfig.Self),
	)), nil
}

func (p *Protocol) applyHashResult(result *eventpb.HashResult) (*events.EventList, error) {

	// Save the received snapshot hash
	p.stateSnapshotHash = result.Digests[0]

	// Request signature
	sigData := serializing.CheckpointForSig(p.epoch, p.seqNr, p.stateSnapshotHash)
	return events.ListOf(events.SignRequest(
		p.moduleConfig.Crypto,
		sigData,
		protobufs.SignOrigin(p.moduleConfig.Self),
	)), nil
}

func (p *Protocol) applySignResult(result *eventpb.SignResult) (*events.EventList, error) {

	eventsOut := events.EmptyList()

	// Save received own checkpoint signature
	p.signatures[p.ownID] = result.Signature
	p.confirmations[p.ownID] = struct{}{}

	// In case the node's own signature is enough to reach quorum, announce the stable checkpoint.
	// This can happen in a small system where no failures are tolerated.
	if p.stable() {
		eventsOut.PushBackList(p.announceStable())
	}

	// Send a checkpoint message to all nodes after persisting checkpoint to the WAL.
	chkpMessage := protobufs.CheckpointMessage(p.moduleConfig.Self, p.epoch, p.seqNr, p.stateSnapshotHash, result.Signature)
	eventsOut.PushBack(events.TimerRepeat(
		"timer",
		[]*eventpb.Event{events.SendMessage(p.moduleConfig.Net, chkpMessage, p.membership)},
		p.resendPeriod,
		t.RetentionIndex(p.epoch)),
	)

	p.Log(logging.LevelDebug, "Sending checkpoint message",
		"epoch", p.epoch,
		"dataLen", len(p.stateSnapshot.AppData),
		"memberships", len(p.stateSnapshot.EpochData.EpochConfig.Memberships),
	)

	// Apply pending Checkpoint messages
	for s, m := range p.pendingMessages {
		eventsOut.PushBackList(p.applyMessage(m, s))
	}
	p.pendingMessages = nil

	// Return resulting WALEvent with the SendMessage event
	// (and potential results of pending message application) appended.
	return eventsOut, nil
}

func (p *Protocol) applyMessage(msg *checkpointpb.Checkpoint, source t.NodeID) *events.EventList {
	eventsOut := events.EmptyList()

	// Notify the protocol about the progress of the source node.
	// If no progress is made for a configured number of epochs,
	// the node is considered to be a straggler and is sent a stable checkpoint to catch up.
	eventsOut.PushBack(protobufs.EpochProgressEvent(p.moduleConfig.Ord, source, t.EpochNr(msg.Epoch)))

	// If checkpoint is already stable, ignore message.
	if p.stable() {
		return eventsOut
	}

	// Check snapshot hash
	if p.stateSnapshotHash == nil {
		// The message is received too early, put it aside
		p.pendingMessages[source] = msg
		return eventsOut
	} else if !bytes.Equal(p.stateSnapshotHash, msg.SnapshotHash) {
		// Snapshot hash mismatch
		p.Log(logging.LevelWarn, "Ignoring Checkpoint message. Mismatching state snapshot hash.", "source", source)
		return eventsOut
	}

	// TODO: Only accept messages from nodes in membership.
	//       This might be more tricky than it seems, especially when the membership is not yet initialized.

	// Ignore duplicate messages.
	if _, ok := p.signatures[source]; ok {
		return eventsOut
	}
	p.signatures[source] = msg.Signature

	// Verify signature of the sender.
	sigData := serializing.CheckpointForSig(p.epoch, p.seqNr, p.stateSnapshotHash)
	eventsOut.PushBack(events.VerifyNodeSigs(
		p.moduleConfig.Crypto,
		[][][]byte{sigData},
		[][]byte{msg.Signature},
		[]t.NodeID{source},
		protobufs.SigVerOrigin(p.moduleConfig.Self),
	))

	return eventsOut
}

func (p *Protocol) applyNodeSigsVerified(result *eventpb.NodeSigsVerified) (*events.EventList, error) {

	// A checkpoint only has one signature and thus each slice of the result only contains one element.
	sourceNode := t.NodeID(result.NodeIds[0])
	err := result.Errors[0]

	if !result.AllOk {
		p.Log(logging.LevelWarn, "Ignoring Checkpoint message. Invalid signature.", "source", sourceNode, "error", err)
		p.signatures[sourceNode] = nil
		return events.EmptyList(), nil
	}

	// Note the reception of a valid Checkpoint message from node `source`.
	p.confirmations[sourceNode] = struct{}{}

	// If, after having applied this message, the checkpoint became stable, produce the necessary events.
	if p.stable() {
		return p.announceStable(), nil
	}

	return events.EmptyList(), nil
}

func (p *Protocol) stable() bool {
	return p.snapshotReady() && len(p.confirmations) >= strongQuorum(len(p.membership))
}

func (p *Protocol) announceStable() *events.EventList {

	// Assemble a multisig certificate from the received signatures.
	cert := make(map[string][]byte)
	for node := range p.confirmations {
		cert[node.Pb()] = p.signatures[node]
	}

	// Create a stable checkpoint object.
	stableCheckpoint := &checkpointpb.StableCheckpoint{
		Sn:       p.seqNr.Pb(),
		Snapshot: p.stateSnapshot,
		Cert:     cert,
	}

	// Announce the stable checkpoint to the ordering protocol.
	return events.ListOf(protobufs.StableCheckpointEvent(p.moduleConfig.Ord, stableCheckpoint))
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
