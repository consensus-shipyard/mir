package synchronizer

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/filecoin-project/mir/pkg/blockchain/utils"
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	bcmpbdsl "github.com/filecoin-project/mir/pkg/pb/blockchainpb/bcmpb/dsl"
	synchronizerpbdsl "github.com/filecoin-project/mir/pkg/pb/blockchainpb/synchronizerpb/dsl"
	synchronizerpbmsgs "github.com/filecoin-project/mir/pkg/pb/blockchainpb/synchronizerpb/msgs"
	blockchainpbtypes "github.com/filecoin-project/mir/pkg/pb/blockchainpb/types"
	transportpbdsl "github.com/filecoin-project/mir/pkg/pb/transportpb/dsl"
	t "github.com/filecoin-project/mir/pkg/types"
)

/**
 * Synchronizer module
 * ===================
 *
 * The synchronizer module assists the blockchain manager (BCM) in resolving cases whe BCM receives an orphan block.
 * An orphan block is a block that cannot be linked to the blockchain because the blockchain does not contain the block that the orphan block is linked to.
 * To do this, the synchronizer module communicates with other nodes to get the missing blocks.
 *
 * Terminology:
 * - internal sync request: a request to synchronize a chain segment that was initiated by this node
 * - external sync request: a request to synchronize a chain segment that was initiated by another node
 *
 * The synchronizer module performs the following tasks:
 * For internal sync requests:
 * 1. When it receives a SyncRequest event, it must register the request and send a ChainRequest message to another node.
 *    It asks one node after another.
 * 2. When it receives a successful ChainResponse message, it sends the blockchain manager (BCM) the chain fixing the missing bit with a NewChain event.
 *	  It then deletes the request.
 * 3. When it receives a unsuccessful ChainResponse message, it sends a ChainRequest message to the next node.
 *    If there are no more nodes to ask, it deletes the request.
 *
 * For external sync requests:
 * 1. When it receives a ChainRequest message, it must register the request and send a GetChainRequest event to the BCM.
 * 2. WHen the BCM responds with a GetChainResponse event, the synchronizer responds to the node that sent the ChainRequest message with a ChainResponse message.
 *
 * IMPORTANT: This module assumes that all other nodes resppond to requests. For this reason, the messages send by the synchronizer do not go through the mangler.
 */

var (
	ErrRequestAlreadyExists  = errors.New("request already exists")
	ErrNoMoreNodes           = errors.New("no more nodes to contact")
	ErrUnknownRequest        = errors.New("unknown request")
	ErrUnkownGetBlockRequest = errors.New("unknown get block request")
)

// NOTE: add 'crashed node detection' - remove nodes that don't respond to requests
// rn it would just get stuck on such a node...

type synchronizerModule struct {
	m              *dsl.Module
	nodeID         t.NodeID                // id of this node
	syncRequests   map[string]*syncRequest // map to keep track of sync requests - for internal sync requests
	getRequests    map[string]*getRequest  // map to keep track of get requests - for external sync request
	otherNodes     []t.NodeID              // ids of all other nodes
	internalLogger logging.Logger          // logger for internal sync requests
	externaLogger  logging.Logger          // logger for external sync requests
}

/*************************
 * Internal sync requests
 *************************/

// struct to keep track of sync requests
// it hold all information needed to send sync requests to other nodes and keeps track of which nodes have already been contacted
type syncRequest struct {
	requestID         string   // id of the request, used as key in syncRequests map
	blockID           uint64   // id of the blocks that cannot be linked to the blockchain
	connectionNodeIDs []uint64 // ids of the block
	nodesContacted    []int    // index of node in otherNodes
}

/**
 * Helper functions for internal requests
 */

// registers a sync request
func (sm *synchronizerModule) registerSyncRequest(block *blockchainpbtypes.Block, connectionNodeIDs []uint64) (string, error) {
	requestId := fmt.Sprint(block.BlockId) + "-" + string(sm.nodeID)
	// check if request already exists
	if _, ok := sm.syncRequests[requestId]; ok {
		return "", ErrRequestAlreadyExists
	}

	sm.syncRequests[requestId] = &syncRequest{requestID: requestId, blockID: block.BlockId, connectionNodeIDs: connectionNodeIDs, nodesContacted: []int{}}

	return requestId, nil
}

func (sm *synchronizerModule) contactNextNode(requestID string) error {
	request, ok := sm.syncRequests[requestID]
	if !ok {
		return ErrUnknownRequest
	}

	var nextNodeIndex int
	// check if we have contacted all nodes
	if len(request.nodesContacted) == len(sm.otherNodes) {
		return ErrNoMoreNodes
	} else if len(request.nodesContacted) == 0 {
		// first node to contact, pick randomly
		nextNodeIndex = rand.Intn(len(sm.otherNodes))

	} else {
		// get next node
		nextNodeIndex = (request.nodesContacted[len(request.nodesContacted)-1] + 1) % len(sm.otherNodes)
	}

	request.nodesContacted = append(request.nodesContacted, nextNodeIndex)
	nextNode := sm.otherNodes[nextNodeIndex]

	sm.internalLogger.Log(logging.LevelDebug, "asking node for block", "block", utils.FormatBlockId(request.blockID), "node", nextNode)

	// send request
	transportpbdsl.SendMessage(*sm.m, "transport", synchronizerpbmsgs.ChainRequest("synchronizer", requestID, request.blockID, request.connectionNodeIDs), []t.NodeID{nextNode})

	return nil
}

