// handles all commmunication between nodes

package broadcast

import (
	"github.com/filecoin-project/mir/pkg/blockchain/utils"
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	bcmpbdsl "github.com/filecoin-project/mir/pkg/pb/blockchainpb/bcmpb/dsl"
	broadcastpbdsl "github.com/filecoin-project/mir/pkg/pb/blockchainpb/broadcastpb/dsl"
	broadcastpbmsgs "github.com/filecoin-project/mir/pkg/pb/blockchainpb/broadcastpb/msgs"
	blockchainpbtypes "github.com/filecoin-project/mir/pkg/pb/blockchainpb/types"
	transportpbdsl "github.com/filecoin-project/mir/pkg/pb/transportpb/dsl"
	t "github.com/filecoin-project/mir/pkg/types"
)

/**
* Broadcast module
* =================
*
* The broadcast module is responsible for broadcasting new blocks to all other nodes.
* It either does this directly via the transport module or via the mangler (parameter mangle).
* If the mangler is used, messages might will be dropped and delayed.
* How many messages should be dropped can be configured by the parameter `dropRate`
* and the delay can be configured by the parameters `minDelay` and `maxDelay`.
 */

func NewBroadcast(otherNodes []t.NodeID, mangle bool, logger logging.Logger) modules.PassiveModule {
	m := dsl.NewModule("broadcast")

	dsl.UponInit(m, func() error {
		return nil
	})

	broadcastpbdsl.UponNewBlock(m, func(block *blockchainpbtypes.Block) error {
		// take the block and send it to all other nodes

		logger.Log(logging.LevelDebug, "broadcasting block", "blockId", utils.FormatBlockId(block.BlockId), "manlge", mangle)

		if mangle {
			// send via mangler
			for _, node := range otherNodes {
				transportpbdsl.SendMessage(m, "mangler", broadcastpbmsgs.NewBlockMessage("broadcast", block), []t.NodeID{node})
			}
		} else {
			transportpbdsl.SendMessage(m, "transport", broadcastpbmsgs.NewBlockMessage("broadcast", block), otherNodes)
		}

		return nil
	})

	broadcastpbdsl.UponNewBlockMessageReceived(m, func(from t.NodeID, block *blockchainpbtypes.Block) error {
		logger.Log(logging.LevelDebug, "new block received", "blockId", utils.FormatBlockId(block.BlockId))

		bcmpbdsl.NewBlock(m, "bcm", block)
		return nil
	})

	return m
}
