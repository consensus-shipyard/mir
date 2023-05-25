package common

import (
	msctypes "github.com/filecoin-project/mir/pkg/availability/multisigcollector/types"
	mscpbtypes "github.com/filecoin-project/mir/pkg/pb/availabilitypb/mscpb/types"
	cryptopbtypes "github.com/filecoin-project/mir/pkg/pb/cryptopb/types"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
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
	InstanceUID []byte                     // unique identifier for this instance used to prevent replay attacks
	Membership  *trantorpbtypes.Membership // the list of participating nodes
	Certs       []*mscpbtypes.Cert         // the list of generated certificates
	Limit       int                        // the maximum number of certificates to generate before a request is completed
	MaxRequests int                        // the maximum number of requests to be provided by this module
}

// SigData is the binary data that should be signed for forming a certificate.
func SigData(instanceUID InstanceUID, batchID msctypes.BatchID) *cryptopbtypes.SignedData {
	return &cryptopbtypes.SignedData{Data: [][]byte{instanceUID.Bytes(), []byte("BATCH_STORED"), []byte(batchID)}}
}
