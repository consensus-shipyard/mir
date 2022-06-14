/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package reqstore

import (
	"fmt"
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/statuspb"
	"sync"

	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// VolatileRequestStore is an in-memory implementation of modules.RequestStore.
// All data is stored in RAM and the Sync() method does nothing.
// TODO: implement pruning of old data.
type VolatileRequestStore struct {
	modules.Module
	sync.RWMutex

	// Stores request entries, indexed by request reference.
	// Each entry holds all information (data, authentication, authenticator) about the referenced request.
	requests map[string]*requestInfo

	// For each request ID (client ID and request number pair), holds a set of request digests stored with the ID.
	// The set of request digests is itself represented as a string map,
	// where the key is the digest's string representation and the value is the digest as a byte slice.
	idIndex map[string]map[string][]byte
}

func (vrs *VolatileRequestStore) ApplyEvents(eventsIn *events.EventList) (*events.EventList, error) {

	eventsOut := &events.EventList{}

	iter := eventsIn.Iterator()
	for event := iter.Next(); event != nil; event = iter.Next() {
		evts, err := vrs.ApplyEvent(event)
		if err != nil {
			return nil, err
		}
		eventsOut.PushBackList(evts)
	}

	return eventsOut, nil
}

func (vrs *VolatileRequestStore) ApplyEvent(event *eventpb.Event) (*events.EventList, error) {
	// Process event based on its type.
	switch e := event.Type.(type) {
	case *eventpb.Event_StoreVerifiedRequest:
		storeEvent := e.StoreVerifiedRequest

		// Store request data.
		if err := vrs.PutRequest(storeEvent.RequestRef, storeEvent.Data); err != nil {
			return nil, fmt.Errorf("cannot store request (c%vr%d) data: %w",
				storeEvent.RequestRef.ClientId,
				storeEvent.RequestRef.ReqNo,
				err)
		}

		// Mark request as authenticated.
		if err := vrs.SetAuthenticated(storeEvent.RequestRef); err != nil {
			return nil, fmt.Errorf("cannot mark request (c%vr%d) as authenticated: %w",
				storeEvent.RequestRef.ClientId,
				storeEvent.RequestRef.ReqNo,
				err)
		}

		// Store request authenticator.
		if err := vrs.PutAuthenticator(storeEvent.RequestRef, storeEvent.Authenticator); err != nil {
			return nil, fmt.Errorf("cannot store authenticator (c%vr%d) of request: %w",
				storeEvent.RequestRef.ClientId,
				storeEvent.RequestRef.ReqNo,
				err)
		}

		return &events.EventList{}, nil
	default:
		return nil, fmt.Errorf("unknown request store event type: %T", event.Type)
	}

}

func (vrs *VolatileRequestStore) Status() (s *statuspb.ProtocolStatus, err error) {
	//TODO implement me
	panic("implement me")
}

// Holds the data stored by a single entry of the VolatileRequestStore.
type requestInfo struct {
	data          []byte
	authenticated bool
	authenticator []byte
}

// Returns the string representation of a request reference.
func requestKey(ref *requestpb.RequestRef) string {
	return fmt.Sprintf("r-%v.%d.%x", ref.ClientId, ref.ReqNo, ref.Digest)
}

// Returns the string representation of a request ID.
func idKey(clientID t.ClientID, reqNo t.ReqNo) string {
	return fmt.Sprintf("i-%v.%d", clientID, reqNo)
}

// Adds a digest to the request ID index.
func (vrs *VolatileRequestStore) updateIDIndex(reqRef *requestpb.RequestRef) {
	// Compute string key.
	key := idKey(t.ClientID(reqRef.ClientId), t.ReqNo(reqRef.ReqNo))

	// Look up index entry, creating a new one if none exists yet.
	entry, ok := vrs.idIndex[key]
	if !ok {
		entry = make(map[string][]byte)
		vrs.idIndex[key] = entry
	}

	// Add a copy of the digest to the index entry.
	d := make([]byte, len(reqRef.Digest))
	copy(d, reqRef.Digest)
	entry[fmt.Sprintf("%x", reqRef.Digest)] = d
}

// Looks up a stored entry and returns a pointer to it.
// Allocates a new one if none is present.
func (vrs *VolatileRequestStore) reqInfo(reqRef *requestpb.RequestRef) *requestInfo {

	// Look up the entry holding the information about this request
	key := requestKey(reqRef)
	reqInfo, ok := vrs.requests[key]
	if !ok {
		// If none is present, allocate a new one
		reqInfo = &requestInfo{
			data:          nil,
			authenticated: false,
			authenticator: nil,
		}
		vrs.requests[key] = reqInfo

		// Add the digest of the newly allocated entry
		vrs.updateIDIndex(reqRef)
	}

	return reqInfo
}

func NewVolatileRequestStore() *VolatileRequestStore {
	return &VolatileRequestStore{
		requests: make(map[string]*requestInfo),
		idIndex:  make(map[string]map[string][]byte),
	}
}

// PutRequest stores request the passed request data associated with the request reference.
func (vrs *VolatileRequestStore) PutRequest(reqRef *requestpb.RequestRef, data []byte) error {
	vrs.Lock()
	defer vrs.Unlock()

	// Look up entry for this request, creating a new one if necessary.
	reqInfo := vrs.reqInfo(reqRef)

	// Copy the request data to the entry (potentially discarding an old one).
	// Note that a full copy is made, so the stored data is not dependent on what happens with the original data.
	reqInfo.data = make([]byte, len(data))
	copy(reqInfo.data, data)

	return nil
}

// GetRequest returns the stored request data associated with the passed request reference.
// If no data is stored under the given reference, the returned error will be non-nil.
func (vrs *VolatileRequestStore) GetRequest(reqRef *requestpb.RequestRef) ([]byte, error) {
	vrs.RLock()
	defer vrs.RUnlock()

	reqInfo, ok := vrs.requests[requestKey(reqRef)]
	if !ok {
		// If the entry does not exist, return an error.
		return nil, fmt.Errorf(fmt.Sprintf("request (%v-%d.%x) not present",
			reqRef.ClientId, reqRef.ReqNo, reqRef.Digest))
	} else if reqInfo.data == nil {
		// If the entry exists, but contains no data, return an error.
		return nil, fmt.Errorf(fmt.Sprintf("request (%v-%d.%x) not present",
			reqRef.ClientId, reqRef.ReqNo, reqRef.Digest))
	}

	// Return a copy of the data (not a pointer to the data itself)
	data := make([]byte, len(reqInfo.data))
	copy(data, reqInfo.data)
	return data, nil
}

// RemoveRequest removes any request data associated with the passed request reference.
func (vrs *VolatileRequestStore) RemoveRequest(reqRef *requestpb.RequestRef) {
	vrs.Lock()
	defer vrs.Unlock()

	delete(vrs.requests, requestKey(reqRef))
}

// SetAuthenticated marks the referenced request as authenticated.
// A request being authenticated means that the local node believes that
// the request has indeed been sent by the client. This does not necessarily mean, however,
// that the local node can convince other nodes about the request's authenticity
// (e.g. if the local node received the request over an authenticated channel but the request is not signed).
func (vrs *VolatileRequestStore) SetAuthenticated(reqRef *requestpb.RequestRef) error {
	vrs.Lock()
	defer vrs.Unlock()

	// Look up entry for this request, creating a new one if necessary.
	reqInfo := vrs.reqInfo(reqRef)

	// Set the authenticated flag.
	reqInfo.authenticated = true

	return nil
}

// IsAuthenticated returns true if the request is authenticated, false otherwise.
func (vrs *VolatileRequestStore) IsAuthenticated(reqRef *requestpb.RequestRef) (bool, error) {
	vrs.RLock()
	defer vrs.RUnlock()

	reqInfo, ok := vrs.requests[requestKey(reqRef)]
	if !ok {
		// If the entry does not exist, return an error.
		return false, fmt.Errorf(fmt.Sprintf("request (%v.%d.%x) not present",
			reqRef.ClientId, reqRef.ReqNo, reqRef.Digest))
	}

	// If an entry for the referenced request is present, return the authenticated flag.
	return reqInfo.authenticated, nil
}

// PutAuthenticator stores an authenticator associated with the referenced request.
// If an authenticator is already stored under the same reference, it will be overwritten.
func (vrs *VolatileRequestStore) PutAuthenticator(reqRef *requestpb.RequestRef, auth []byte) error {
	vrs.Lock()
	defer vrs.Unlock()

	// Look up entry for this request, creating a new one if necessary.
	reqInfo := vrs.reqInfo(reqRef)

	// Copy the authenticator to the entry (potentially discarding an old one).
	// Note that a full copy is made, so the stored authenticator
	// is not dependent on what happens with the original authenticator passed to the function.
	reqInfo.authenticator = make([]byte, len(auth))
	copy(reqInfo.authenticator, auth)

	return nil
}

// GetAuthenticator returns the stored authenticator associated with the passed request reference.
// If no authenticator is stored under the given reference, the returned error will be non-nil.
func (vrs *VolatileRequestStore) GetAuthenticator(reqRef *requestpb.RequestRef) ([]byte, error) {
	vrs.RLock()
	defer vrs.RUnlock()

	reqInfo, ok := vrs.requests[requestKey(reqRef)]
	if !ok {
		// If the entry does not exist, return an error.
		return nil, fmt.Errorf(fmt.Sprintf("request (%v.%d.%x) not present",
			reqRef.ClientId, reqRef.ReqNo, reqRef.Digest))
	} else if reqInfo.authenticator == nil {
		// If the entry exists, but contains no authenticator, return an error.
		return nil, fmt.Errorf(fmt.Sprintf("request (%v.%d.%x) not present",
			reqRef.ClientId, reqRef.ReqNo, reqRef.Digest))
	}

	// Return a copy of the authenticator (not a pointer to the data itself)
	auth := make([]byte, len(reqInfo.authenticator))
	copy(auth, reqInfo.authenticator)
	return auth, nil
}

// GetDigestsByID returns a list of request digests for which any information
// (request data, authentication, or authenticator) is stored in the RequestStore.
func (vrs *VolatileRequestStore) GetDigestsByID(clientID t.ClientID, reqNo t.ReqNo) ([][]byte, error) {
	vrs.RLock()
	defer vrs.RUnlock()

	// Look up  index entry and allocate result structure.
	indexEntry, ok := vrs.idIndex[idKey(clientID, reqNo)]
	digests := make([][]byte, 0)

	if ok {
		// If an index entry is present, copy all the associated digests to a list and return it.
		// (If no entry is present, no digests will be added and an empty slice will be returned.)
		for _, digest := range indexEntry {
			// For each digest in the entry, append a copy of it to the list of digests that will be returned
			d := make([]byte, len(digest))
			copy(d, digest)
			digests = append(digests, d)
		}
	}

	return digests, nil
}

// Sync does nothing in this volatile (in-memory) RequestStore implementation.
func (vrs *VolatileRequestStore) Sync() error {
	return nil
}
