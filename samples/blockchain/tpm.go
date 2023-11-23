// transaction pool manager

package main

import (
	"math/rand"
	"time"

	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	minerdsl "github.com/filecoin-project/mir/pkg/pb/blockchainpb/minerpb/dsl"
	"github.com/filecoin-project/mir/pkg/pb/blockchainpb/payloadpb"
	tpmpb "github.com/filecoin-project/mir/pkg/pb/blockchainpb/tpmpb/dsl"
	"github.com/filecoin-project/mir/samples/blockchain/utils"
)

func NewTPM(logger logging.Logger) modules.PassiveModule {

	m := dsl.NewModule("tpm")
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	dsl.UponInit(m, func() error {
		return nil
	})

	tpmpb.UponNewBlockRequest(m, func(headId uint64) error {
		// just generating random payloads for now
		// generate random string
		logger.Log(logging.LevelInfo, "Processing block request", "headId", utils.FormatBlockId(headId))
		value := int64(r.Intn(10000))
		payload := &payloadpb.Payload{AddMinus: value}

		minerdsl.BlockRequest(m, "miner", headId, payload)

		return nil
	})

	return m
}
