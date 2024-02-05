package bcm

import (
	"errors"
	"slices"
	"time"

	"github.com/filecoin-project/mir/pkg/blockchain/utils"
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	applicationpbdsl "github.com/filecoin-project/mir/pkg/pb/blockchainpb/applicationpb/dsl"
	bcmpbdsl "github.com/filecoin-project/mir/pkg/pb/blockchainpb/bcmpb/dsl"
	interceptorpbdsl "github.com/filecoin-project/mir/pkg/pb/blockchainpb/interceptorpb/dsl"
	minerpbdsl "github.com/filecoin-project/mir/pkg/pb/blockchainpb/minerpb/dsl"
	payloadpbtypes "github.com/filecoin-project/mir/pkg/pb/blockchainpb/payloadpb/types"
	statepbtypes "github.com/filecoin-project/mir/pkg/pb/blockchainpb/statepb/types"
	synchronizerpbdsl "github.com/filecoin-project/mir/pkg/pb/blockchainpb/synchronizerpb/dsl"
	blockchainpbtypes "github.com/filecoin-project/mir/pkg/pb/blockchainpb/types"
	t "github.com/filecoin-project/mir/pkg/types"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/types/known/timestamppb"
)

/**
 * Blockchain Manager Module (BCM)
 * ===============================
 *
 * The blockchain manager module is responsible for managing the blockchain.
 * It keeps track of all blocks and links them together to form a tree.
 * In particular, it keeps track of the head of the blockchain, all leaves and so-called checkpoints.
 * A checkpoint is a block stored by the BCM that has a state stored with it.
 * Technically, checkpoints are not necessary as the state can be computed from the blocks.
 * However, it is convenient not to have to recompute the state from the genesis block every time it is needed.
 *
 * The BCM must perform the following tasks:
 * 1. Initialize the blockchain by receiving an InitBlockchain event from the application module which contains the initial state that is associated with the genesis block.
 * 2. Add new blocks to the blockchain. If a block with a parent that that is not in the blockchain is added, the BCM requests the missing block from the synchronizer.
 *    Blocks that are missing their parent are called orphans.
 *    All blocks added to the blockchain are verified in two steps:
 *    - It has the application module verify that the payloads are valid given the chain that the block is part of.
 *    - The BCM must verify that the blocks link together correctly.
 *   Additionally, it sends a TreeUpdate event to the interceptor module. This is solely for debugging/visualization purposes and not necessary for the operation of the blockchain.
 *   Also, it instructs the miner to start mining on the new head (NewHead event).
 * 3. Register checkpoints when receiving a RegisterCheckpoint event from the application module.
 * 4. It must provide the synchronizer with chains when requested. This is to resolve orphan blocks in other nodes.
 * 5. When the head changes, it sends a HeadChange event to the application module. This event contains all information necessary for the application to compute the state at the new head
 *    as well as information about which payloads are now part of the canonical (i.e., longest) and which ones are no longer part of the canonical chain.
 * 6. Provide a chain of blocks from a checkpoint to the current head and the state associated with the checkpoint when receiving a GetChainToHeadRequest.
 *	  This is used by the application to query the current state.
 */

var (
	ErrParentNotFound    = errors.New("parent not found")
	ErrNodeAlreadyInTree = errors.New("node already in tree")
	ErrRollback          = errors.New("rollback - should never happen")
	ErrUninitialized     = errors.New("uninitialized - the application must initialize the blockchain first by sending it a initBlockchain message")
	ErrNoHit             = errors.New("no hit")
)

type bcmBlock struct {
	block     *blockchainpbtypes.Block
	state     *statepbtypes.State
	parent    *bcmBlock
	depth     uint64
	scanCount uint64
}

type bcmModule struct {
	m                *dsl.Module
	blocks           []bcmBlock // IMPORTANT: only to be used for visualization, not for actual block lookup
	leaves           map[uint64]*bcmBlock
	head             *bcmBlock
	genesis          *bcmBlock
	checkpoints      map[uint64]*bcmBlock
	currentScanCount uint64
	logger           logging.Logger
	initialized      bool
}

/**
 * Helper Functions
 */

// checkInitialization checks whether the blockchain has been initialized and returns an error if not
func (bcm *bcmModule) checkInitialization() error {
	if !bcm.initialized {
		bcm.logger.Log(logging.LevelError, ErrUninitialized.Error())
		return ErrUninitialized
	}

	return nil
}

