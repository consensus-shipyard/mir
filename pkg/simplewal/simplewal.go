/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// TODO: This is the original old code with very few modifications.
//       Go through all of it, comment what is to be kept and delete what is not needed.

// TODO: Decide whether to keep an explicit index or only use the Append function,
//       truncating using a more abstract concept (e.g. increasing non-unique numbers attached to entries).

// Package simplewal is a basic WAL implementation meant to be the first 'real' WAL
// option for mir. More sophisticated WALs with checksums, byte alignments, etc.
// may be produced in the future, but this is just a simple place to start.
package simplewal

import (
	"context"
	"fmt"
	"sync"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	t "github.com/filecoin-project/mir/pkg/types"

	"github.com/pkg/errors"
	"github.com/tidwall/wal"
	"google.golang.org/protobuf/proto"
)

type WAL struct {
	mutex sync.Mutex
	log   *wal.Log

	// Index of the next entry to append at the level of the underlying wal.
	idx uint64

	// Retention index as defined by the semantics of the WAL.
	// It needs to be persisted as well, because the WAL entries are not necessarily ordered by retention index.
	// It is required to skip outdated entries when loading.
	// Otherwise it could be completely ephemeral.
	// TODO: Implement persisting and loading the retentionIndex
	retentionIndex t.RetentionIndex
}

func (w *WAL) LoadAll(_ context.Context) (*events.EventList, error) {
	storedEvents := events.EmptyList()

	// Add all events from the WAL to the new EventList.
	if err := w.loadAll(func(retIdx t.RetentionIndex, event *eventpb.Event) {
		storedEvents.PushBack(event)
	}); err != nil {
		return nil, fmt.Errorf("could not load WAL events: %w", err)
	}

	return storedEvents, nil
}

func (w *WAL) ApplyEvents(eventsIn *events.EventList) (*events.EventList, error) {
	return modules.ApplyEventsSequentially(eventsIn, w.ApplyEvent)
}

func (w *WAL) ApplyEvent(event *eventpb.Event) (*events.EventList, error) {

	// Perform the necessary action based on event type.
	switch e := event.Type.(type) {
	case *eventpb.Event_Init:
		// no actions on init
	case *eventpb.Event_WalAppend:
		if err := w.Append(e.WalAppend.Event, t.RetentionIndex(e.WalAppend.RetentionIndex)); err != nil {
			return nil, fmt.Errorf("could not persist event (retention index %d) to WAL: %w",
				e.WalAppend.RetentionIndex, err)
		}
	case *eventpb.Event_WalTruncate:
		if err := w.Truncate(t.RetentionIndex(e.WalTruncate.RetentionIndex)); err != nil {
			return nil, fmt.Errorf("could not truncate WAL (retention index %d): %w",
				e.WalTruncate.RetentionIndex, err)
		}
	default:
		return nil, fmt.Errorf("unexpected type of WAL event: %T", event.Type)
	}

	// Then sync the WAL to disk.
	if err := w.Sync(); err != nil {
		return nil, fmt.Errorf("failed to sync WAL: %w", err)
	}

	return events.EmptyList(), nil
}

// The ImplementsModule method only serves the purpose of indicating that this is a Module and must not be called.
func (w *WAL) ImplementsModule() {}

func Open(path string) (*WAL, error) {

	// Create underlying log
	log, err := wal.Open(path, &wal.Options{
		NoSync: true,
		NoCopy: true,
	})
	if err != nil {
		return nil, errors.WithMessage(err, "could not open WAL")
	}

	// Initialize index.
	// The LastIndex obtained form the tidwall implementation happens to be the next index
	// (in terms of our WAL abstraction), as the underlying implementation starts counting at 1 and we start at 0.
	idx, err := log.LastIndex()
	if err != nil {
		return nil, fmt.Errorf("failed obtaining last WAL index: %w", err)
	}

	// TODO: Load retentionIndex from a (probably separate) file.

	// Return new object implementing the WAL abstraction.
	return &WAL{
		log: log,
		idx: idx,
	}, nil
}

func (w *WAL) IsEmpty() (bool, error) {
	firstIndex, err := w.log.FirstIndex()
	if err != nil {
		return false, errors.WithMessage(err, "could not read first index")
	}

	return firstIndex == 0, nil
}

func (w *WAL) loadAll(forEach func(index t.RetentionIndex, p *eventpb.Event)) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	firstIndex, err := w.log.FirstIndex()
	if err != nil {
		return errors.WithMessage(err, "could not read first index")
	}

	if firstIndex == 0 {
		// WAL is empty
		return nil
	}

	lastIndex, err := w.log.LastIndex()
	if err != nil {
		return errors.WithMessage(err, "could not read last index")
	}

	for i := firstIndex; i <= lastIndex; i++ {
		data, err := w.log.Read(i)
		if err != nil {
			return errors.WithMessagef(err, "could not read index %d", i)
		}

		result := &WALEntry{}
		err = proto.Unmarshal(data, result)
		if err != nil {
			return errors.WithMessage(err, "error decoding to proto, is the WAL corrupt?")
		}

		if t.RetentionIndex(result.RetentionIndex) >= w.retentionIndex {
			forEach(t.RetentionIndex(result.RetentionIndex), result.Event)
		}
	}

	return nil
}

func (w *WAL) write(index uint64, entry *WALEntry) error {

	// Check whether the index corresponds to the next index
	if w.idx != index {
		return fmt.Errorf("invalid wal index: expected %d, got %d", w.idx, index)
	}

	data, err := proto.Marshal(entry)
	if err != nil {
		return errors.WithMessage(err, "could not marshal")
	}

	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.idx = index + 1
	return w.log.Write(index+1, data) // The log implementation seems to be indexing starting with 1.
}

func (w *WAL) Append(event *eventpb.Event, retentionIndex t.RetentionIndex) error {
	return w.write(w.idx, &WALEntry{
		RetentionIndex: retentionIndex.Pb(),
		Event:          event,
	})
}

func (w *WAL) Truncate(retentionIndex t.RetentionIndex) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// TODO: Persist retention index first, probably in a separate file in the same directory.

	return w.log.TruncateFront(retentionIndex.Pb())
}

func (w *WAL) Sync() error {
	return w.log.Sync()
}

func (w *WAL) Close() error {
	return w.log.Close()
}
