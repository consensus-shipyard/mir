package checkpoint

import (
	"reflect"

	"github.com/fxamacker/cbor/v2"
	es "github.com/go-errors/errors"
	"google.golang.org/protobuf/proto"

	"github.com/filecoin-project/mir/pkg/clientprogress"
	"github.com/filecoin-project/mir/pkg/crypto"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/checkpointpb"
	checkpointpbtypes "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
	"github.com/filecoin-project/mir/pkg/util/membutil"
)

// StableCheckpoint represents a stable checkpoint.
type StableCheckpoint checkpointpbtypes.StableCheckpoint

// StableCheckpointFromPb creates a new StableCheckpoint from its protobuf representation.
// The given protobuf object is assumed to not be modified after calling StableCheckpointFromPb.
// Modifying it may lead to undefined behavior.
func StableCheckpointFromPb(checkpoint *checkpointpb.StableCheckpoint) *StableCheckpoint {
	return (*StableCheckpoint)(checkpointpbtypes.StableCheckpointFromPb(checkpoint))
}

// Pb returns a protobuf representation of the stable checkpoint.
func (sc *StableCheckpoint) Pb() *checkpointpb.StableCheckpoint {
	return (*checkpointpbtypes.StableCheckpoint)(sc).Pb()
}

// Serialize returns the stable checkpoint serialized as a byte slice.
// It is the inverse of Deserialize, to which the returned byte slice can be passed to restore the checkpoint.
func (sc *StableCheckpoint) Serialize() ([]byte, error) {
	em, err := cbor.CoreDetEncOptions().EncMode()
	if err != nil {
		return nil, err
	}

	// TODO: This is a DIRTY DIRTY hack. It works around the problem of protocol buffers not distinguishing between
	//   an empty slice and a nil value (both get serialized to the same bytes and deserialized to nil).
	//   CBOR, however does make this distinction. The checkpoint data structure does contain slices that may be empty
	//   (DeliveredTXs inside ClientProgress). Thus, if a checkpoint were directly serialized locally,
	//   and the very same checkpoint transmitted to another node
	//   (which implies protobuf marshalling and unmarshalling) and serialized there, the two outputs (local and remote)
	//   would not be the same (an empty slice locally and nil on the remote node).
	//   To keep the serialization consistent, we always perform a protobuf marshalling-unmarshalling of the checkpoint.
	//   Note that we cannot use the protobuf serialization directly either, as it is not deterministic.
	//   This is very very inefficient and something should be done about it.
	pbData, err := proto.Marshal(sc.Pb())
	if err != nil {
		return nil, err
	}
	var pbChkp checkpointpb.StableCheckpoint
	err = proto.Unmarshal(pbData, &pbChkp)
	if err != nil {
		return nil, err
	}

	return em.Marshal(&pbChkp)
}

// Deserialize populates its fields from the serialized representation
// previously returned from StableCheckpoint.Serialize.
func (sc *StableCheckpoint) Deserialize(data []byte) error {

	// TODO: See comment in the serialization function for an explanation of
	//   why we serialize the protobuf and not the checkpoint object directly.
	var pbChkp checkpointpb.StableCheckpoint
	if err := cbor.Unmarshal(data, &pbChkp); err != nil {
		return es.Errorf("failed to CBOR unmarshal stable checkpoint: %w", err)
	}
	*sc = StableCheckpoint(*checkpointpbtypes.StableCheckpointFromPb(&pbChkp))
	return nil
}

// StripCert returns a stable new stable checkpoint with the certificate stripped off.
// The returned copy is a shallow one, sharing the data with the original.
func (sc *StableCheckpoint) StripCert() *StableCheckpoint {
	return (*StableCheckpoint)(&checkpointpbtypes.StableCheckpoint{
		Sn:       sc.Sn,
		Snapshot: sc.Snapshot,
		Cert:     nil,
	})
}

// AttachCert returns a new stable checkpoint with the given certificate attached.
// If the stable checkpoint already had a certificate attached, the old certificate is replaced by the new one.
func (sc *StableCheckpoint) AttachCert(cert *Certificate) *StableCheckpoint {
	return (*StableCheckpoint)(&checkpointpbtypes.StableCheckpoint{
		Sn:       sc.Sn,
		Snapshot: sc.Snapshot,
		Cert:     *cert,
	})
}

// SeqNr returns the sequence number of the stable checkpoint.
// It is defined as the number of sequence numbers comprised in the checkpoint, or, in other words,
// the first (i.e., lowest) sequence number not included in the checkpoint.
func (sc *StableCheckpoint) SeqNr() tt.SeqNr {
	return sc.Sn
}

