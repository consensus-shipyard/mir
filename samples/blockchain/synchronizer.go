// blockchain manager

package main

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/blockchainpb"
	bcmpbdsl "github.com/filecoin-project/mir/pkg/pb/blockchainpb/bcmpb/dsl"
	synchronizerpbdsl "github.com/filecoin-project/mir/pkg/pb/blockchainpb/synchronizerpb/dsl"
	synchronizerpbmsgs "github.com/filecoin-project/mir/pkg/pb/blockchainpb/synchronizerpb/msgs"
	transportpbdsl "github.com/filecoin-project/mir/pkg/pb/transportpb/dsl"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/samples/blockchain/utils"
)

const (
	STRIKE_THRESHOLD = 2
)

var (
	ErrRequestAlreadyExists  = errors.New("request already exists")
	ErrNoMoreNodes           = errors.New("no more nodes to contact")
	ErrUnknownRequest        = errors.New("unknown request")
	ErrUnkownGetBlockRequest = errors.New("unknown get block request")
)

// NOTE: add 'crashed node detection' - remove nodes that don't respond to requests
// rn it would just get stuck on such a node...

type synchronizerModule struct {
	m            *dsl.Module
	nodeID       t.NodeID // id of this node
	syncRequests map[string]*syncRequest
	getRequests  map[string]*getRequest
	otherNodes   []t.NodeID // we remove nodes that we presume have crashed (see strikeList) // TODO: add a timeout to 're-add' them? exponential backoff?
	// stikeList    map[t.NodeID]int // a node gets a strike whenever it failed to handle a request, after STRIKE_THRESHOLD strikes it is removed from the otherNodes. Set to 0 if it handles a request successfully
	mangle         bool
	internalLogger logging.Logger
	externaLogger  logging.Logger
}

/*************************
 * Outgoing sync requests
 *************************/

type syncRequest struct {
	requestID      string
	blockID        uint64
	leaveIDs       []uint64
	nodesContacted []int // index of node in otherNodes
}

func (sm *synchronizerModule) registerSyncRequest(block *blockchainpb.Block, leaves []uint64) (string, error) {
	requestId := fmt.Sprint(block.BlockId) + "-" + string(sm.nodeID)
	// check if request already exists
	if _, ok := sm.syncRequests[requestId]; ok {
		return "", ErrRequestAlreadyExists
	}

	sm.syncRequests[requestId] = &syncRequest{requestID: requestId, blockID: block.BlockId, leaveIDs: leaves, nodesContacted: []int{}}

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
		// NOTE: stupid to keep an array of indices but it's more flexible in case I want to change the way I pick the next node
		nextNodeIndex = (request.nodesContacted[len(request.nodesContacted)-1] + 1) % len(sm.otherNodes)
	}

	request.nodesContacted = append(request.nodesContacted, nextNodeIndex)
	nextNode := sm.otherNodes[nextNodeIndex]

	sm.internalLogger.Log(logging.LevelDebug, "asking node for block", "block", utils.FormatBlockId(request.blockID), "node", nextNode, "mangle", sm.mangle)
	// send request
	if sm.mangle {
		transportpbdsl.SendMessage(*sm.m, "mangler", synchronizerpbmsgs.ChainRequest("synchronizer", requestID, request.blockID, request.leaveIDs), []t.NodeID{nextNode})
	} else {
		transportpbdsl.SendMessage(*sm.m, "transport", synchronizerpbmsgs.ChainRequest("synchronizer", requestID, request.blockID, request.leaveIDs), []t.NodeID{nextNode})
	}

	return nil
}

func (sm *synchronizerModule) handleSyncRequest(orphanBlock *blockchainpb.Block, leaveIds []uint64) error {

	// register request
	requestId, err := sm.registerSyncRequest(orphanBlock, leaveIds)
	if err != nil {
		switch err {
		case ErrRequestAlreadyExists:
			// ignore request
			sm.internalLogger.Log(logging.LevelDebug, "sync request already exists", "orphanBlock", utils.FormatBlockId(orphanBlock.BlockId), "leaveIds", utils.FormatBlockIdSlice(leaveIds))
			return nil
		default:
			sm.internalLogger.Log(logging.LevelError, "sync registration failed", "error", err.Error(), "orphanBlock", utils.FormatBlockId(orphanBlock.BlockId), "leaveIds", utils.FormatBlockIdSlice(leaveIds))
			return err
		}
	}

	if err := sm.contactNextNode(requestId); err != nil {
		// could check for 'ErrNoMoreNodes' here but there should always be at least one node
		sm.internalLogger.Log(logging.LevelError, "sync registration failed", "error", err.Error(), "orphanBlock", utils.FormatBlockId(orphanBlock.BlockId), "leaveIds", utils.FormatBlockIdSlice(leaveIds))
		return err
	}

	return nil
}

