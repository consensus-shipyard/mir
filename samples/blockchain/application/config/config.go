package config

import (
	"github.com/filecoin-project/mir/pkg/pb/blockchainpb/payloadpb"
	"github.com/filecoin-project/mir/pkg/pb/blockchainpb/statepb"
)

var InitialState = &statepb.State{
	MessageHistory: []string{},
}

var EmptyPayload = &payloadpb.Payload{}