// The blockTreeTraversal function traverses the block tree starting from the leaves and applying the traversalFunc to each block.
// It will stop when the traversalFunc returns true and return the current block for which the traversalFunc returned true.
// If the traversalFunc never returns true, the function will return nil and ErrNoHit.
// if the traversalFunc returns an error, the function will return nil and the error.
func (bcm *bcmModule) blockTreeTraversal(traversalFunc func(currBlock *bcmBlock) (bool, error)) (*bcmBlock, error) {
	bcm.currentScanCount = bcm.currentScanCount + 1
	queue := make([]*bcmBlock, 0)

	for _, v := range bcm.leaves {
		queue = append(queue, v)
	}

	for len(queue) > 0 {
		// pop from queue
		curr := queue[0]
		queue = queue[1:]

		// mark as scanned
		curr.scanCount = bcm.currentScanCount

		// check whether current block matches
		if match, err := traversalFunc(curr); err != nil {
			return nil, err
		} else if match {
			return curr, nil
		}

		// add curr's parents to queue if it exist hasn't been scanned yet
		// if curr.parent == nil, we are at the genesis block and can stop
		if curr.parent != nil && curr.parent.scanCount <= bcm.currentScanCount {
			queue = append(queue, curr.parent)
		}
	}

	return nil, ErrNoHit
}

func (bcm *bcmModule) handleNewHead(newHead, oldHead *bcmBlock) {
	bcm.logger.Log(logging.LevelInfo, "Head changed", "head", utils.FormatBlockId(newHead.block.BlockId))
	// TODO: short circuit if newHead is oldHead's parent

	// compute delta where removeChain is the chain leading to the old head and addChain is the chain leading to the new head
	// the first block in each chain is the common ancestor (fork root)
	removeChain, addChain, err := bcm.computeDelta(oldHead, newHead)
	if err != nil {
		// should never happen
		bcm.logger.Log(logging.LevelError, "Error computing delta", "error", err)
		panic(err)
	}

	// get fork state
	// TODO: fork might not happen right at checkpoint... find segment, add segment to fork update
	forkRoot, err := bcm.findBlock(addChain[0].BlockId)
	if err != nil {
		bcm.logger.Log(logging.LevelError, "Failed to get fork parent - this should not happen", "blockId", utils.FormatBlockId(addChain[0].BlockId))
		panic(err)
	}
	checkpointToForkRootChain, checkpointState := bcm.getChainFromCheckpointToBlock(forkRoot)

	applicationpbdsl.HeadChange(*bcm.m, "application",
		removeChain[1:],
		addChain[1:],
		checkpointToForkRootChain,
		checkpointState)

	minerpbdsl.NewHead(*bcm.m, "miner", newHead.block.BlockId)
}

