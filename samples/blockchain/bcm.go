// blockchain manager

package main

import (
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/blockchainpb"
	bcmpbdsl "github.com/filecoin-project/mir/pkg/pb/blockchainpb/bcmpb/dsl"
	tpmpbdsl "github.com/filecoin-project/mir/pkg/pb/blockchainpb/tpmpb/dsl"
	"github.com/filecoin-project/mir/samples/blockchain/ws/ws"
	"google.golang.org/protobuf/proto"
)

var (
	parentNotFound = errors.New("parent not found")
)

type bcmBlock struct {
	block     *blockchainpb.Block
	parent    *bcmBlock
	depth     uint64
	scanCount uint64
}

type bcmModule struct {
	m          *dsl.Module
	blocks     []bcmBlock
	leaves     map[uint64]*bcmBlock
	head       *bcmBlock
	blockCount uint64
	writeMutex *sync.Mutex // can be smarted about locking but just locking everything for now
	updateChan chan string
}

// TODO: might need to add some "cleanup" to handle very old leaves in case this grows too much

func (bcm *bcmModule) addBlock(block *blockchainpb.Block) error {
	println("Adding block...", block.BlockId)

	bcm.writeMutex.Lock()
	defer bcm.writeMutex.Unlock()

	parentId := block.PreviousBlockId
	// check leaves for parent with its id
	// return nil if it isn't there
	if parent, ok := bcm.leaves[parentId]; ok {
		println("Found parent in leaves")
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

	// NOTE: should probably create a structure where very old leaves are scanned later than the newer ones
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

		// check if curr is parent
		if curr.block.BlockId == parentId {
			println("Found parent in tree")
			blockNode := bcmBlock{block: block, parent: curr, depth: curr.depth + 1, scanCount: 0}
			// add new leave
			bcm.leaves[block.BlockId] = &blockNode
			bcm.blocks = append(bcm.blocks, blockNode)
			if blockNode.depth > bcm.head.depth { // update head only if new block is deeper, i.e. is the longest chain
				bcm.head = &blockNode
			}
			bcm.blockCount++
			return nil
		}

		// add curr's parents to queue if it hasn't been scanned yet
		if curr.parent != nil && curr.parent.scanCount <= currentScanCount {
			queue = append(queue, curr.parent)
		}
	}

	println("Couldn't find parent - invalid block. (TODO: add \"ask around for parent\")")

	return parentNotFound

}

func (bcm *bcmModule) getHead() *blockchainpb.Block {
	return bcm.head.block
}

func (bcm *bcmModule) handleNewBlock(block *blockchainpb.Block) {
	currentHead := bcm.getHead()
	// insert block
	if err := bcm.addBlock(block); err != nil {
		// silently ignore, just treat it as an invalid block
		// TODO: handle "rebuild" of chain if one missed a block or started later
		println(err)
	}

	bcm.updateChan <- strconv.FormatUint(block.BlockId, 10)

	// if head changed, trigger get tpm to prep new block
	if newHead := bcm.getHead(); newHead != currentHead {
		// new head
		println("New longest chain - head changed!")
		tpmpbdsl.NewHead(*bcm.m, "tpm", newHead.BlockId)
	}
}

func NewBCM(chainServerPort int) modules.PassiveModule {
	m := dsl.NewModule("bcm")
	var bcm bcmModule

	dsl.UponInit(m, func() error {

		// making up a genisis block
		genesis := &blockchainpb.Block{
			BlockId:         0,
			PreviousBlockId: 0,
			Payload:         &blockchainpb.Payload{Text: "genesis"},
			Timestamp:       0, // unix 0
		}

		hash := hashBlock(genesis)
		genesis.BlockId = hash

		// init bcm with genisis block
		bcm = bcmModule{m: &m, leaves: make(map[uint64]*bcmBlock), head: &bcmBlock{block: genesis, parent: nil, depth: 0}, blockCount: 1}
		bcm.leaves[hash] = bcm.head

		// init mutex
		bcm.writeMutex = &sync.Mutex{}

		// setup update channel
		bcm.updateChan = make(chan string)

		// start chain server (for visualization)
		go bcm.chainServer(chainServerPort)

		// kick off tpm with genisis block
		tpmpbdsl.NewHead(m, "tpm", hash)

		return nil
	})

	bcmpbdsl.UponNewBlock(m, func(block *blockchainpb.Block) error {
		// self-mined block - no need to verify
		println("~ Received block from miner")
		bcm.handleNewBlock(block)
		return nil
	})

	return m
}

func (bcm *bcmModule) chainServer(port int) error {
	fmt.Println("Starting chain server")

	websocket, send, _ := ws.NewWsServer(port)
	go websocket.StartServers()

	for {
		_ = <-bcm.updateChan

		bcm.writeMutex.Lock()
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

		bcm.writeMutex.Unlock()

		fmt.Printf("%d BLOCK COUNT @@@@@@@@\n", len(blocks))

		blockTreeMessage := blockchainpb.Blocktree{Blocks: blocks, Leaves: leaves}
		payload, err := proto.Marshal(&blockTreeMessage)
		if err != nil {
			panic(err)
		}

		send <- ws.WsMessage{MessageType: 2, Payload: payload}
	}

	// for {
	// 	_ = <-bcm.updateChan
	// 	bcm.writeMutex.Lock()
	// 	fmt.Println("new print")
	// 	queue := make([]bcmBlock, 0)
	// 	for _, v := range bcm.leaves {
	// 		queue = append(queue, v)
	// 	}
	// 	for len(queue) > 0 {
	// 		// pop from queue
	// 		curr := queue[0]
	// 		queue = queue[1:]

	// 		fmt.Println(curr.block.BlockId)

	// 		if curr.parent != nil {
	// 			queue = append(queue, *curr.parent)
	// 		}
	// 	}
	// 	bcm.writeMutex.Unlock()
	// }
	fmt.Println("Chain server stopped")

	return nil
}
