// TODO: Finish writing proper comments in this file.

package checkpoint

import (
	"bytes"
	"time"

	"github.com/filecoin-project/mir/pkg/iss/config"

	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	apppbdsl "github.com/filecoin-project/mir/pkg/pb/apppb/dsl"
	checkpointpbdsl "github.com/filecoin-project/mir/pkg/pb/checkpointpb/dsl"
	checkpointpbmsgs "github.com/filecoin-project/mir/pkg/pb/checkpointpb/msgs"
	checkpointpbtypes "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	cryptopbdsl "github.com/filecoin-project/mir/pkg/pb/cryptopb/dsl"
	eventpbdsl "github.com/filecoin-project/mir/pkg/pb/eventpb/dsl"
	eventpbtypes "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	hasherpbdsl "github.com/filecoin-project/mir/pkg/pb/hasherpb/dsl"
	transportpbevents "github.com/filecoin-project/mir/pkg/pb/transportpb/events"
	trantorpbdsl "github.com/filecoin-project/mir/pkg/pb/trantorpb/dsl"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	"github.com/filecoin-project/mir/pkg/timer/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
)

const (
	DefaultResendPeriod = types.Duration(time.Second)
)

type State struct {
	// State snapshot associated with this checkpoint.
	StateSnapshot *trantorpbtypes.StateSnapshot

	// Hash of the state snapshot data associated with this checkpoint.
	StateSnapshotHash []byte

	// Set of (potentially invalid) nodes' Signatures.
	Signatures map[t.NodeID][]byte

	// Set of nodes from which a valid Checkpoint messages has been received.
	SigReceived map[t.NodeID]struct{}

	// Set of Checkpoint messages that were received ahead of time.
	PendingMessages map[t.NodeID]*checkpointpbtypes.Checkpoint

	// Flag ensuring that the stable checkpoint is only Announced once.
	// Set to true when announcing a stable checkpoint for the first time.
	// When true, stable checkpoints are not Announced anymore.
	Announced bool
}

// NewModule allocates and returns a new instance of the ModuleParams associated with sequence number sn.
func NewModule(
	moduleConfig ModuleConfig,
	params *ModuleParams,
	logger logging.Logger) modules.PassiveModule {

	state := &State{
		StateSnapshot: &trantorpbtypes.StateSnapshot{
			AppData: nil,
			EpochData: &trantorpbtypes.EpochData{
				EpochConfig:        params.EpochConfig,
				ClientProgress:     nil, // This will be filled by a separate event.
				LeaderPolicy:       params.LeaderPolicyData,
				PreviousMembership: params.Membership,
			},
		},
		Announced:       false,
		Signatures:      make(map[t.NodeID][]byte),
		SigReceived:     make(map[t.NodeID]struct{}),
		PendingMessages: make(map[t.NodeID]*checkpointpbtypes.Checkpoint),
	}

	m := dsl.NewModule(moduleConfig.Self)

	apppbdsl.UponSnapshot(m, func(appData []uint8) error {
		// Treat nil data as an empty byte slice.
		if appData == nil {
			appData = []byte{}
		}

		// Save the received app snapshot if there is none yet.
		if state.StateSnapshot.AppData == nil {
			state.StateSnapshot.AppData = appData
			if state.SnapshotReady() {
				processStateSnapshot(m, state, moduleConfig)
			}
		}
		return nil
	})

	hasherpbdsl.UponResultOne(m, func(digest []uint8, _ *struct{}) error {
		// Save the received snapshot hash
		state.StateSnapshotHash = digest

		// Request signature
		sigData := serializeCheckpointForSig(params.EpochConfig.EpochNr, params.EpochConfig.FirstSn, state.StateSnapshotHash)

		cryptopbdsl.SignRequest(m, moduleConfig.Crypto, sigData, &struct{}{})

		return nil
	})

	cryptopbdsl.UponSignResult(m, func(sig []uint8, _ *struct{}) error {

		// Save received own checkpoint signature
		state.Signatures[params.OwnID] = sig
		state.SigReceived[params.OwnID] = struct{}{}

		// In case the node's own signature is enough to reach quorum, announce the stable checkpoint.
		// This can happen in a small system where no failures are tolerated.
		if state.Stable(params) {
			announceStable(m, params, state, moduleConfig)
		}

		// Send a checkpoint message to all nodes after persisting checkpoint to the WAL.
		chkpMessage := checkpointpbmsgs.Checkpoint(moduleConfig.Self, params.EpochConfig.EpochNr, params.EpochConfig.FirstSn, state.StateSnapshotHash, sig)
		sortedMembership := maputil.GetSortedKeys(params.Membership.Nodes)
		eventpbdsl.TimerRepeat(m,
			"timer",
			[]*eventpbtypes.Event{transportpbevents.SendMessage(moduleConfig.Net, chkpMessage, sortedMembership)},
			params.ResendPeriod,
			tt.RetentionIndex(params.EpochConfig.EpochNr),
		)

		logger.Log(logging.LevelDebug, "Sending checkpoint message",
			"epoch", params.EpochConfig.EpochNr,
			"dataLen", len(state.StateSnapshot.AppData),
			"memberships", len(state.StateSnapshot.EpochData.EpochConfig.Memberships),
		)

		// Apply pending Checkpoint messages
		for s, msg := range state.PendingMessages {
			if err := applyCheckpointReceived(m, params, state, moduleConfig, s, msg.Epoch, msg.Sn, msg.SnapshotHash, msg.Signature, logger); err != nil {
				logger.Log(logging.LevelWarn, "Error applying pending Checkpoint message", "error", err, "msg", msg)
				return err
			}

		}
		state.PendingMessages = nil

		return nil
	})

	cryptopbdsl.UponSigVerified(m, func(nodeId t.NodeID, err error, c *verificationContext) error {
		if err != nil {
			logger.Log(logging.LevelWarn, "Ignoring Checkpoint message. Invalid signature.", "source", nodeId, "error", err)
			return nil
		}

		// Note the reception of a valid Checkpoint message from node `source`.
		state.Signatures[nodeId] = c.signature

		// If, after having applied this message, the checkpoint became stable, produce the necessary events.
		if state.Stable(params) {
			announceStable(m, params, state, moduleConfig)
		}

		return nil
	})

	trantorpbdsl.UponClientProgress(m, func(progress map[tt.ClientID]*trantorpbtypes.DeliveredTXs) error {
		// Save the received client progress if there is none yet.
		if state.StateSnapshot.EpochData.ClientProgress == nil {
			state.StateSnapshot.EpochData.ClientProgress = &trantorpbtypes.ClientProgress{
				Progress: progress,
			}
			if state.SnapshotReady() {
				processStateSnapshot(m, state, moduleConfig)
			}
		}
		return nil
	})

	checkpointpbdsl.UponCheckpointReceived(m, func(from t.NodeID, epoch tt.EpochNr, sn tt.SeqNr, snapshotHash []uint8, signature []uint8) error {
		return applyCheckpointReceived(m, params, state, moduleConfig, from, epoch, sn, snapshotHash, signature, logger)
	})

	return m
}

