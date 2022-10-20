// Copyright Contributors to the Mir project
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/filecoin-project/mir/pkg/checkpoint"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/commonpb"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
)

type App struct {
	logging.Logger

	Membership map[t.NodeID]t.NodeAddress
}

func (a *App) ApplyTXs(txs []*requestpb.Request) error {
	for _, req := range txs {
		a.Log(logging.LevelDebug, fmt.Sprintf("Delivered request %v from client %v", req.ReqNo, req.ClientId))
	}
	return nil
}

func (a *App) NewEpoch(_ t.EpochNr) (map[t.NodeID]t.NodeAddress, error) {
	return maputil.Copy(a.Membership), nil
}

func (a *App) Snapshot() ([]byte, error) {
	return nil, nil
}

func (a *App) RestoreState(appData []byte, epochConfig *commonpb.EpochConfig) error {
	return nil
}

func (a *App) Checkpoint(_ *checkpoint.StableCheckpoint) error {
	return nil
}
