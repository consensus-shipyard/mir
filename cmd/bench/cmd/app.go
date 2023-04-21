// Copyright Contributors to the Mir project
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/filecoin-project/mir/pkg/checkpoint"
	"github.com/filecoin-project/mir/pkg/logging"
	commonpbtypes "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
)

type App struct {
	logging.Logger

	Membership *commonpbtypes.Membership
}

func (a *App) ApplyTXs(txs []*requestpb.Request) error {
	for _, req := range txs {
		a.Log(logging.LevelDebug, fmt.Sprintf("Delivered request %v from client %v", req.ReqNo, req.ClientId))
	}
	return nil
}

func (a *App) NewEpoch(_ tt.EpochNr) (*commonpbtypes.Membership, error) {
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
