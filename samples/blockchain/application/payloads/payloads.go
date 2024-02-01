package payloads

import (
	"slices"

	payloadpbtypes "github.com/filecoin-project/mir/pkg/pb/blockchainpb/payloadpb/types"
	"github.com/mitchellh/hashstructure"
)

/**
 * Payload manager
 * ===============
 *
 * The payload manager is part of the application and responsible for managing payloads.
 * It keeps track of all incoming payloads and provides the application with the oldest payload.
 * If no payloads are available, it returns nil.
 * The application can then use this payload to create a new block.
 * Payloads can be added and removed from the payload manager.
 * Equivalence of payloads is determined by their hash.
 */

type storedPayload struct {
	hash    uint64
	payload *payloadpbtypes.Payload
}

type PayloadManager struct {
	payloads []storedPayload // sorted by timestamp
}

func (tm *PayloadManager) PoolSize() int {
	return len(tm.payloads)
}

func New() *PayloadManager {
	return &PayloadManager{
		payloads: []storedPayload{},
	}
}

func (tm *PayloadManager) GetPayload() *payloadpbtypes.Payload {
	// return oldest payload, where timestamp is a field in the payload
	if len(tm.payloads) == 0 {
		return nil
	}
	// sort by timestamp
	// TODO: keep it sorted
	slices.SortFunc(tm.payloads, func(i, j storedPayload) int {
		return i.payload.Timestamp.AsTime().Compare(j.payload.Timestamp.AsTime())
	})
	return tm.payloads[0].payload
}

func (tm *PayloadManager) AddPayload(payload *payloadpbtypes.Payload) error {
	hash, err := hashstructure.Hash(payload, nil)
	if err != nil {
		return err
	}

	if slices.ContainsFunc(tm.payloads, func(t storedPayload) bool {
		return t.hash == hash
	}) {
		// already exists, ignore
		return nil
	}

	storedPaylod := storedPayload{
		hash:    hash,
		payload: payload,
	}

	tm.payloads = append(tm.payloads, storedPaylod)

	return nil
}

func (tm *PayloadManager) RemovePayload(payload *payloadpbtypes.Payload) error {
	hash, err := hashstructure.Hash(payload, nil)
	if err != nil {
		return err
	}

	// goes through all storedPayloads and removes the one with the matching hash
	// NOTE: consider adding (hash) index to improve performance - not important rn
	tm.payloads = slices.DeleteFunc(tm.payloads, func(t storedPayload) bool {
		return t.hash == hash
	})

	return nil
}
