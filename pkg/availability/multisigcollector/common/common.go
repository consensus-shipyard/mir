package common

import (
	msctypes "github.com/filecoin-project/mir/pkg/availability/multisigcollector/types"
	cryptopbtypes "github.com/filecoin-project/mir/pkg/pb/cryptopb/types"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

// InstanceUID is used to uniquely identify an instance of multisig collector.
// It is used to prevent cross-instance signature replay attack and should be unique across all executions.
type InstanceUID []byte

// Bytes returns the binary representation of the InstanceUID.
func (uid InstanceUID) Bytes() []byte {
	return uid
}

// ModuleConfig sets the module ids. All replicas are expected to use identical module configurations.
type ModuleConfig struct {
	Self t.ModuleID // id of this module

	BatchDB t.ModuleID
	Crypto  t.ModuleID
	Mempool t.ModuleID
	Net     t.ModuleID
}

// ModuleParams sets the values for the parameters of an instance of the protocol.
// All replicas are expected to use identical module parameters.
type ModuleParams struct {

	// InstanceUID is a unique identifier for this instance used to prevent replay attacks.
	InstanceUID []byte

	// Membership defines the set of nodes participating in a particular instance of the multisig collector.
	Membership *trantorpbtypes.Membership

	// Limit is the maximum number of certificates to generate before a request is completed.
	Limit int

	// MaxRequests is the maximum number of certificate requests to be handled by this module.
	// It prevents the multisig collector from continuing operation (transaction dissemination)
	// after no more certificate requests are going to arrive.
	MaxRequests int

	// EpochNr is the epoch in which the instance of MultisigCollector operates.
	// It is used as the RetentionIndex to associate with all newly stored batches.
	// and batches requested from the mempool.
	EpochNr tt.EpochNr
}

// SigData is the binary data that should be signed for forming a certificate.
func SigData(instanceUID InstanceUID, batchID msctypes.BatchID) *cryptopbtypes.SignedData {
	return &cryptopbtypes.SignedData{Data: [][]byte{instanceUID.Bytes(), []byte("BATCH_STORED"), []byte(batchID)}}
}
