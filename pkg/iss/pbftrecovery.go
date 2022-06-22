package iss

// TODO: This file is supposed to contain code for recovering the state of the PBFT protocol
//       from the WAL after a restart. Implement recovery here.

import (
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/isspbftpb"
)

// applyPbftPersistPreprepare processes a preprepare message loaded from the WAL.
func (pbft *pbftInstance) applyPbftPersistPreprepare(_ *isspbftpb.Preprepare) *events.EventList {

	// TODO: Implement this.
	pbft.logger.Log(logging.LevelDebug, "Loading WAL event: Preprepare (unimplemented)")
	return events.EmptyList()
}
