// blockchain manager

package main

import (
	"errors"
	"strconv"

	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/blockchainpb"
	bcmpbdsl "github.com/filecoin-project/mir/pkg/pb/blockchainpb/bcmpb/dsl"
	minerpbdsl "github.com/filecoin-project/mir/pkg/pb/blockchainpb/minerpb/dsl"
	synchronizerpbdsl "github.com/filecoin-project/mir/pkg/pb/blockchainpb/synchronizerpb/dsl"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/samples/blockchain/ws/ws"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/proto"
)

var (
	ErrParentNotFound    = errors.New("parent not found")
	ErrNodeAlreadyInTree = errors.New("node already in tree")
)

type bcmBlock struct {
	block     *blockchainpb.Block
	parent    *bcmBlock
	depth     uint64
	scanCount uint64
}

type bcmModule struct {
	m          *dsl.Module
	blocks     []bcmBlock // IMPORTANT: only to be used for visualization, not for actual block lookup
	leaves     map[uint64]*bcmBlock
	head       *bcmBlock
	genesis    *bcmBlock
	blockCount uint64
	updateChan chan string
	logger     logging.Logger
}

// traversalFunc should return true if traversal should continue, false otherwise
// if it shouldn't continue, the function will return the current block
// if it should continue but there are no more blocks, the function will return nil
// if an error is thrown, the traversal is aborted and nil is returned
func (bcm *bcmModule) blockTreeTraversal(traversalFunc func(currBlock *bcmBlock) (bool, error)) *bcmBlock {
	currentScanCount := uint64(0) // to mark nodes that have been scanned in this round
	queue := make([]*bcmBlock, 0)
	for _, v := range bcm.leaves {
		queue = append(queue, v)
		if v.scanCount > currentScanCount {
			currentScanCount = v.scanCount
		}
	}
	for len(queue) > 0 {
		// pop from queue
		curr := queue[0]
		queue = queue[1:]

		// mark as scanned
		curr.scanCount = currentScanCount + 1

		if match, err := traversalFunc(curr); err != nil {
			return nil
		} else if match {
			return curr
		}

		// add curr's parents to queue if it hasn't been scanned yet
		if curr.parent != nil && curr.parent.scanCount <= currentScanCount {
			queue = append(queue, curr.parent)
		}
	}
	return nil
}

// TODO: might need to add some "cleanup" to handle very old leaves in case this grows too much
func (bcm *bcmModule) addBlock(block *blockchainpb.Block) error {
	bcm.logger.Log(logging.LevelInfo, "Adding block...", "blockId", formatBlockId(block.BlockId))

	// check if block is already in the leaves, reject if so
	if _, ok := bcm.leaves[block.BlockId]; ok {
		bcm.logger.Log(logging.LevelDebug, "Block already in leaves - ignore", "blockId", formatBlockId(block.BlockId))
		return nil
	}

	parentId := block.PreviousBlockId
	// check leaves for parent with its id
	// return nil if it isn't there
	if parent, ok := bcm.leaves[parentId]; ok {
		bcm.logger.Log(logging.LevelDebug, "Found parend in leaves", "blockId", formatBlockId(block.BlockId), "parentId", formatBlockId(parentId))
		blockNode := bcmBlock{block: block, parent: parent, depth: parent.depth + 1}
		bcm.blocks = append(bcm.blocks, blockNode)
		// replace leave
		delete(bcm.leaves, parentId)
		bcm.leaves[block.BlockId] = &blockNode
		if blockNode.depth > bcm.head.depth { // update head only if new block is deeper, i.e. is the longest chain
			bcm.head = &blockNode
		}
		bcm.blockCount++
		return nil
	}

	if hit := bcm.blockTreeTraversal(func(currBlock *bcmBlock) (bool, error) {
		// check if curr matches the block to be added - if so, ignore
		if currBlock.block.BlockId == block.BlockId {
			bcm.logger.Log(logging.LevelDebug, "Block already in tree", "blockId", formatBlockId(block.BlockId))
			return false, ErrNodeAlreadyInTree
		}

		// check if curr is parent
		if currBlock.block.BlockId == parentId {
			bcm.logger.Log(logging.LevelDebug, "Found parend in tree", "blockId", formatBlockId(block.BlockId), "parentId", formatBlockId(parentId))
			blockNode := bcmBlock{block: block, parent: currBlock, depth: currBlock.depth + 1, scanCount: 0}
			// add new leave
			bcm.leaves[block.BlockId] = &blockNode
			bcm.blocks = append(bcm.blocks, blockNode)
			if blockNode.depth > bcm.head.depth { // update head only if new block is deeper, i.e. is the longest chain
				bcm.head = &blockNode
			}
			bcm.blockCount++
			return true, nil
		}

		return false, nil
	}); hit == nil {
		bcm.logger.Log(logging.LevelInfo, "Couldn't find parent", "blockId", formatBlockId(block.BlockId))
		return ErrParentNotFound
	}

	return nil
}

func (bcm *bcmModule) getHead() *blockchainpb.Block {
	return bcm.head.block
}

