/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package deploytest

import (
	"fmt"

	"github.com/filecoin-project/mir/pkg/checkpoint"
	"github.com/filecoin-project/mir/pkg/pb/trantorpb"
	"github.com/filecoin-project/mir/pkg/serializing"
	t "github.com/filecoin-project/mir/pkg/types"
)

// FakeApp represents a dummy stub application used for testing only.
type FakeApp struct {
	ProtocolModule t.ModuleID

	// The state of the FakeApp only consists of a counter of processed transactions.
	TransactionsProcessed uint64
}

func (fa *FakeApp) ApplyTXs(txs []*trantorpb.Transaction) error {
	for _, req := range txs {
		fa.TransactionsProcessed++
		fmt.Printf("Received request: %q. Processed transactions: %d\n", string(req.Data), fa.TransactionsProcessed)
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
