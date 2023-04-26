// TODO: Finish writing proper comments in this file.

package checkpoint

import (
	"bytes"
	"time"

	"github.com/filecoin-project/mir/pkg/checkpoint/common"
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	apppbdsl "github.com/filecoin-project/mir/pkg/pb/apppb/dsl"
	checkpointpbdsl "github.com/filecoin-project/mir/pkg/pb/checkpointpb/dsl"
	checkpointpbmsgs "github.com/filecoin-project/mir/pkg/pb/checkpointpb/msgs"
	checkpointpbtypes "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	cryptopbdsl "github.com/filecoin-project/mir/pkg/pb/cryptopb/dsl"
	cryptopbtypes "github.com/filecoin-project/mir/pkg/pb/cryptopb/types"
	eventpbdsl "github.com/filecoin-project/mir/pkg/pb/eventpb/dsl"
	eventpbtypes "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	hasherpbdsl "github.com/filecoin-project/mir/pkg/pb/hasherpb/dsl"
	hasherpbtypes "github.com/filecoin-project/mir/pkg/pb/hasherpb/types"
	transportpbevents "github.com/filecoin-project/mir/pkg/pb/transportpb/events"
	trantorpbdsl "github.com/filecoin-project/mir/pkg/pb/trantorpb/dsl"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	"github.com/filecoin-project/mir/pkg/timer/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"
)

const (
	DefaultResendPeriod = types.Duration(time.Second)
)

