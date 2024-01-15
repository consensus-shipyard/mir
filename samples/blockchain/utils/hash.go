package utils

import (
	blockchainpbtypes "github.com/filecoin-project/mir/pkg/pb/blockchainpb/types"
	"github.com/mitchellh/hashstructure"
)

func HashBlock(block *blockchainpbtypes.Block) uint64 {
	hashBlock := &blockchainpbtypes.Block{BlockId: 0, PreviousBlockId: block.PreviousBlockId, Payload: block.Payload, Timestamp: block.Timestamp}
	hash, err := hashstructure.Hash(hashBlock, nil)
	if err != nil {
		panic(err)
	}
	return hash
}