func (bcm *bcmModule) handleNewBlock(block *blockchainpb.Block) {
	currentHead := bcm.getHead()
	// insert block
	if err := bcm.addBlock(block); err != nil {
		if err == ErrParentNotFound {
			// ask synchronizer for help
			leavesPlusGenesis := append(maps.Keys(bcm.leaves), bcm.genesis.block.BlockId)
			bcm.logger.Log(logging.LevelDebug, "Sending sync request", "blockId", formatBlockId(block.BlockId), "leaves+genesis", formatBlockIdSlice(leavesPlusGenesis))
			synchronizerpbdsl.SyncRequest(*bcm.m, "synchronizer", block, leavesPlusGenesis)
		} else {
			// some other error
			bcm.logger.Log(logging.LevelError, "Unexpected error adding block", "blockId", formatBlockId(block.BlockId), "error", err)
		}

	}

	bcm.updateChan <- strconv.FormatUint(block.BlockId, 10)

	// if head changed, trigger get tpm to prep new block
	if newHead := bcm.getHead(); newHead != currentHead {
		// new head
		bcm.logger.Log(logging.LevelInfo, "Head changed", "head", formatBlockId(newHead.BlockId))
		minerpbdsl.NewHead(*bcm.m, "miner", newHead.BlockId)
	}
}

func (bcm *bcmModule) handleNewChain(blocks []*blockchainpb.Block) {

	currentHead := bcm.getHead()
	// insert block
	for _, block := range blocks {
		if err := bcm.addBlock(block); err != nil {
			bcm.logger.Log(logging.LevelError, "Unexpected error adding block from chain", "blockId", formatBlockId(block.BlockId), "error", err)
		}
	}

	// get last block
	bcm.updateChan <- strconv.FormatUint(blocks[len(blocks)-1].BlockId, 10)

	// if head changed, trigger get tpm to prep new block
	if newHead := bcm.getHead(); newHead != currentHead {
		// new head

		bcm.logger.Log(logging.LevelInfo, "Head changed", "head", formatBlockId(newHead.BlockId))
		minerpbdsl.NewHead(*bcm.m, "miner", newHead.BlockId)
	}
}

func NewBCM(chainServerPort int, logger logging.Logger) modules.PassiveModule {
	m := dsl.NewModule("bcm")

	// making up a genisis block
	genesis := &blockchainpb.Block{
		BlockId:         0,
		PreviousBlockId: 0,
		Payload:         &blockchainpb.Payload{Text: "genesis"},
		Timestamp:       0, // unix 0
	}

	hash := hashBlock(genesis)
	genesis.BlockId = hash
	genesisBcm := bcmBlock{block: genesis, parent: nil, depth: 0}

	// init bcm genisis block
	bcm := bcmModule{
		m:          &m,
		blocks:     []bcmBlock{genesisBcm},
		leaves:     make(map[uint64]*bcmBlock),
		head:       &genesisBcm,
		genesis:    &genesisBcm,
		blockCount: 1,
		logger:     logger,
	}

	// add genesis to leaves
	bcm.leaves[hash] = bcm.head

	// setup update channel
	bcm.updateChan = make(chan string)

	// start chain server (for visualization)
	go bcm.chainServer(chainServerPort)

	dsl.UponInit(m, func() error {

		// kick off tpm with genisis block
		minerpbdsl.NewHead(m, "miner", hash)

		return nil
	})

	bcmpbdsl.UponNewBlock(m, func(block *blockchainpb.Block) error {
		// self-mined block - no need to verify
		bcm.handleNewBlock(block)
		return nil
	})

	bcmpbdsl.UponNewChain(m, func(blocks []*blockchainpb.Block) error {
		bcm.logger.Log(logging.LevelInfo, "Received chain from synchronizer")
		bcm.handleNewChain(blocks)
		return nil
	})

	bcmpbdsl.UponGetBlockRequest(m, func(requestID uint64, sourceModule t.ModuleID, blockID uint64) error {
		bcm.logger.Log(logging.LevelInfo, "Received get block request", "requestId", formatBlockId(requestID), "sourceModule", sourceModule)
		// check if block is in tree
		hit := bcm.blockTreeTraversal(func(currBlock *bcmBlock) (bool, error) {
			// check if curr matches the block to be added - if so, ignore
			if currBlock.block.BlockId == blockID {
				bcm.logger.Log(logging.LevelDebug, "Found block in tree", "requestId", formatBlockId(requestID))
				// respond to sync request
				return true, nil
			}

			return false, nil
		})

		if hit == nil {
			bcm.logger.Log(logging.LevelDebug, "Block not found", "requestId", formatBlockId(requestID))
			bcmpbdsl.GetBlockResponse(*bcm.m, sourceModule, requestID, false, nil)
			return nil
		}

		bcmpbdsl.GetBlockResponse(*bcm.m, sourceModule, requestID, true, hit.block)

		return nil
	})

	return m
}

func (bcm *bcmModule) chainServer(port int) error {
	bcm.logger.Log(logging.LevelInfo, "Starting chain server")

	websocket, send, _ := ws.NewWsServer(port)
	go websocket.StartServers()

	for {
		<-bcm.updateChan

		blocks := func() []*blockchainpb.Block {
			blocks := make([]*blockchainpb.Block, 0, len(bcm.blocks))
			for _, v := range bcm.blocks {
				blocks = append(blocks, v.block)
			}
			return blocks
		}()
		leaves := func() []uint64 {
			leaves := make([]uint64, 0, len(bcm.leaves))
			for _, v := range bcm.leaves {
				leaves = append(leaves, v.block.BlockId)
			}
			return leaves
		}()

		blockTreeMessage := blockchainpb.Blocktree{Blocks: blocks, Leaves: leaves}
		payload, err := proto.Marshal(&blockTreeMessage)
		if err != nil {
			panic(err)
		}

		send <- ws.WsMessage{MessageType: 2, Payload: payload}
	}

	return nil
}