func (sm *synchronizerModule) handleChainResponseReceived(from t.NodeID, requestID string, found bool, chain []*blockchainpb.Block) error {
	// check if request exists
	request, ok := sm.syncRequests[requestID]
	if !ok {
		// assume this is a delayed response and we already handled it
		return nil
	}

	if !found {
		// check whether the response came fromt he last node we contacted
		// NOTE: this is getting messy
		if sm.otherNodes[request.nodesContacted[len(request.nodesContacted)-1]] != from {
			// there is still a request out in the ether - don't send another one
			return nil

		}
		// send request to the next node
		if err := sm.contactNextNode(requestID); err == ErrNoMoreNodes {
			sm.internalLogger.Log(logging.LevelError, "no more nodes to contact - forgetting request - shouldn't happen if not mangling", "error", err.Error(), "requestId", requestID, "mangle", sm.mangle)
			delete(sm.syncRequests, requestID)
			if !sm.mangle {
				panic(err) // NOTE: this should never happen as long as we don't start mangling
			}
		} else if err != nil {
			sm.internalLogger.Log(logging.LevelError, "Unexpected error contacting next node", "error", err.Error(), "requestId", requestID, "mangle", sm.mangle)
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
 * Outgoing sync requests
 *************************/
type getRequest struct {
	requestID string
	from      t.NodeID
}

func (sm *synchronizerModule) handleChainRequestReceived(from t.NodeID, requestID string, blockID uint64, leaveIds []uint64) error {
	// check if request is already being processed
	if _, ok := sm.getRequests[requestID]; ok {
		sm.externaLogger.Log(logging.LevelError, "Get block request already exists", "from", from, "requestId", requestID, "mangle", sm.mangle)
		return nil
	}

	// register request
	sm.getRequests[requestID] = &getRequest{requestID: requestID, from: from}

	// send get request to blockchain manager
	bcmpbdsl.GetChainRequest(*sm.m, "bcm", requestID, "synchronizer", blockID, leaveIds)

	return nil

}

func (sm *synchronizerModule) handleGetChainResponse(requestID string, found bool, chain []*blockchainpb.Block) error {
	request, ok := sm.getRequests[requestID]
	if !ok {
		sm.externaLogger.Log(logging.LevelError, "Unknown get block request", "requestId", requestID, "mangle", sm.mangle)
		return ErrUnkownGetBlockRequest
	}

	sm.externaLogger.Log(logging.LevelInfo, "Responsing to block request", "requestId", requestID, "found", found, "mangle", sm.mangle)

	// respond to sync request
	if sm.mangle {
		transportpbdsl.SendMessage(*sm.m, "transport", synchronizerpbmsgs.ChainResponse("synchronizer", requestID, found, chain), []t.NodeID{request.from})
	} else {
		transportpbdsl.SendMessage(*sm.m, "mangler", synchronizerpbmsgs.ChainResponse("synchronizer", requestID, found, chain), []t.NodeID{request.from})
	}

	// delete request
	delete(sm.getRequests, requestID)

	return nil
}

func NewSynchronizer(nodeID t.NodeID, otherNodes []t.NodeID, mangle bool, logger logging.Logger) modules.PassiveModule {
	m := dsl.NewModule("synchronizer")
	var sm = synchronizerModule{
		m:              &m,
		nodeID:         nodeID,
		getRequests:    make(map[string]*getRequest),
		syncRequests:   make(map[string]*syncRequest),
		otherNodes:     otherNodes,
		mangle:         mangle,
		internalLogger: logging.Decorate(logger, "Intern - "),
		externaLogger:  logging.Decorate(logger, "External - "),
	}

	dsl.UponInit(m, func() error {
		return nil
	})

	// outgoing sync requests
	synchronizerpbdsl.UponSyncRequest(m, sm.handleSyncRequest)
	synchronizerpbdsl.UponChainResponseReceived(m, sm.handleChainResponseReceived)

	// for incoming sync requests
	synchronizerpbdsl.UponChainRequestReceived(m, sm.handleChainRequestReceived)
	bcmpbdsl.UponGetChainResponse(m, sm.handleGetChainResponse)

	return m
}
