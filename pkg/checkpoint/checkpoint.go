package checkpoint

import (
	"fmt"

	"github.com/fxamacker/cbor/v2"

	"github.com/filecoin-project/mir/pkg/clientprogress"
	"github.com/filecoin-project/mir/pkg/crypto"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/checkpointpb"
	"github.com/filecoin-project/mir/pkg/pb/commonpb"
	"github.com/filecoin-project/mir/pkg/serializing"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
)

// StableCheckpoint represents a stable checkpoint.
type StableCheckpoint checkpointpb.StableCheckpoint

// StableCheckpointFromPb creates a new StableCheckpoint from its protobuf representation.
// The given protobuf object is assumed to not be modified after calling StableCheckpointFromPb.
// Modifying it may lead to undefined behavior.
func StableCheckpointFromPb(checkpoint *checkpointpb.StableCheckpoint) *StableCheckpoint {
	return (*StableCheckpoint)(checkpoint)
}

// Pb returns a protobuf representation of the stable checkpoint.
func (sc *StableCheckpoint) Pb() *checkpointpb.StableCheckpoint {
	return (*checkpointpb.StableCheckpoint)(sc)
}

// Serialize returns the stable checkpoint serialized as a byte slice.
// It is the inverse of Deserialize, to which the returned byte slice can be passed to restore the checkpoint.
func (sc *StableCheckpoint) Serialize() ([]byte, error) {
	em, err := cbor.CoreDetEncOptions().EncMode()
	if err != nil {
		return nil, err
	}

	return em.Marshal(sc)
}

// Deserialize populates its fields from the serialized representation
// previously returned from StableCheckpoint.Serialize.
func (sc *StableCheckpoint) Deserialize(data []byte) error {
	if err := cbor.Unmarshal(data, sc); err != nil {
		return fmt.Errorf("failed to CBOR unmarshal stable checkpoint: %w", err)
	}
	return nil
}

// StripCert returns a stable new stable checkpoint with the certificate stripped off.
// The returned copy is a shallow one, sharing the data with the original.
func (sc *StableCheckpoint) StripCert() *StableCheckpoint {
	return StableCheckpointFromPb(&checkpointpb.StableCheckpoint{
		Sn:       sc.Sn,
		Snapshot: sc.Snapshot,
		Cert:     nil,
	})
}

// AttachCert returns a new stable checkpoint with the given certificate attached.
// If the stable checkpoint already had a certificate attached, the old certificate is replaced by the new one.
func (sc *StableCheckpoint) AttachCert(cert *Certificate) *StableCheckpoint {
	return StableCheckpointFromPb(&checkpointpb.StableCheckpoint{
		Sn:       sc.Sn,
		Snapshot: sc.Snapshot,
		Cert:     cert.Pb(),
	})
}

// SeqNr returns the sequence number of the stable checkpoint.
// It is defined as the number of sequence numbers comprised in the checkpoint, or, in other words,
// the first (i.e., lowest) sequence number not included in the checkpoint.
func (sc *StableCheckpoint) SeqNr() t.SeqNr {
	return t.SeqNr(sc.Sn)
}

// Memberships returns the memberships configured for the epoch of this checkpoint
// and potentially several subsequent ones.
func (sc *StableCheckpoint) Memberships() []map[t.NodeID]t.NodeAddress {
	return t.MembershipSlice(sc.Snapshot.EpochData.EpochConfig.Memberships)
}

// Epoch returns the epoch associated with this checkpoint.
// It is the epoch **started** by this checkpoint, **not** the last one included in it.
func (sc *StableCheckpoint) Epoch() t.EpochNr {
	return t.EpochNr(sc.Snapshot.EpochData.EpochConfig.EpochNr)
}

// StateSnapshot returns the serialized application state and system configuration associated with this checkpoint.
func (sc *StableCheckpoint) StateSnapshot() *commonpb.StateSnapshot {
	return sc.Snapshot
}

func (sc *StableCheckpoint) ClientProgress(logger logging.Logger) *clientprogress.ClientProgress {
	return clientprogress.FromPb(sc.Snapshot.EpochData.ClientProgress, logger)
}

func (sc *StableCheckpoint) Certificate() *Certificate {
	m := maputil.Transform(sc.Cert,
		func(k string) t.NodeID {
			return t.NodeID(k)
		}, func(v []byte) []byte {
			return v
		},
	)
	return (*Certificate)(&m)
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
func (sc *StableCheckpoint) VerifyCert(h crypto.HashImpl, v Verifier, membership map[t.NodeID]t.NodeAddress) error {

	// Check if there is enough signatures.
	n := len(membership)
	f := (n - 1) / 3
	if len(sc.Cert) < f+1 {
		return fmt.Errorf("not enough signatures in certificate: got %d, expected more than %d",
			len(sc.Cert), f+1)
	}

	// Check whether all signatures are valid.
	snapshotData := serializing.SnapshotForHash(sc.StateSnapshot())
	snapshotHash := hash(snapshotData, h)
	sigData := serializing.CheckpointForSig(sc.Epoch(), sc.SeqNr(), snapshotHash)
	for nodeID, sig := range sc.Cert {
		// For each signature in the certificate...

		// Check if the signing node is also in the given membership, thus "authorized" to sign.
		// TODO: Once nodes are identified by more than their ID
		//   (e.g., if a separate putlic key is part of their identity), adapt the check accordingly.
		if _, ok := membership[t.NodeID(nodeID)]; !ok {
			return fmt.Errorf("node %v not in membership", nodeID)
		}

		// Check if the signature is valid.
		if err := v.Verify(sigData, sig, t.NodeID(nodeID)); err != nil {
			return fmt.Errorf("signature verification error (node %v): %w", nodeID, err)
		}
	}
	return nil
}

// Genesis returns a stable checkpoint that serves as the starting checkpoint of the first epoch (epoch 0).
// Its certificate is empty and is always considered valid,
// as there is no previous epoch's membership to verify it against.
func Genesis(initialStateSnapshot *commonpb.StateSnapshot) *StableCheckpoint {
	return &StableCheckpoint{
		Sn:       0,
		Snapshot: initialStateSnapshot,
		Cert:     (&Certificate{}).Pb(),
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
