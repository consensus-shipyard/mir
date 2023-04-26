package client

import (
	"github.com/filecoin-project/mir/pkg/pb/trantorpb"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
)

// VolatileClient is a simple implementation of the SMR Client interface that does not provide any persistence.
// It is meant for testing purposes and for cases where the machine running it is assumed to never crash.
type VolatileClient struct {
	clientID  tt.ClientID
	nextTxNo  tt.TxNo
	pendingTX map[tt.TxNo]*trantorpb.Transaction
}

// NewVolatileClient returns a new instance of VolatileClient.
// All transaction it produces will be identified with the given clientID.
func NewVolatileClient(clientID tt.ClientID) *VolatileClient {
	return &VolatileClient{
		clientID:  clientID,
		nextTxNo:  0,
		pendingTX: make(map[tt.TxNo]*trantorpb.Transaction),
	}
}

// NewTX creates and returns a new transaction of type txType, containing data.
// The returned transaction will be assigned the next transaction number in sequence,
// corresponding to the number of transactions previously created by this client.
// Until Done is called with the returned transaction's number,
// the transaction will be pending, i.e., among the transactions returned by Pending.
func (vc *VolatileClient) NewTX(txType uint64, data []byte) (*trantorpb.Transaction, error) {
	tx := &trantorpb.Transaction{
		ClientId: vc.clientID.Pb(),
		TxNo:     vc.nextTxNo.Pb(),
		Type:     txType,
		Data:     data,
	}
	vc.pendingTX[vc.nextTxNo] = tx
	vc.nextTxNo++
	return tx, nil
}

// Done marks a transaction as done. It will no longer be among the transactions returned by Pending.
func (vc *VolatileClient) Done(txNo tt.TxNo) error {
	delete(vc.pendingTX, txNo)
	return nil
}

// Pending returns all transactions previously returned by NewTX that have not been marked as done.
func (vc *VolatileClient) Pending() ([]*trantorpb.Transaction, error) {
	return maputil.GetValuesOf(vc.pendingTX, maputil.GetSortedKeys(vc.pendingTX)), nil
}

// Sync does nothing.
// VolatileClient does not provide any persistence.
func (vc *VolatileClient) Sync() error {
	return nil
}
