package multisigcollector

import (
	"github.com/filecoin-project/mir/pkg/availability/multisigcollector/internal/common"
	"github.com/filecoin-project/mir/pkg/availability/multisigcollector/internal/parts/batchreconstruction"
	"github.com/filecoin-project/mir/pkg/availability/multisigcollector/internal/parts/certcreation"
	"github.com/filecoin-project/mir/pkg/availability/multisigcollector/internal/parts/certverification"
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/modules"
	t "github.com/filecoin-project/mir/pkg/types"
)

type ModuleConfig = common.ModuleConfig

type ModuleParams = common.ModuleParams

// NewModule creates a new instance of the multisig collector module.
// Multisig collector is the simplest implementation of the availability layer.
// Whenever an availability certificate is requested, it pulls a batch from the mempool module,
// sends it to all replicas and collects a quorum (i.e., more than (N+F)/2) of signatures confirming that
// other nodes have persistently stored the batch.
func NewModule(mc *ModuleConfig, params *ModuleParams, nodeID t.NodeID) modules.PassiveModule {
	m := dsl.NewModule(mc.Self)

	commonState := &common.State{
		BatchStore:       make(map[t.BatchID][]t.TxID),
		TransactionStore: make(map[t.TxID][]byte),
	}

	certcreation.IncludeCreatingCertificates(m, mc, params, nodeID, commonState)
	certverification.IncludeVerificationOfCertificates(m, mc, params, nodeID, commonState)
	batchreconstruction.IncludeBatchReconstruction(m, mc, params, nodeID, commonState)

	return m
}
