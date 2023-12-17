/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package deploytest

import (
	"fmt"

	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	t "github.com/filecoin-project/mir/stdtypes"

	"github.com/filecoin-project/mir/pkg/checkpoint"
	"github.com/filecoin-project/mir/pkg/serializing"
)

// FakeApp represents a dummy stub application used for testing only.
type FakeApp struct {
	ProtocolModule t.ModuleID

	// The state of the FakeApp only consists of a counter of processed transactions.
	TransactionsProcessed uint64
}

func (fa *FakeApp) ApplyTXs(txs []*trantorpbtypes.Transaction) error {
	for _, tx := range txs {
		fa.TransactionsProcessed++
		fmt.Printf("Received transaction: %q. Processed transactions: %d\n", string(tx.Data), fa.TransactionsProcessed)
	}
	return nil
}

func (fa *FakeApp) Snapshot() ([]byte, error) {
	return serializing.Uint64ToBytes(fa.TransactionsProcessed), nil
}

func (fa *FakeApp) RestoreState(checkpoint *checkpoint.StableCheckpoint) error {
	fa.TransactionsProcessed = serializing.Uint64FromBytes(checkpoint.Snapshot.AppData)
	return nil
}

func (fa *FakeApp) Checkpoint(_ *checkpoint.StableCheckpoint) error {
	return nil
}

func NewFakeApp() *FakeApp {
	return &FakeApp{
		TransactionsProcessed: 0,
	}
}