// NewModule allocates and returns a new instance of the ModuleParams associated with sequence number sn.
func NewModule(
	moduleConfig *common.ModuleConfig,
	ownID t.NodeID,
	membership *trantorpbtypes.Membership,
	epochConfig *trantorpbtypes.EpochConfig,
	leaderPolicyData []byte,
	resendPeriod types.Duration,
	logger logging.Logger,
) modules.PassiveModule {
	params := &common.ModuleParams{
		Logger:          logger,
		ModuleConfig:    moduleConfig,
		OwnID:           ownID,
		SeqNr:           epochConfig.FirstSn,
		Epoch:           epochConfig.EpochNr,
		ResendPeriod:    resendPeriod,
		Announced:       false,
		Signatures:      make(map[t.NodeID][]byte),
		Confirmations:   make(map[t.NodeID]struct{}),
		PendingMessages: make(map[t.NodeID]*checkpointpbtypes.Checkpoint),
		Membership:      maputil.GetSortedKeys(membership.Nodes),
		StateSnapshot: &trantorpbtypes.StateSnapshot{
			AppData: nil,
			EpochData: &trantorpbtypes.EpochData{
				EpochConfig:        epochConfig,
				ClientProgress:     nil, // This will be filled by a separate event.
				LeaderPolicy:       leaderPolicyData,
				PreviousMembership: membership,
			},
		},
	}

	m := dsl.NewModule(moduleConfig.Self)

	apppbdsl.UponSnapshot(m, func(appData []uint8) error {
		// Treat nil data as an empty byte slice.
		if appData == nil {
			appData = []byte{}
		}

		// Save the received app snapshot if there is none yet.
		if params.StateSnapshot.AppData == nil {
			params.StateSnapshot.AppData = appData
			if params.SnapshotReady() {
				processStateSnapshot(m, params)
			}
		}
		return nil
	})

	hasherpbdsl.UponResult(m, func(digests [][]uint8, _ *t.EmptyContext) error {
		// Save the received snapshot hash
		params.StateSnapshotHash = digests[0]

		// Request signature
		sigData := serializeCheckpointForSig(params.Epoch, params.SeqNr, params.StateSnapshotHash)

		cryptopbdsl.SignRequest(m, params.ModuleConfig.Crypto, sigData, &t.EmptyContext{})

		return nil
	})

	applyCheckpointReceived := func(from t.NodeID,
		epoch tt.EpochNr, sn tt.SeqNr, snapshotHash []uint8, signature []uint8) error {

		// check if from is part of the membership
		if !sliceutil.Contains(params.Membership, from) {
			params.Logger.Log(logging.LevelWarn, "sender %s is not a member.\n", from)
			return nil
		}
		// Notify the protocol about the progress of the from node.
		// If no progress is made for a configured number of epochs,
		// the node is considered to be a straggler and is sent a stable checkpoint to catch uparams.
		checkpointpbdsl.EpochProgress(m, params.ModuleConfig.Ord, from, epoch)

		// If checkpoint is already stable, ignore message.
		if params.Stable() {
			return nil
		}

		// Check snapshot hash
		if params.StateSnapshotHash == nil {
			// The message is received too early, put it aside
			params.PendingMessages[from] = (&checkpointpbtypes.Checkpoint{
				Epoch:        epoch,
				Sn:           sn,
				SnapshotHash: snapshotHash,
				Signature:    signature,
			})
			return nil
		} else if !bytes.Equal(params.StateSnapshotHash, snapshotHash) {
			// Snapshot hash mismatch
			params.Log(logging.LevelWarn, "Ignoring Checkpoint message. Mismatching state snapshot hash.", "from", from)
			return nil
		}

		// Ignore duplicate messages.
		if _, ok := params.Signatures[from]; ok {
			return nil
		}
		params.Signatures[from] = signature

		// Verify signature of the sender.
		sigData := serializeCheckpointForSig(params.Epoch, params.SeqNr, params.StateSnapshotHash)
		cryptopbdsl.VerifySigs(m,
			params.ModuleConfig.Crypto,
			[]*cryptopbtypes.SignedData{sigData},
			[][]byte{params.Signatures[from]},
			[]t.NodeID{from},
			&t.EmptyContext{},
		)

		return nil
	}

	cryptopbdsl.UponSignResult(m, func(sig []uint8, _ *t.EmptyContext) error {

		// Save received own checkpoint signature
		params.Signatures[params.OwnID] = sig
		params.Confirmations[params.OwnID] = struct{}{}

		// In case the node's own signature is enough to reach quorum, announce the stable checkpoint.
		// This can happen in a small system where no failures are tolerated.
		if params.Stable() {
			announceStable(m, params)
		}

		// Send a checkpoint message to all nodes after persisting checkpoint to the WAL.
		chkpMessage := checkpointpbmsgs.Checkpoint(params.ModuleConfig.Self, params.Epoch, params.SeqNr, params.StateSnapshotHash, sig)
		eventpbdsl.TimerRepeat(m,
			"timer",
			[]*eventpbtypes.Event{transportpbevents.SendMessage(params.ModuleConfig.Net, chkpMessage, params.Membership)},
			params.ResendPeriod,
			tt.RetentionIndex(params.Epoch),
		)

		params.Log(logging.LevelDebug, "Sending checkpoint message",
			"epoch", params.Epoch,
			"dataLen", len(params.StateSnapshot.AppData),
			"memberships", len(params.StateSnapshot.EpochData.EpochConfig.Memberships),
		)

		// Apply pending Checkpoint messages
		for s, msg := range params.PendingMessages {
			if err := applyCheckpointReceived(s, msg.Epoch, msg.Sn, msg.SnapshotHash, msg.Signature); err != nil {
				params.Log(logging.LevelWarn, "Error applying pending Checkpoint message", "error", err, "msg", msg)
			}

		}
		params.PendingMessages = nil

		return nil
	})

	cryptopbdsl.UponSigsVerified(m, func(nodeIds []t.NodeID, errors []error, allOk bool, _ *t.EmptyContext) error {

		// A checkpoint only has one signature and thus each slice of the result only contains one element.
		sourceNode := nodeIds[0]
		err := errors[0]

		if !allOk {
			params.Log(logging.LevelWarn, "Ignoring Checkpoint message. Invalid signature.", "source", sourceNode, "error", err)
			params.Signatures[sourceNode] = nil
			return nil
		}

		// Note the reception of a valid Checkpoint message from node `source`.
		params.Confirmations[sourceNode] = struct{}{}

		// If, after having applied this message, the checkpoint became stable, produce the necessary events.
		if params.Stable() {
			announceStable(m, params)
		}

		return nil
	})

	trantorpbdsl.UponClientProgress(m, func(progress map[tt.ClientID]*trantorpbtypes.DeliveredTXs) error {
		// Save the received client progress if there is none yet.
		if params.StateSnapshot.EpochData.ClientProgress == nil {
			params.StateSnapshot.EpochData.ClientProgress = &trantorpbtypes.ClientProgress{
				Progress: progress,
			}
			if params.SnapshotReady() {
				processStateSnapshot(m, params)
			}
		}
		return nil
	})

	checkpointpbdsl.UponCheckpointReceived(m, applyCheckpointReceived)

	return m
}

func processStateSnapshot(m dsl.Module, p *common.ModuleParams) {

	// Initiate computing the hash of the snapshot.
	hasherpbdsl.Request(m,
		p.ModuleConfig.Hasher,
		[]*hasherpbtypes.HashData{serializeSnapshotForHash(p.StateSnapshot)},
		&t.EmptyContext{},
	)
}

func announceStable(m dsl.Module, p *common.ModuleParams) {

	// Only announce the stable checkpoint once.
	if p.Announced {
		return
	}
	p.Announced = true

	// Assemble a multisig certificate from the received signatures.
	cert := make(map[t.NodeID][]byte)
	for node := range p.Confirmations {
		cert[node] = p.Signatures[node]
	}

	// Announce the stable checkpoint to the ordering protocol.
	checkpointpbdsl.StableCheckpoint(m, p.ModuleConfig.Ord, p.SeqNr, p.StateSnapshot, cert)
}