func (bcm *bcmModule) computeDelta(removeHead *bcmBlock, addHead *bcmBlock) ([]*blockchainpbtypes.Block, []*blockchainpbtypes.Block, error) {
	bcm.logger.Log(logging.LevelDebug, "Computing Delta", "removeHead", removeHead.block.BlockId, "addHead", addHead.block.BlockId)

	// short circuit for most simple cases
	if removeHead.block.BlockId == addHead.block.BlockId {
		// no delta
		return []*blockchainpbtypes.Block{addHead.block}, []*blockchainpbtypes.Block{addHead.block}, nil
	} else if addHead.block.PreviousBlockId == removeHead.block.BlockId {
		// just appending
		return []*blockchainpbtypes.Block{removeHead.block}, []*blockchainpbtypes.Block{removeHead.block, addHead.block}, nil
	} else if removeHead.block.PreviousBlockId == addHead.block.BlockId {
		// rollbacks should never happen, this only checks for the most simple case
		return nil, nil, ErrRollback
	}

	// traverse backwards from both heads until we find a common ancestor
	// current ancecor is given when we find a block was already scanned by the other traversal
	// this is the case if
	initialScanCount := bcm.currentScanCount

	removeChain := make([]*blockchainpbtypes.Block, 0)
	addChain := make([]*blockchainpbtypes.Block, 0)

	currRemove := removeHead
	currAdd := addHead

	// TODO: consider weird cases like when they only have the genesis block in common
	// at least one hasn't reached the end yet
	for (currRemove != nil) || (currAdd != nil) {

		// handle remove step
		if currRemove != nil {

			removeChain = append(removeChain, currRemove.block)

			// check for intersection
			if currRemove.scanCount > initialScanCount {
				// remove chain intersects with add chain, remove chain is ok, add chain needs to be truncated
				// find currRemove in addChain
				currRemoveBlockId := currRemove.block.BlockId
				index := slices.IndexFunc(addChain, func(i *blockchainpbtypes.Block) bool { return i.BlockId == currRemoveBlockId })
				if index == -1 {
					// should never happen
					bcm.logger.Log(logging.LevelError, "Intersection trucation failed (remove intersection case) - this should never happen", "blockId", utils.FormatBlockId(currRemoveBlockId))
					panic("Intersection trucation failed (remove into add intersection case)")
				}
				return utils.Reverse(removeChain), utils.Reverse(addChain[:index+1]), nil
			}

			currRemove.scanCount = initialScanCount + 1
			currRemove = currRemove.parent
		}

		// handle add step
		if currAdd != nil {

			addChain = append(addChain, currAdd.block)

			// check for intersection
			if currAdd.scanCount > initialScanCount {
				// add chain intersects with remove chain, add chain is ok, remove chain needs to be truncated
				// find addRemove in removeChain
				currAddBlockId := currAdd.block.BlockId
				index := slices.IndexFunc(removeChain, func(i *blockchainpbtypes.Block) bool { return i.BlockId == currAddBlockId })
				if index == -1 {
					// should never happen
					bcm.logger.Log(logging.LevelError, "Intersection trucation failed (add intersection case) - this should never happen", "blockId", utils.FormatBlockId(currAddBlockId))
					panic("Intersection trucation failed (add into remove intersection case)")
				}
				return utils.Reverse(removeChain[:index+1]), utils.Reverse(addChain), nil
			}

			currAdd.scanCount = initialScanCount + 1
			currAdd = currAdd.parent
		}

	}

	// if we get here, we didn't find a common ancestor, this should never happen
	bcm.logger.Log(logging.LevelError, "Intersection not found - this should never happen")
	panic("intersection not found - this should never happen")
}

// TODO: might need to add some "cleanup" to handle very old leaves in case this grows too much
func (bcm *bcmModule) addBlock(block *blockchainpbtypes.Block) error {
	bcm.logger.Log(logging.LevelInfo, "Adding block...", "blockId", utils.FormatBlockId(block.BlockId), "parentId", utils.FormatBlockId(block.PreviousBlockId))

	// check if block is already in the leaves, reject if so
	if _, ok := bcm.leaves[block.BlockId]; ok {
		bcm.logger.Log(logging.LevelDebug, "Block already in leaves - ignore", "blockId", utils.FormatBlockId(block.BlockId))
		return nil
	}

	parentId := block.PreviousBlockId
	// check leaves for parent with its id
	// return nil if it isn't there
	if parent, ok := bcm.leaves[parentId]; ok {
		bcm.logger.Log(logging.LevelDebug, "Found parend in leaves", "blockId", utils.FormatBlockId(block.BlockId), "parentId", utils.FormatBlockId(parentId))
		blockNode := bcmBlock{
			block:  block,
			parent: parent,
			depth:  parent.depth + 1,
		}
		bcm.blocks = append(bcm.blocks, blockNode)
		// replace leave
		delete(bcm.leaves, parentId)
		bcm.leaves[block.BlockId] = &blockNode
		if blockNode.depth > bcm.head.depth { // update head only if new block is deeper, i.e. is the longest chain
			bcm.head = &blockNode
		}
		return nil
	}

	if _, err := bcm.blockTreeTraversal(func(currBlock *bcmBlock) (bool, error) {
		// check if curr matches the block to be added - if so, ignore
		if currBlock.block.BlockId == block.BlockId {
			bcm.logger.Log(logging.LevelDebug, "Block already in tree", "blockId", utils.FormatBlockId(block.BlockId))
			return true, ErrNodeAlreadyInTree
		}

		// check if curr is parent
		if currBlock.block.BlockId == parentId {
			bcm.logger.Log(logging.LevelDebug, "Found parend in tree", "blockId", utils.FormatBlockId(block.BlockId), "parentId", utils.FormatBlockId(parentId))
			blockNode := bcmBlock{
				block:     block,
				parent:    currBlock,
				depth:     currBlock.depth + 1,
				scanCount: 0,
			}
			// add new leave
			bcm.leaves[block.BlockId] = &blockNode
			bcm.blocks = append(bcm.blocks, blockNode)
			if blockNode.depth > bcm.head.depth { // update head only if new block is deeper, i.e. is the longest chain
				bcm.head = &blockNode
			}
			return true, nil
		}

		return false, nil
	}); err != nil {
		bcm.logger.Log(logging.LevelInfo, "Couldn't find parent", "blockId", utils.FormatBlockId(block.BlockId), "error", err)
		return err
	}

	return nil
}

