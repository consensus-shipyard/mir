// handles all commmunication between nodes

package main

import (
	"math/rand"

	"log"

	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/blockchainpb"
	bcmpbdsl "github.com/filecoin-project/mir/pkg/pb/blockchainpb/bcmpb/dsl"
	communicationpbdsl "github.com/filecoin-project/mir/pkg/pb/blockchainpb/communicationpb/dsl"
	communicationpbmsgs "github.com/filecoin-project/mir/pkg/pb/blockchainpb/communicationpb/msgs"
	transportpbdsl "github.com/filecoin-project/mir/pkg/pb/transportpb/dsl"
	t "github.com/filecoin-project/mir/pkg/types"
)

/* Note on simulated message loss and similar:
 * Although package loss is very easy to simulate here, it might be more appropriate to use the event mangler.
 * Not sure about this yet but something to keep in mind later
 */

func NewCommunication(otherNodes []t.NodeID, probabilityMessageLost float32) modules.PassiveModule {
	m := dsl.NewModule("communication")

	dsl.UponInit(m, func() error {
		return nil
	})

	communicationpbdsl.UponNewBlock(m, func(block *blockchainpb.Block) error {
		// take the block and send it to all other nodes
		// could add so randomization here - only send to a subset of nodes

		// simulate message loss
		receivers := make([]t.NodeID, 0, len(otherNodes))
		if probabilityMessageLost == 0 {
			receivers = otherNodes
		} else {
			for _, node := range otherNodes {
				if probabilityMessageLost > 1 && rand.Float32() < probabilityMessageLost {
					receivers = append(receivers, node)
				}
			}
		}

		if len(receivers) == 0 {
			log.Printf("Block %d send to NO nodes", block.BlockId)
			return nil
		} else if len(receivers) == len(otherNodes) {
			log.Printf("Block %d send to all nodes", block.BlockId)
		} else {
			log.Printf("Block %d send to %d/%d nodes", block.BlockId, len(receivers), len(otherNodes))
		}

		transportpbdsl.SendMessage(m, "transport", communicationpbmsgs.NewBlockMessage("communication", block), receivers)

		return nil
	})

	communicationpbdsl.UponNewBlockMessageReceived(m, func(from t.NodeID, block *blockchainpb.Block) error {
		// take the block and add it to the blockchain
		// could add some randomization here - delay/drop (drop implemented in send)

		bcmpbdsl.NewBlock(m, "bcm", block)
		return nil
	})

	return m
}