// Memberships returns the memberships configured for the epoch of this checkpoint
// and potentially several subsequent ones.
func (sc *StableCheckpoint) Memberships() []*trantorpbtypes.Membership {
	return sc.Snapshot.EpochData.EpochConfig.Memberships
}

// PreviousMembership returns the membership of the epoch preceding the epoch the checkpoint is associated with
// (i.e. the membership of sc.Epoch()-1).
// This is the membership that created the checkpoint and must be used to verify its certificate.
// Note that this membership is contained in the checkpoint itself and thus can be forged.
// Using PreviousMembership as an argument to VerifyCert without independently checking its validity is not secure
// (in this sense, the checkpoint certificate is self-signed).
func (sc *StableCheckpoint) PreviousMembership() *trantorpbtypes.Membership {
	return sc.Snapshot.EpochData.PreviousMembership
}

// Epoch returns the epoch associated with this checkpoint.
// It is the epoch **started** by this checkpoint, **not** the last one included in it.
func (sc *StableCheckpoint) Epoch() tt.EpochNr {
	return sc.Snapshot.EpochData.EpochConfig.EpochNr
}

// StateSnapshot returns the serialized application state and system configuration associated with this checkpoint.
func (sc *StableCheckpoint) StateSnapshot() *trantorpbtypes.StateSnapshot {
	return sc.Snapshot
}

func (sc *StableCheckpoint) ClientProgress(logger logging.Logger) *clientprogress.ClientProgress {
	return clientprogress.FromPb(sc.Snapshot.EpochData.ClientProgress.Pb(), logger)
}

func (sc *StableCheckpoint) Certificate() Certificate {
	return sc.Cert
}

// VerifyCert verifies the certificate of the stable checkpoint using the provided hash implementation and verifier.
// The same (or corresponding) modules must have been used when the certificate was created by the checkpoint module.
// The has implementation is a crypto.HashImpl used to create a Mir hasher module and the verifier interface
// is a subset of the crypto.Crypto interface (narrowed down to only the Verify function).
// Thus, the same (or equivalent) crypto implementation that was used to create checkpoint
// can be used as a Verifier to verify it.
//
// Note that VerifyCert performs all the necessary hashing and signature verifications synchronously
// (only returns when the signature is verified). This may become a very computationally expensive operation.
// It is thus recommended not to use this function directly within a sequential protocol implementation,
// and rather delegating the hashing and signature verification tasks
// to dedicated modules using the corresponding events.
// Also, in case the verifier implementation is used by other goroutines,
// make sure that calling Vetify on it is thread-safe.
//
// For simplicity, we require all nodes that signed the certificate to be contained in the provided membership,
// as well as all signatures to be valid.
// Moreover, the number of nodes that signed the certificate must be greater than one third of the membership size.
func (sc *StableCheckpoint) VerifyCert(h crypto.HashImpl, v Verifier, membership *trantorpbtypes.Membership) error {

	// Check if there is enough signatures.
	if !membutil.HaveWeakQuorum(membership, maputil.GetKeys(sc.Cert)) {
		return es.New("not enough signatures in certificate to have weak quorum")
	}

	// Check whether all signatures are valid.
	snapshotData, err := serializeSnapshotForHash(sc.StateSnapshot())
	if err != nil {
		return es.Errorf("failed serializing snapshot data: %w", err)
	}
	snapshotHash := hash(snapshotData.Data, h)
	signedData := serializeCheckpointForSig(sc.Epoch(), sc.SeqNr(), snapshotHash)
	for nodeID, sig := range sc.Cert {
		// For each signature in the certificate...

		// Check if the signing node is also in the given membership, thus "authorized" to sign.
		// TODO: Once nodes are identified by more than their ID
		//   (e.g., if a separate putlic key is part of their identity), adapt the check accordingly.
		if _, ok := membership.Nodes[nodeID]; !ok {
			return es.Errorf("node %v not in membership", nodeID)
		}

		// Check if the signature is valid.
		if err := v.Verify(signedData.Data, sig, nodeID); err != nil {
			return es.Errorf("signature verification error (node %v): %w", nodeID, err)
		}
	}
	return nil
}