/**
 * Mir handlers for internal requests
 */

func (sm *synchronizerModule) handleSyncRequest(orphanBlock *blockchainpbtypes.Block, connectionNodeIDs []uint64) error {

	// register request
	requestId, err := sm.registerSyncRequest(orphanBlock, connectionNodeIDs)
	if err != nil {
		switch err {
		case ErrRequestAlreadyExists:
			// ignore request
			sm.internalLogger.Log(logging.LevelDebug, "sync request already exists", "orphanBlock", utils.FormatBlockId(orphanBlock.BlockId), "connectionNodeIDs", utils.FormatBlockIdSlice(connectionNodeIDs))
			return nil
		default:
			sm.internalLogger.Log(logging.LevelError, "sync registration failed", "error", err.Error(), "orphanBlock", utils.FormatBlockId(orphanBlock.BlockId), "connectionNodeIDs", utils.FormatBlockIdSlice(connectionNodeIDs))
			return err
		}
	}

	if err := sm.contactNextNode(requestId); err != nil {
		// could check for 'ErrNoMoreNodes' here but there should always be at least one node
		sm.internalLogger.Log(logging.LevelError, "sync registration failed", "error", err.Error(), "orphanBlock", utils.FormatBlockId(orphanBlock.BlockId), "connectionNodeIDs", utils.FormatBlockIdSlice(connectionNodeIDs))
		return err
	}

	return nil
}

func (sm *synchronizerModule) handleChainResponseReceived(from t.NodeID, requestID string, found bool, chain []*blockchainpbtypes.Block) error {
	// check if request exists
	request, ok := sm.syncRequests[requestID]
	if !ok {
		// assume this is a delayed response and we already handled it
		return nil
	}

	if !found {
		// check whether the response came from he last node we contacted
		// NOTE: this is getting messy
		if sm.otherNodes[request.nodesContacted[len(request.nodesContacted)-1]] != from {
			// there is still a request out in the ether - don't send another one
			return nil

		}
		// send request to the next node
		if err := sm.contactNextNode(requestID); err == ErrNoMoreNodes {
			sm.internalLogger.Log(logging.LevelError, "no more nodes to contact - forgetting request - shouldn't happen ", "error", err.Error(), "requestId", requestID)
			delete(sm.syncRequests, requestID)
		} else if err != nil {
			sm.internalLogger.Log(logging.LevelError, "Unexpected error contacting next node", "error", err.Error(), "requestId", requestID)
			// TODO: should we actually fail here?
			panic(err)
		}

		return nil
	}

	delete(sm.syncRequests, requestID)
	bcmpbdsl.NewChain(*sm.m, "bcm", chain)

	// we got a block
	return nil
}

/*************************
 * External sync requests
 *************************/
type getRequest struct {
	requestID string
	from      t.NodeID
}

/**
 * Mir handlers for external requests
 */

func (sm *synchronizerModule) handleChainRequestReceived(from t.NodeID, requestID string, blockID uint64, connectionNodeIDs []uint64) error {
	// check if request is already being processed
	if _, ok := sm.getRequests[requestID]; ok {
		sm.externaLogger.Log(logging.LevelError, "Get block request already exists", "from", from, "requestId", requestID)
		return nil
	}

	// register request
	sm.getRequests[requestID] = &getRequest{requestID: requestID, from: from}

	// send get request to blockchain manager
	bcmpbdsl.GetChainRequest(*sm.m, "bcm", requestID, "synchronizer", blockID, connectionNodeIDs)

	return nil

}

func (sm *synchronizerModule) handleGetChainResponse(requestID string, found bool, chain []*blockchainpbtypes.Block) error {
	request, ok := sm.getRequests[requestID]
	if !ok {
		sm.externaLogger.Log(logging.LevelError, "Unknown get block request", "requestId", requestID)
		return ErrUnkownGetBlockRequest
	}

	sm.externaLogger.Log(logging.LevelDebug, "Responsing to block request", "requestId", requestID, "found", found)

	// respond to sync request
	transportpbdsl.SendMessage(*sm.m, "transport", synchronizerpbmsgs.ChainResponse("synchronizer", requestID, found, chain), []t.NodeID{request.from})

	// delete request
	delete(sm.getRequests, requestID)

	return nil
}

func NewSynchronizer(nodeID t.NodeID, otherNodes []t.NodeID, logger logging.Logger) modules.PassiveModule {
	m := dsl.NewModule("synchronizer")
	var sm = synchronizerModule{
		m:              &m,
		nodeID:         nodeID,
		getRequests:    make(map[string]*getRequest),
		syncRequests:   make(map[string]*syncRequest),
		otherNodes:     otherNodes,
		internalLogger: logging.Decorate(logger, "Intern - "),
		externaLogger:  logging.Decorate(logger, "External - "),
	}

	dsl.UponInit(m, func() error {
		return nil
	})

	// internal sync requests
	synchronizerpbdsl.UponChainRequestReceived(m, sm.handleChainRequestReceived)
	bcmpbdsl.UponGetChainResponse(m, sm.handleGetChainResponse)

	// external sync requests
	synchronizerpbdsl.UponSyncRequest(m, sm.handleSyncRequest)
	synchronizerpbdsl.UponChainResponseReceived(m, sm.handleChainResponseReceived)

	return m
}
