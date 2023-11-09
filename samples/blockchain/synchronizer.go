// blockchain manager

package main

import (
	"errors"
	"math/rand"

	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/blockchainpb"
	bcmpbdsl "github.com/filecoin-project/mir/pkg/pb/blockchainpb/bcmpb/dsl"
	synchronizerpbdsl "github.com/filecoin-project/mir/pkg/pb/blockchainpb/synchronizerpb/dsl"
	synchronizerpbmsgs "github.com/filecoin-project/mir/pkg/pb/blockchainpb/synchronizerpb/msgs"
	transportpbdsl "github.com/filecoin-project/mir/pkg/pb/transportpb/dsl"
	t "github.com/filecoin-project/mir/pkg/types"
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
	syncRequests map[uint64]*syncRequest
	getRequests  map[uint64]*getRequest
	otherNodes   []*t.NodeID // we remove nodes that we presume have crashed (see strikeList) // TODO: add a timeout to 're-add' them? exponential backoff?
	// stikeList    map[t.NodeID]int // a node gets a strike whenever it failed to handle a request, after STRIKE_THRESHOLD strikes it is removed from the otherNodes. Set to 0 if it handles a request successfully
}

/*************************
 * Outgoing sync requests
 *************************/

type syncRequest struct {
	requestID      uint64
	blockID        uint64
	leaveIDs       []uint64
	nodesContacted []int // index of node in otherNodes
	chain          []*blockchainpb.Block
}

func (sm *synchronizerModule) registerSyncRequest(blockID uint64) (uint64, error) {
	requestId := blockID

	// check if request already exists
	if _, ok := sm.syncRequests[requestId]; ok {
		return 0, ErrRequestAlreadyExists
	}

	sm.syncRequests[requestId] = &syncRequest{requestID: requestId, blockID: blockID, nodesContacted: []int{}}

	return requestId, nil
}

func (sm *synchronizerModule) contactNextNode(requestID uint64) error {
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
		nextNodeIndex = (request.nodesContacted[len(request.nodesContacted)] + 1) % len(sm.otherNodes)
	}

	request.nodesContacted = append(request.nodesContacted, nextNodeIndex)
	nextNode := *sm.otherNodes[nextNodeIndex]

	// send request
	transportpbdsl.SendMessage(*sm.m, "synchronizer", synchronizerpbmsgs.BlockRequest("synchronizer", requestID, request.blockID), []t.NodeID{nextNode})

	return nil
}

func (sm *synchronizerModule) handleSyncRequest(orphanBlockId uint64, leaveIds uint64) error {
	return nil
}

func (sm *synchronizerModule) handleBlockResponseReceived(from t.NodeID, requestID uint64, found bool, block *blockchainpb.Block) error {
	// check if request exists
	request, ok := sm.syncRequests[requestID]
	if !ok {
		// assume this is a delayed response and we already handled it
		return nil
	}

	if !found {
		// check whether the response came fromt he last node we contacted
		// NOTE: this is getting messy
		if *sm.otherNodes[request.nodesContacted[len(request.nodesContacted)-1]] != from {
			// there is still a request out in the ether - don't send another one
			return nil

		}
		// send request to the next node
		sm.contactNextNode(requestID)
		return nil
	}

	// we got a block
	request.chain = append(request.chain, block)
	// check if block.parentId is in leaveIDs
	for _, leaveID := range request.leaveIDs {
		if block.PreviousBlockId == leaveID {
			// we are done, send the chain to the blockchain manager, delete the request
			for i, j := 0, len(request.chain)-1; i < j; i, j = i+1, j-1 {
				request.chain[i], request.chain[j] = request.chain[j], request.chain[i]
			}
			bcmpbdsl.NewChain(*sm.m, "bcm", request.chain)
			delete(sm.syncRequests, requestID)
			return nil
		}
	}

	// update request
	request.blockID = block.PreviousBlockId
	request.nodesContacted = []int{}
	sm.contactNextNode(requestID)
	return nil
}

/*************************
 * Outgoing sync requests
 *************************/
type getRequest struct {
	requestID uint64
	from      t.NodeID
}

func (sm *synchronizerModule) handleBlockRequestReceived(from t.NodeID, requestID uint64, blockID uint64) error {
	// check if request is already being processed
	if _, ok := sm.getRequests[requestID]; ok {
		return ErrRequestAlreadyExists
	}

	// register request
	sm.getRequests[requestID] = &getRequest{requestID: requestID, from: from}

	// send get request to blockchain manager
	bcmpbdsl.GetBlockRequest(*sm.m, "bcm", requestID, "synchronizer", blockID)

	return nil

}

func (sm *synchronizerModule) handleGetBlockResponse(requestID uint64, found bool, block *blockchainpb.Block) error {
	request, ok := sm.getRequests[requestID]
	if !ok {
		return ErrUnkownGetBlockRequest
	}

	// respond to sync request
	transportpbdsl.SendMessage(*sm.m, "transport", synchronizerpbmsgs.BlockResponse("synchronizer", requestID, found, block), []t.NodeID{request.from})

	// delete request
	delete(sm.getRequests, requestID)

	return nil
}

func NewSynchronizer(otherNodes []*t.NodeID) modules.PassiveModule {
	m := dsl.NewModule("synchronizer")
	var sm synchronizerModule

	dsl.UponInit(m, func() error {
		return nil
	})

	// outgoing sync requests
	synchronizerpbdsl.UponSyncRequest(m, sm.handleSyncRequest)
	synchronizerpbdsl.UponBlockResponseReceived(m, sm.handleBlockResponseReceived)

	// for incoming sync requests
	synchronizerpbdsl.UponBlockRequestReceived(m, sm.handleBlockRequestReceived)
	bcmpbdsl.UponGetBlockResponse(m, sm.handleGetBlockResponse)

	return m
}
