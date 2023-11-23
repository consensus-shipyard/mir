package application

import "github.com/filecoin-project/mir/pkg/pb/blockchainpb"

type state struct {
	head        uint64
	state       map[uint64]int64
	headCounter int
}

func initState() *state {

	initialState := state{
		head:  0, // referenced by genesis block
		state: make(map[uint64]int64),
	}

	initialState.state[0] = 0

	return &initialState
}

func (s *state) applyBlock(block *blockchainpb.Block) {
	if _, ok := s.state[block.PreviousBlockId]; !ok {
		panic("previous block not found")
	}

	s.state[block.BlockId] = s.state[block.PreviousBlockId] + block.Payload.AddMinus

}

func (s *state) setHead(head uint64) {
	if _, ok := s.state[head]; !ok {
		panic("head not found")
	}
	s.head = head
}

func (s *state) getCurrentState() int64 {
	return s.state[s.head]
}