func (bcm *bcmModule) findBlock(blockId uint64) (*bcmBlock, error) {
	return bcm.blockTreeTraversal(func(currBlock *bcmBlock) (bool, error) {
		if currBlock.block.BlockId == blockId {
			return true, nil
		}
		return false, nil
	})
}

// first send off to application to verify that payloads are valid, then check that everything links together
func (bcm *bcmModule) processNewChain(blocks []*blockchainpbtypes.Block) error {
	parent, _ := bcm.findBlock(blocks[0].PreviousBlockId)
	if parent == nil {
		bcm.logger.Log(logging.LevelError, "Parent not found. Could be invalid chain/block or we are missing a part. Sending sync req to sync end of chain.", "blockId", utils.FormatBlockId(blocks[0].BlockId))
		connectionNodeIDs := append(maps.Keys(bcm.leaves), bcm.genesis.block.BlockId)
		bcm.logger.Log(logging.LevelDebug, "Sending sync request", "blockId", utils.FormatBlockId(blocks[len(blocks)-1].BlockId), "leaves+genesis", utils.FormatBlockIdSlice(connectionNodeIDs))
		synchronizerpbdsl.SyncRequest(*bcm.m, "synchronizer", blocks[len(blocks)-1], connectionNodeIDs)
		return nil
	}

	checkpointToParent, checkpointState := bcm.getChainFromCheckpointToBlock(parent)

	applicationpbdsl.VerifyBlocksRequest(*bcm.m, "application", checkpointState, checkpointToParent, blocks)
	return nil
}

func (bcm *bcmModule) getChainFromCheckpointToBlock(block *bcmBlock) ([]*blockchainpbtypes.Block, *statepbtypes.State) {

	chain := make([]*blockchainpbtypes.Block, 0)

	currentBlock := block
	for currentBlock != nil {
		chain = append(chain, currentBlock.block)
		if bcmBlock, ok := bcm.checkpoints[currentBlock.block.BlockId]; ok {
			return utils.Reverse(chain), bcmBlock.state
		}
		currentBlock = currentBlock.parent
	}

	bcm.logger.Log(logging.LevelWarn, "Cannot link up with checkpoints", "source", utils.FormatBlockId(block.block.BlockId), "chain", chain)
	return nil, nil
}

/**
 * For visualization only, not relevant for the actual operation
 */

// Sends all blocks to the interceptor for visualization
func (bcm *bcmModule) sendTreeUpdate() {
	blocks := func() []*blockchainpbtypes.Block {
		blocks := make([]*blockchainpbtypes.Block, 0, len(bcm.blocks))
		for _, v := range bcm.blocks {
			blocks = append(blocks, v.block)
		}
		return blocks
	}()

	interceptorpbdsl.TreeUpdate(*bcm.m, "devnull", blocks, bcm.head.block.BlockId)
}

/**
 * Mir Event Handlers
 */

