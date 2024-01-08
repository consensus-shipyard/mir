package application

import "github.com/filecoin-project/mir/pkg/pb/blockchainpb/statepb"

var InitialState = &statepb.State{
	MessageHistory: []string{},
}