func processStateSnapshot(m dsl.Module, state *State, mc ModuleConfig) {

	// Initiate computing the hash of the snapshot.
	hasherpbdsl.RequestOne(m,
		mc.Hasher,
		serializeSnapshotForHash(state.StateSnapshot),
		&struct{}{},
	)
}

func announceStable(m dsl.Module, p *ModuleParams, state *State, mc ModuleConfig) {

	// Only announce the stable checkpoint once.
	if state.Announced {
		return
	}
	state.Announced = true

	// Assemble a multisig certificate from the received signatures.
	cert := make(map[t.NodeID][]byte)
	for node, sig := range state.Signatures {
		cert[node] = sig
	}

	// Announce the stable checkpoint to the ordering protocol.
	checkpointpbdsl.StableCheckpoint(m, mc.Ord, p.EpochConfig.FirstSn, state.StateSnapshot, cert)
}

func applyCheckpointReceived(m dsl.Module,
	p *ModuleParams,
	state *State,
	moduleConfig ModuleConfig,
	from t.NodeID,
	epoch tt.EpochNr,
	sn tt.SeqNr,
	snapshotHash []uint8,
	signature []uint8,
	logger logging.Logger) error {

	// check if from is part of the membership
	if _, ok := p.Membership.Nodes[from]; !ok {
		logger.Log(logging.LevelWarn, "sender %s is not a member.\n", from)
		return nil
	}
	// Notify the protocol about the progress of the from node.
	// If no progress is made for a configured number of epochs,
	// the node is considered to be a straggler and is sent a stable checkpoint to catch uparams.
	checkpointpbdsl.EpochProgress(m, moduleConfig.Ord, from, epoch)

	// If checkpoint is already stable, ignore message.
	if state.Stable(p) {
		return nil
	}

	// Check snapshot hash
	if state.StateSnapshotHash == nil {
		// The message is received too early, put it aside
		state.PendingMessages[from] = &checkpointpbtypes.Checkpoint{
			Epoch:        epoch,
			Sn:           sn,
			SnapshotHash: snapshotHash,
			Signature:    signature,
		}
		return nil
	} else if !bytes.Equal(state.StateSnapshotHash, snapshotHash) {
		// Snapshot hash mismatch
		logger.Log(logging.LevelWarn, "Ignoring Checkpoint message. Mismatching state snapshot hash.", "from", from)
		return nil
	}

	// Ignore duplicate messages.
	if _, ok := state.SigReceived[from]; ok {
		return nil
	}
	state.SigReceived[from] = struct{}{}

	// Verify signature of the sender.
	sigData := serializeCheckpointForSig(p.EpochConfig.EpochNr, p.EpochConfig.FirstSn, state.StateSnapshotHash)
	cryptopbdsl.VerifySig(m,
		moduleConfig.Crypto,
		sigData,
		signature,
		from,
		&verificationContext{signature: signature},
	)

	return nil
}

func (state *State) SnapshotReady() bool {
	return state.StateSnapshot.AppData != nil &&
		state.StateSnapshot.EpochData.ClientProgress != nil
}

func (state *State) Stable(p *ModuleParams) bool {
	return state.SnapshotReady() && len(state.Signatures) >= config.StrongQuorum(len(p.Membership.Nodes))
}

type verificationContext struct {
	signature []uint8
}