func (bcm *bcmModule) handleVerifyBlocksResponse(blocks []*blockchainpbtypes.Block) error {
	bcm.logger.Log(logging.LevelDebug, "Received verify blocks response", "first block", utils.FormatBlockId(blocks[0].BlockId), "last block", utils.FormatBlockId(blocks[len(blocks)-1].BlockId))

	currentHead := bcm.head
	// insert block
	for _, block := range blocks {
		if err := bcm.addBlock(block); err != nil {
			switch err {
			case ErrNodeAlreadyInTree:
				// ignore
				bcm.logger.Log(logging.LevelDebug, "Block already in tree - ignore", "blockId", utils.FormatBlockId(block.BlockId))
			case ErrParentNotFound:
				// this should not happen
				bcm.logger.Log(logging.LevelError, "Parent not found - this should not happen", "blockId", utils.FormatBlockId(block.BlockId))
				panic(err)
			default:
				// some other error
				bcm.logger.Log(logging.LevelError, "Unexpected error adding block", "blockId", utils.FormatBlockId(block.BlockId), "error", err)
				panic(err)
			}
		}
	}

	// if head changed, trigger get tpm to prep new block (only need to do this once)
	if newHead := bcm.head; newHead != currentHead {
		bcm.handleNewHead(newHead, currentHead)
	}

	// send to chainviewer
	bcm.sendTreeUpdate()

	return nil
}

func (bcm *bcmModule) handleNewBlock(block *blockchainpbtypes.Block) error {
	bcm.logger.Log(logging.LevelDebug, "Adding new block", "blockId", utils.FormatBlockId(block.BlockId))
	return bcm.processNewChain([]*blockchainpbtypes.Block{block})
}

func (bcm *bcmModule) handleNewChain(blocks []*blockchainpbtypes.Block) error {
	if len(blocks) == 0 {
		bcm.logger.Log(logging.LevelError, "Received empty chain")
		return nil
	}

	bcm.logger.Log(logging.LevelDebug, "Adding new chain", "first block", utils.FormatBlockId(blocks[0].BlockId), "last block", utils.FormatBlockId(blocks[len(blocks)-1].BlockId))

	return bcm.processNewChain(blocks)
}

func (bcm *bcmModule) handleRegisterCheckpoint(block_id uint64, state *statepbtypes.State) error {
	bcm.logger.Log(logging.LevelInfo, "Received register checkpoint", "blockId", utils.FormatBlockId(block_id))
	if err := bcm.checkInitialization(); err != nil {
		return err
	}
	if _, ok := bcm.checkpoints[block_id]; ok {

		bcm.logger.Log(logging.LevelDebug, "Checkpoint already registered - ignore", "blockId", utils.FormatBlockId(block_id))
		return nil
	}

	// find block
	// should be a leave or a block really close
	if hit, err := bcm.findBlock(block_id); err != nil {
		bcm.logger.Log(logging.LevelError, "Block not found - can't register checkpoint", "blockId", utils.FormatBlockId(block_id), "error", err)
		panic("block not found when registering checkpoint") // should never happen
	} else {
		bcm.logger.Log(logging.LevelDebug, "Found block - registering checkpoint", "blockId", utils.FormatBlockId(block_id))
		hit.state = state
		bcm.checkpoints[block_id] = hit
	}

	return nil
}

func (bcm *bcmModule) handleGetChainRequest(requestID string, sourceModule t.ModuleID, endBlockId uint64, sourceBlockIds []uint64) error {
	bcm.logger.Log(logging.LevelInfo, "Received get chain request", "requestId", requestID, "sourceModule", sourceModule, "endBlockId", utils.FormatBlockId(endBlockId), "sourceBlockIds", utils.FormatBlockIdSlice(sourceBlockIds))

	chain := make([]*blockchainpbtypes.Block, 0)
	// for easier lookup...
	sourceBlockIdsMap := make(map[uint64]uint64)
	for _, v := range sourceBlockIds {
		sourceBlockIdsMap[v] = v
	}
	currentBlock, err := bcm.findBlock(endBlockId)

	if err != nil {
		bcm.logger.Log(logging.LevelDebug, "Block not found - can't build chain", "requestId", requestID, "error", err)
		bcmpbdsl.GetChainResponse(*bcm.m, sourceModule, requestID, false, nil)
		return nil
	}

	bcm.logger.Log(logging.LevelDebug, "Found block in tree - building chain", "requestId", requestID)

	// initialize chain
	chain = append(chain, currentBlock.block)

	for currentBlock.parent != nil {
		// if parent is in sourceBlockIds, stop
		if _, ok := sourceBlockIdsMap[currentBlock.parent.block.BlockId]; ok {
			// chain build
			bcm.logger.Log(logging.LevelDebug, "Found link with sourceBlockIds - chain build", "requestId", requestID)
			// reversing chain to send it in the right order
			for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
				chain[i], chain[j] = chain[j], chain[i]
			}
			bcmpbdsl.GetChainResponse(*bcm.m, sourceModule, requestID, true, chain)
			return nil
		}

		chain = append(chain, currentBlock.parent.block)
		currentBlock = currentBlock.parent
	}

	// if we get here, we didn't find a link with sourceBlockIds
	bcm.logger.Log(logging.LevelDebug, "No link with sourceBlockIds - can't build chain", "requestId", requestID)
	bcmpbdsl.GetChainResponse(*bcm.m, sourceModule, requestID, false, nil)

	return nil

}