// Verify makes the necessary checks to verify a checkpoint
func (sc *StableCheckpoint) Verify(
	configOffset int,
	hashImpl crypto.HashImpl,
	chkpVerifier Verifier,
	membership *trantorpbtypes.Membership,
) error {

	if err := sc.SyntacticCheck(configOffset); err != nil {
		return es.Errorf("invalid checkpoint: %w", err)
	}

	// Only verify certificate if not the first epoch
	if sc.Epoch() > 0 {
		if err := sc.VerifyCert(hashImpl, chkpVerifier, membership); err != nil {
			return es.Errorf("invalid checkpoint: %w", err)
		}
	} else if len(sc.PreviousMembership().Nodes) > 0 {
		return es.Errorf("invalid checkpoint: certificate not empty for first epoch")
	}

	// Verify that memberships are consistent with each other
	if err := sc.verifyMembershipConsistency(membership); err != nil {
		return es.Errorf("invalid checkpoint memberships: %w", err)
	}

	return nil
}

// verifyMembershipConsistency verifies that if the same node appears then the same parameters are always used
func (sc *StableCheckpoint) verifyMembershipConsistency(membership *trantorpbtypes.Membership) error {
	membershipConsistency := map[t.NodeID]*trantorpbtypes.NodeIdentity{}
	for _, chkpMembership := range sc.Memberships() {
		for nodeID, node := range chkpMembership.Nodes {
			if _, ok := membershipConsistency[nodeID]; !ok {
				membershipConsistency[nodeID] = node
			} else {
				// check that all parameters are consistent
				if !reflect.DeepEqual(membershipConsistency[nodeID], node) {
					return es.Errorf("inconsistent membership parameters for node %v", nodeID)
				}
			}
		}
	}

	if sc.Epoch() > 0 && !reflect.DeepEqual(sc.PreviousMembership(), membership) {
		return es.Errorf("inconsistent sc.PreviousMembership() with membership")
	}

	return nil
}

// Genesis returns a stable checkpoint that serves as the starting checkpoint of the first epoch (epoch 0).
// Its certificate is empty and is always considered valid,
// as there is no previous epoch's membership to verify it against.
func Genesis(initialStateSnapshot *trantorpbtypes.StateSnapshot) *StableCheckpoint {
	return &StableCheckpoint{
		Sn:       0,
		Snapshot: initialStateSnapshot,
		Cert:     Certificate{},
	}
}

// The Verifier interface represents a subset of the crypto.Crypto interface
// that can be used for verifying stable checkpoint certificates.
type Verifier interface {
	// Verify verifies a signature produced by the node with ID nodeID over data.
	// Returns nil on success (i.e., if the given signature is valid) and a non-nil error otherwise.
	Verify(data [][]byte, signature []byte, nodeID t.NodeID) error
}

func hash(data [][]byte, hasher crypto.HashImpl) []byte {
	h := hasher.New()
	for _, d := range data {
		h.Write(d)
	}
	return h.Sum(nil)
}

// SyntacticCheck checks whether the stable checkpoint is well-formed.
func (sc *StableCheckpoint) SyntacticCheck(
	configOffset int,
) error {

	if sc.StateSnapshot() == nil {
		return es.Errorf("snapshot is nil")
	}

	if sc.StateSnapshot().EpochData == nil {
		return es.Errorf("epoch data is nil")
	}

	if sc.StateSnapshot().EpochData.EpochConfig == nil {
		return es.Errorf("epoch config is nil")
	}

	if sc.Memberships() == nil {
		return es.Errorf("memberships is nil")
	}

	// Check if checkpoint contains the configured number of configurations.
	if len(sc.Memberships()) != configOffset+1 {
		return es.Errorf("invalid checkpoint: number of memberships %v does not match expected %v",
			configOffset+1,
			len(sc.Memberships()))
	}

	// Check the contained memberships for validity.
	for _, membership := range sc.Memberships() {
		if err := membutil.Valid(membership); err != nil {
			return es.Errorf("invalid membership: %w", err)
		}
	}

	if sc.PreviousMembership() == nil {
		return es.Errorf("previous membership is nil")
	}

	if sc.Epoch() > 0 {
		if err := membutil.Valid(sc.PreviousMembership()); err != nil {
			return es.Errorf("invalid previous membership: %w", err)
		}
	}

	if sc.Certificate() == nil {
		return es.Errorf("certificate is nil")
	}

	if sc.StateSnapshot().AppData == nil {
		return es.Errorf("app data is nil")
	}

	if sc.StateSnapshot().EpochData.ClientProgress == nil {
		return es.Errorf("client progress is nil")
	}

	if sc.StateSnapshot().EpochData.LeaderPolicy == nil {
		return es.Errorf("leader policy is nil")
	}

	return nil
}
