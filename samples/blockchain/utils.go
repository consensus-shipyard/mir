package main

import (
	"fmt"

	"github.com/filecoin-project/mir/pkg/pb/blockchainpb"
	"github.com/mitchellh/hashstructure"
)

func hashBlock(block *blockchainpb.Block) uint64 {
	hashBlock := &blockchainpb.Block{BlockId: 0, PreviousBlockId: block.PreviousBlockId, Payload: block.Payload, Timestamp: block.Timestamp}
	hash, err := hashstructure.Hash(hashBlock, nil)
	if err != nil {
		panic(err)
	}
	return hash
}

func formatBlockId(blockId uint64) string {
	strId := fmt.Sprint(blockId)
	// just in case we have a very small id for some reason
	if len(strId) > 6 {
		strId = strId[:6]
	}
	return strId
}

func formatBlockIdSlice(blockIds []uint64) string {
	str := ""
	for _, blockId := range blockIds {
		str += formatBlockId(blockId) + " "
	}
	return str
}
