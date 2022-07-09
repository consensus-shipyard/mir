package iss

import (
	"fmt"
	"github.com/filecoin-project/mir/pkg/pb/isspb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// epochInfo holds epoch-specific information that becomes irrelevant on advancing to the next epoch.
type epochInfo struct {

	// Epoch number.
	Nr t.EpochNr

	// IDs of nodes participating in the ordering protocol in this epoch.
	Membership []t.NodeID

	// Orderers associated with the epoch.
	Orderers []sbInstance

	// Checkpoint sub-protocol state.
	Checkpoint *checkpointTracker
}

// validateSBMessage checks whether an SBMessage is valid in this epoch.
// Returns nil if validation succeeds.
// If validation fails, returns the reason for which the message is considered invalid.
func (e *epochInfo) validateSBMessage(message *isspb.SBMessage, from t.NodeID) error {

	// Message must be destined for the current epoch.
	if t.EpochNr(message.Epoch) != e.Nr {
		return fmt.Errorf("invalid epoch: %v (expected %v)", message.Epoch, e.Nr)
	}

	// Message must refer to a valid SB instance.
	if int(message.Instance) > len(e.Orderers) {
		return fmt.Errorf("invalid SB instance number: %d", message.Instance)
	}

	// Message must be sent by a node in the current membership.
	// TODO: This lookup is extremely inefficient, computing the membership set on each message validation.
	//       Cache the output of membershipSet() throughout the epoch.
	if _, ok := membershipSet(e.Membership)[from]; !ok {
		return fmt.Errorf("sender of SB message not in the membership: %v", from)
	}

	return nil
}
