package main

import (
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
