package config

import (
	statepbtypes "github.com/filecoin-project/mir/pkg/pb/blockchainpb/statepb/types"
)

var InitialState = &statepbtypes.State{
	MessageHistory:     []string{},
	LastSentTimestamps: []*statepbtypes.LastSentTimestamp{},
}
