package iss

import (
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

// The CommitLogEntry type represents an entry of the commit log, the final output of the ordering process.
// Whenever an orderer delivers an availability certificate (or a special abort value),
// it is inserted to the commit log in form of a commitLogEntry.
type CommitLogEntry struct {
	// Sequence number at which this entry has been ordered.
	Sn tt.SeqNr

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
