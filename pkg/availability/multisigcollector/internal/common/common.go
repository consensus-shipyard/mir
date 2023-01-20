package common

import (
	mscpbtypes "github.com/filecoin-project/mir/pkg/pb/availabilitypb/mscpb/types"
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
	Self    t.ModuleID // id of this module
	Mempool t.ModuleID
	BatchDB t.ModuleID
	Net     t.ModuleID
	Crypto  t.ModuleID
}

// ModuleParams sets the values for the parameters of an instance of the protocol.
// All replicas are expected to use identical module parameters.
type ModuleParams struct {
	InstanceUID []byte             // unique identifier for this instance of BCB, used to prevent cross-instance replay attacks
	AllNodes    []t.NodeID         // the list of participating nodes
	F           int                // the maximum number of failures tolerated. Must be less than (len(AllNodes)-1) / 2
	Certs       []*mscpbtypes.Cert // the list of generated certificates
	Limit       int                // the maximum number of certificates to generate before a request is completed
	MaxRequests int                // the maximum number of requests to be provided by this module
}

// SigData is the binary data that should be signed for forming a certificate.
func SigData(instanceUID InstanceUID, batchID t.BatchID) [][]byte {
	return [][]byte{instanceUID.Bytes(), []byte("BATCH_STORED"), batchID}
}
