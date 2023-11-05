// miner

package main

import (
	"context"
	"math/rand"
	"time"

	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/blockchainpb"
	bcmmsgs "github.com/filecoin-project/mir/pkg/pb/blockchainpb/bcmpb/msgs"
	minerdsl "github.com/filecoin-project/mir/pkg/pb/blockchainpb/minerpb/dsl"
	transportpbdsl "github.com/filecoin-project/mir/pkg/pb/transportpb/dsl"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/mitchellh/hashstructure"
)

type minerModule struct {
	ctx    context.Context // to cancel ongoing mining when new block is requested
	cancel context.CancelFunc
}

type blockRequest struct {
	HeadId  uint64
	Payload *blockchainpb.Payload
}

func NewMiner(otherNodes []t.NodeID) modules.PassiveModule {

	m := dsl.NewModule("miner")
	var miner minerModule
	// blockRequests := make(chan blockRequest)

	dsl.UponInit(m, func() error {
		ctx, cancel := context.WithCancel(context.Background())
		miner = minerModule{ctx: ctx, cancel: cancel}

		println("Othr nodes:")
		for _, node := range otherNodes {
			println(node)
		}
		// go mineWorker(&m, blockRequests, otherNodes)

		return nil
	})

	minerdsl.UponBlockRequest(m, func(head_id uint64, payload *blockchainpb.Payload) error {

		// blockRequests <- blockRequest{head_id, payload}

		println("Mining block on top of", head_id)

		// cancel ongoing mining
		miner.cancel()
		// start new mining
		miner.ctx, miner.cancel = context.WithCancel(context.Background())
		ownCtx := &miner.ctx
		// random mining time to simulate PoW
		endtime := time.Now().Add(time.Duration(rand.ExpFloat64() * float64(2*time.Minute)))
		for {
			if time.Now().After(endtime) {
				break
			}

			select {
			case <-(*ownCtx).Done():
				println("Mining aborted")
				return nil
			default:
				// mine block
				time.Sleep(time.Second)
			}
		}
		println("Block mined!")
		block := &blockchainpb.Block{BlockId: 0, PreviousBlockId: head_id, Payload: payload, Timestamp: time.Now().Unix()}
		hash, err := hashstructure.Hash(block, nil)
		if err != nil {
			panic(err)
		}
		block.BlockId = hash
		// Block mined! Broadcast to blockchain and send event to bcm.
		// bcmdsl.NewBlock(m, "bcm", block)
		transportpbdsl.SendMessage(m, "transport", bcmmsgs.NewBlockMessage("bcm", block), otherNodes)
		println("Block broadcasted!")
		return nil

	})

	return m
}

// func mineWorker(m *dsl.Module, blockRequests chan blockRequest, otherNodes []t.NodeID) {
// 	ctx, cancel := context.WithCancel(context.Background())

// 	for {
// 		blockRequest := <-blockRequests
// 		cancel()                                               // abort ongoing mining
// 		ctx, cancel = context.WithCancel(context.Background()) // new context for new mining
// 		println("Mining block on top of", blockRequest.HeadId)
// 		go func() {
// 			endtime := time.Now().Add(time.Duration(rand.ExpFloat64() * float64(time.Minute)))
// 			for {
// 				if time.Now().After(endtime) {
// 					println("Block mined!")
// 					block := &blockchainpb.Block{BlockId: 0, PreviousBlockId: blockRequest.HeadId, Payload: blockRequest.Payload, Timestamp: time.Now().Unix()}
// 					hash, err := hashstructure.Hash(block, nil)
// 					if err != nil {
// 						panic(err)
// 					}
// 					block.BlockId = hash
// 					// Block mined! Broadcast to blockchain and send event to bcm.
// 					bcmdsl.NewBlock(*m, "bcm", block)
// 					transportpbdsl.SendMessage(*m, "transport", bcmmsgs.NewBlockMessage("bcm", block), otherNodes)
// 					println("Block broadcasted!")
// 					return
// 				}

// 				select {
// 				case <-ctx.Done():
// 					println("Mining aborted")
// 					return
// 				default:
// 					// mine block
// 					time.Sleep(time.Second)
// 				}
// 			}
// 		}()
// 		println("not blocking...")
// 	}
// }
