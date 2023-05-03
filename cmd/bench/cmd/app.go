// Copyright Contributors to the Mir project
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/filecoin-project/mir/pkg/checkpoint"
	"github.com/filecoin-project/mir/pkg/logging"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
)

type App struct {
	logging.Logger

	Membership *trantorpbtypes.Membership
}

func (a *App) ApplyTXs(txs []*trantorpbtypes.Transaction) error {
	for _, tx := range txs {
		a.Log(logging.LevelDebug, fmt.Sprintf("Delivered transaction %v from client %v", tx.TxNo, tx.ClientId))
	}
	return nil
}

func (a *App) NewEpoch(_ tt.EpochNr) (*trantorpbtypes.Membership, error) {
	return a.Membership, nil
}

func (a *App) Snapshot() ([]byte, error) {
	return nil, nil
}

func (a *App) RestoreState(_ *checkpoint.StableCheckpoint) error {
	return nil
}

func (a *App) Checkpoint(_ *checkpoint.StableCheckpoint) error {
	return nil
}
