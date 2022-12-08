/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package deploytest

import (
	"encoding/binary"
	"fmt"

	"github.com/filecoin-project/mir/pkg/checkpoint"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// FakeApp represents a dummy stub application used for testing only.
type FakeApp struct {
	ProtocolModule t.ModuleID

	Membership map[t.NodeID]t.NodeAddress

	// The state of the FakeApp only consists of a counter of processed requests.
	RequestsProcessed uint64
}

func (fa *FakeApp) ApplyTXs(txs []*requestpb.Request) error {
	for _, req := range txs {
		fa.RequestsProcessed++
		fmt.Printf("Received request: %q. Processed requests: %d\n", string(req.Data), fa.RequestsProcessed)
	}
	return nil
}

func (fa *FakeApp) Snapshot() ([]byte, error) {
	return uint64ToBytes(fa.RequestsProcessed), nil
}

func (fa *FakeApp) RestoreState(checkpoint *checkpoint.StableCheckpoint) error {
	fa.RequestsProcessed = uint64FromBytes(checkpoint.Snapshot.AppData)
	return nil
}

func (fa *FakeApp) Checkpoint(_ *checkpoint.StableCheckpoint) error {
	return nil
}

func NewFakeApp() *FakeApp {
	return &FakeApp{
		RequestsProcessed: 0,
	}
}

func uint64ToBytes(value uint64) []byte {
	byteValue := make([]byte, 8)
	binary.LittleEndian.PutUint64(byteValue, value)
	return byteValue
}

func uint64FromBytes(bytes []byte) uint64 {
	return binary.LittleEndian.Uint64(bytes)
}
