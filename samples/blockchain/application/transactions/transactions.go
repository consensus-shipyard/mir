package transactions

import (
	"cmp"
	"slices"

	payloadpbtypes "github.com/filecoin-project/mir/pkg/pb/blockchainpb/payloadpb/types"
	"github.com/mitchellh/hashstructure"
)

type transaction struct {
	hash    uint64
	payload *payloadpbtypes.Payload
}

type TransactionManager struct {
	transactions []transaction // sorted by timestamp
}

func (tm *TransactionManager) PoolSize() int {
	return len(tm.transactions)
}

func New() *TransactionManager {
	return &TransactionManager{
		transactions: []transaction{},
	}
}

func (tm *TransactionManager) GetPayload() *payloadpbtypes.Payload {
	// return oldest transaction, where timestamp is a field in the payload
	if len(tm.transactions) == 0 {
		return nil
	}
	// sort by timestamp
	// TODO: keep it sorted
	slices.SortFunc(tm.transactions, func(i, j transaction) int {
		return cmp.Compare[int64](i.payload.Timestamp, j.payload.Timestamp)
	})
	return tm.transactions[0].payload
}

func (tm *TransactionManager) AddPayload(payload *payloadpbtypes.Payload) error {
	hash, err := hashstructure.Hash(payload, nil)
	if err != nil {
		return err
	}

	if slices.ContainsFunc(tm.transactions, func(t transaction) bool {
		return t.hash == hash
	}) {
		// already exists, ignore
		return nil
	}

	transaction := transaction{
		hash:    hash,
		payload: payload,
	}

	tm.transactions = append(tm.transactions, transaction)

	return nil
}

func (tm *TransactionManager) RemovePayload(payload *payloadpbtypes.Payload) error {
	hash, err := hashstructure.Hash(payload, nil)
	if err != nil {
		return err
	}

	// goes through all transactions and removes the one with the matching hash
	// NOTE: consider adding (hash) index to improve performance - not important rn
	tm.transactions = slices.DeleteFunc(tm.transactions, func(t transaction) bool {
		return t.hash == hash
	})

	return nil
}
