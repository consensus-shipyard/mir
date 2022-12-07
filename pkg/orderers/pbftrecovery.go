package orderers

// TODO: This file is supposed to contain code for recovering the state of the PBFT protocol
//       from the WAL after a restart. Implement recovery here.

import (
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/ordererspbftpb"
)

// applyPbftPersistPreprepare processes a preprepare message loaded from the WAL.
func (orderer *Orderer) applyPbftPersistPreprepare(_ *ordererspbftpb.Preprepare) *events.EventList {

	// TODO: Implement this.
	orderer.logger.Log(logging.LevelDebug, "Loading WAL event: Preprepare (unimplemented)")
	return events.EmptyList()
}
