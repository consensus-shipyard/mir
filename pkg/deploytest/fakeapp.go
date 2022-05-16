/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package deploytest

import (
	"encoding/binary"
	"fmt"

	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
)

// FakeApp represents a dummy stub application used for testing only.
type FakeApp struct {

	// Request store maintained by the FakeApp
	ReqStore modules.RequestStore

	// The state of the FakeApp only consists of a counter of processed requests.
	RequestsProcessed uint64
}

// Apply implements Apply.
func (fa *FakeApp) Apply(batch *requestpb.Batch) error {
	for _, req := range batch.Requests {
		fa.ReqStore.RemoveRequest(req)
		fa.RequestsProcessed++
		fmt.Printf("Processed requests: %d\n", fa.RequestsProcessed)
	}
	return nil
}

func (fa *FakeApp) Snapshot() ([]byte, error) {
	return uint64ToBytes(fa.RequestsProcessed), nil
}

func (fa *FakeApp) RestoreState(snapshot []byte) error {
	return fmt.Errorf("we don't support state transfer in this test (yet)")
}

func uint64ToBytes(value uint64) []byte {
	byteValue := make([]byte, 8)
	binary.LittleEndian.PutUint64(byteValue, value)
	return byteValue
}
