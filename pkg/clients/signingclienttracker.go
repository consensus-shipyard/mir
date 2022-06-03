/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package clients

import (
	"fmt"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	"github.com/filecoin-project/mir/pkg/pb/statuspb"
	"github.com/filecoin-project/mir/pkg/serializing"
)

type SigningClientTracker struct {
	logger logging.Logger

	unverifiedRequests map[string]*requestpb.Request

	protocolModuleName string
}

func SigningTracker(protocolModuleName string, logger logging.Logger) *SigningClientTracker {

	// Ignore all logging if no logger was specified.
	if logger == nil {
		logger = logging.NilLogger
	}

	return &SigningClientTracker{
		logger:             logger,
		unverifiedRequests: make(map[string]*requestpb.Request),
		protocolModuleName: protocolModuleName,
	}
}

// ApplyEvent processes an event incoming to the SigningClientTracker module
// and produces a (potentially empty) list of new events to be processed by the node.
func (ct *SigningClientTracker) ApplyEvent(event *eventpb.Event) (*events.EventList, error) {
	switch e := event.Type.(type) {

	case *eventpb.Event_Request:
		// Request received from a client. Have the digest computed.

		req := e.Request
		return (&events.EventList{}).PushBack(events.HashRequest(
			"hasher",
			[][][]byte{serializing.RequestForHash(req)},
			&eventpb.HashOrigin{Module: "clientTracker", Type: &eventpb.HashOrigin_Request{Request: req}},
		)), nil

	case *eventpb.Event_HashResult:
		// Digest for a client request.
		// Verify request signature.
		// TODO: Implement request number watermarks.

		// Create a reference to the received request, including the computed hash.
		req := e.HashResult.Origin.Type.(*eventpb.HashOrigin_Request).Request
		digest := e.HashResult.Digests[0]
		reqRef := &requestpb.RequestRef{
			ClientId: req.ClientId,
			ReqNo:    req.ReqNo,
			Digest:   digest,
		}

		// Store a reference to the request until the signature verification result is available.
		// An alternative would be to store the payload in the request store directly
		// and delete it if authentication fails.
		// Another alternative would be to drag (a pointer to) the payload along in the VerifyRequestSig
		// and RequestSigVerified events. Both could be viable.
		ct.unverifiedRequests[reqStrKey(reqRef)] = req

		// Output a request authentication event.
		// This client tracker implementation assumes that client signatures are used for authenticating requests
		// and uses the VerifyRequestSig event (submitted to the Crypto module) to verify the signature.
		return (&events.EventList{}).PushBack(events.VerifyRequestSig("crypto", reqRef, req.Authenticator)), nil

	case *eventpb.Event_RequestSigVerified:

		// Convenience variables
		reqRef := e.RequestSigVerified.RequestRef
		req := ct.unverifiedRequests[reqStrKey(reqRef)]

		// Request is no longer unverified
		delete(ct.unverifiedRequests, reqStrKey(reqRef))

		if e.RequestSigVerified.Valid {
			// If signature is valid,
			// store the verified request in the request store and, submit a reference to it to the protocol.
			// It is important to first persist the request and only then submit it to the protocol,
			// in case the node crashes in between.
			storeEvent := events.StoreVerifiedRequest("crypto", reqRef, req.Data, req.Authenticator)
			storeEvent.Next = []*eventpb.Event{events.RequestReady(ct.protocolModuleName, reqRef)}
			return (&events.EventList{}).PushBack(storeEvent), nil
		}

		// If signature is not valid, ignore request
		ct.logger.Log(logging.LevelWarn, "Ignoring invalid request",
			"clID", reqRef.ClientId, "reqNo", reqRef.ReqNo, "err", e.RequestSigVerified.Error)
		return &events.EventList{}, nil
	default:
		return nil, fmt.Errorf("unknown signing client tracker event type: %T", event.Type)
	}
}

// TODO: Implement and document.
func (ct *SigningClientTracker) Status() (s *statuspb.ProtocolStatus, err error) {
	return nil, nil
}

// reqStrKey takes a request reference and transforms it to a string for using as a map key.
func reqStrKey(reqRef *requestpb.RequestRef) string {
	return fmt.Sprintf("%v-%d.%v", reqRef.ClientId, reqRef.ReqNo, reqRef.Digest)
}
