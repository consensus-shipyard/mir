package emptybatchid

import (
	"reflect"

	msctypes "github.com/filecoin-project/mir/pkg/availability/multisigcollector/types"
)

func EmptyBatchID() []byte {
	return []byte{0}
}

func IsEmptyBatchID(batchID msctypes.BatchID) bool {
	return reflect.DeepEqual(batchID, EmptyBatchID())
}
