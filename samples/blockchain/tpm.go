// transaction pool manager

package main

import (
	"math/rand"
	"time"

	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/blockchainpb"
	minerdsl "github.com/filecoin-project/mir/pkg/pb/blockchainpb/minerpb/dsl"
	tpmpb "github.com/filecoin-project/mir/pkg/pb/blockchainpb/tpmpb/dsl"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func NewTPM(logger logging.Logger) modules.PassiveModule {

	m := dsl.NewModule("tpm")

	dsl.UponInit(m, func() error {
		return nil
	})

	tpmpb.UponNewBlockRequest(m, func(headId uint64) error {
		// just generating random payloads for now
		// generate random string
		logger.Log(logging.LevelInfo, "Processing block request", "headId", formatBlockId(headId))
		var seed *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
		length := seed.Intn(15)
		b := make([]byte, length)
		for i := range b {
			b[i] = charset[seed.Intn(len(charset))]
		}
		payload := &blockchainpb.Payload{Text: string(b)}

		minerdsl.BlockRequest(m, "miner", headId, payload)

		return nil
	})

	return m
}
