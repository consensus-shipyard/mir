package emptybatchid

import (
	"reflect"

	t "github.com/filecoin-project/mir/pkg/types"
)

func EmptyBatchID() []byte {
	return []byte{0}
}

func IsEmptyBatchID(batchID t.BatchID) bool {
	return reflect.DeepEqual(batchID, EmptyBatchID())
}
