// handles all commmunication between nodes

package main

import (
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/blockchainpb"
	bcmpbdsl "github.com/filecoin-project/mir/pkg/pb/blockchainpb/bcmpb/dsl"
	communicationpbdsl "github.com/filecoin-project/mir/pkg/pb/blockchainpb/communicationpb/dsl"
	communicationpbmsgs "github.com/filecoin-project/mir/pkg/pb/blockchainpb/communicationpb/msgs"
	transportpbdsl "github.com/filecoin-project/mir/pkg/pb/transportpb/dsl"
	t "github.com/filecoin-project/mir/pkg/types"
)

func NewCommunication(otherNodes []t.NodeID, mangle bool, logger logging.Logger) modules.PassiveModule {
	m := dsl.NewModule("communication")

	dsl.UponInit(m, func() error {
		return nil
	})

	communicationpbdsl.UponNewBlock(m, func(block *blockchainpb.Block) error {
		// take the block and send it to all other nodes

		logger.Log(logging.LevelDebug, "broadcasting block", "blockId", formatBlockId(block.BlockId), "manlge", mangle)

		if mangle {
			// send via mangles
			for _, node := range otherNodes {
				transportpbdsl.SendMessage(m, "mangler", communicationpbmsgs.NewBlockMessage("communication", block), []t.NodeID{node})
			}
		} else {
			transportpbdsl.SendMessage(m, "transport", communicationpbmsgs.NewBlockMessage("communication", block), otherNodes)
		}

		return nil
	})

	communicationpbdsl.UponNewBlockMessageReceived(m, func(from t.NodeID, block *blockchainpb.Block) error {
		logger.Log(logging.LevelDebug, "new block received", "blockId", formatBlockId(block.BlockId))

		bcmpbdsl.NewBlock(m, "bcm", block)
		return nil
	})

	return m
}
