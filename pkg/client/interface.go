package client

import (
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// The Client represents an SMR client that produces new transactions (using NewTX)
// that can be submitted to the SMR system.
// It sequentially assigns transaction numbers to all created transactions.
// The Client is informed (using Done)
// when a transaction is considered applied (or otherwise safe to stop caring about).
// All created transactions that have not been confirmed this way are considered pending.
// The client is expected to persistently store all pending transactions,
// in case they need to be resubmitted after recovering from a potential crash.
type Client interface {

	// NewTX creates and returns a new transaction of type txType, containing data.
	// The returned transaction will be assigned the next transaction number in sequence,
	// corresponding to the number of transactions previously created by this client.
	// Until Done is called with the returned transaction's number,
	// the transaction will be pending, i.e., among the transactions returned by Pending.
	// The transaction need not necessarily yet be written to persistent storage when NewTX returns.
	// Use the separate Sync() method to guarantee persistence.
	NewTX(txType uint64, data []byte) (*requestpb.Request, error)

	// Done marks a transaction as done. It will no longer be among the transactions returned by Pending.
	// The effect of this call need not be written to persistent storage until Sync is called.
	Done(txNo t.ReqNo) error

	// Pending returns all transactions previously returned by NewTX that have not been marked as done.
	Pending() ([]*requestpb.Request, error)

	// Sync ensures that the effects of all previous calls to NewTX and Done have been written to persistent storage.
	Sync() error
}