func (bcm *bcmModule) handleGetChainToHeadRequest(sourceModule t.ModuleID) error {
	bcm.logger.Log(logging.LevelInfo, "Received get chain to head request", "sourceModule", sourceModule)
	chain, checkpontState := bcm.getChainFromCheckpointToBlock(bcm.head)
	bcmpbdsl.GetChainToHeadResponse(*bcm.m, sourceModule, chain, checkpontState)
	return nil

}

func (bcm *bcmModule) handleInitBlockchain(initialState *statepbtypes.State) error {
	// initialize blockchain

	// making up a genisis block
	genesis := &blockchainpbtypes.Block{
		BlockId:         0,
		PreviousBlockId: 0,
		Payload:         &payloadpbtypes.Payload{},
		Timestamp:       timestamppb.New(time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)), // unix 0
		MinerId:         "",                                                        // no node mined this block, leaving node id empty
	}

	hash := utils.HashBlock(genesis) //uint64(0)
	genesis.BlockId = hash
	genesisBcm := bcmBlock{
		block:  genesis,
		state:  initialState,
		parent: nil,
		depth:  0,
	}

	bcm.head = &genesisBcm
	bcm.genesis = &genesisBcm

	bcm.blocks = []bcmBlock{genesisBcm}
	bcm.leaves[hash] = &genesisBcm
	bcm.checkpoints[hash] = &genesisBcm

	bcm.initialized = true

	// initialized, start operations

	bcm.logger.Log(logging.LevelInfo, "Initialized blockchain")
	minerpbdsl.NewHead(*bcm.m, "miner", hash)

	return nil
}

/**
 * Module Definition
 */

func NewBCM(logger logging.Logger) modules.PassiveModule {
	m := dsl.NewModule("bcm")

	// init bcm genisis block
	bcm := bcmModule{
		m:                &m,
		blocks:           nil, // will be initialized in handleInitBlockchain, triggered by application
		leaves:           make(map[uint64]*bcmBlock),
		head:             nil, // will be initialized in handleInitBlockchain, triggered by application
		genesis:          nil, // will be initialized in handleInitBlockchain, triggered by application
		checkpoints:      make(map[uint64]*bcmBlock),
		currentScanCount: 0,
		logger:           logger,
		initialized:      false,
	}

	dsl.UponInit(m, func() error {
		return nil
	})

	bcmpbdsl.UponNewBlock(m, func(block *blockchainpbtypes.Block) error {
		if err := bcm.checkInitialization(); err != nil {
			return err
		}
		return bcm.handleNewBlock(block)
	})

	bcmpbdsl.UponNewChain(m, func(blocks []*blockchainpbtypes.Block) error {
		bcm.logger.Log(logging.LevelInfo, "Received chain from synchronizer")
		if err := bcm.checkInitialization(); err != nil {
			return err
		}
		return bcm.handleNewChain(blocks)
	})

	bcmpbdsl.UponGetChainRequest(m, func(requestID string, sourceModule t.ModuleID, endBlockId uint64, sourceBlockIds []uint64) error {
		if err := bcm.checkInitialization(); err != nil {
			return err
		}
		return bcm.handleGetChainRequest(requestID, sourceModule, endBlockId, sourceBlockIds)
	})

	bcmpbdsl.UponGetChainToHeadRequest(m, func(sourceModule t.ModuleID) error {
		if err := bcm.checkInitialization(); err != nil {
			return err
		}
		return bcm.handleGetChainToHeadRequest(sourceModule)
	})

	bcmpbdsl.UponRegisterCheckpoint(m, bcm.handleRegisterCheckpoint)

	bcmpbdsl.UponInitBlockchain(m, bcm.handleInitBlockchain)

	applicationpbdsl.UponVerifyBlocksResponse(m, bcm.handleVerifyBlocksResponse)

	return m
}
