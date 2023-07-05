// Package threshcrypto provides an implementation of the MirModule module.
// It supports TBLS signatures.
package threshcrypto

import (
	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/modules"
	eventpbtypes "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	tcEvents "github.com/filecoin-project/mir/pkg/pb/threshcryptopb/events"
	threshcryptopbtypes "github.com/filecoin-project/mir/pkg/pb/threshcryptopb/types"
)

type MirModule struct {
	threshCrypto ThreshCrypto
}

func New(threshCrypto ThreshCrypto) *MirModule {
	return &MirModule{threshCrypto: threshCrypto}
}

func (c *MirModule) ApplyEvents(eventsIn *events.EventList) (*events.EventList, error) {
	return modules.ApplyEventsConcurrently(eventsIn, c.ApplyEvent)
}

func (c *MirModule) ApplyEvent(event *eventpbtypes.Event) (*events.EventList, error) {
	switch e := event.Type.(type) {
	case *eventpbtypes.Event_Init:
		// no actions on init
		return events.EmptyList(), nil
	case *eventpbtypes.Event_ThreshCrypto:
		return c.applyTCEvent(e.ThreshCrypto)
	default:
		// Complain about all other incoming event types.
		return nil, es.Errorf("unexpected type of event in threshcrypto MirModule: %T", event.Type)
	}
}

// apply a thresholdcryptopbtypes.Event
func (c *MirModule) applyTCEvent(event *threshcryptopbtypes.Event) (*events.EventList, error) {
	switch e := event.Type.(type) {
	case *threshcryptopbtypes.Event_SignShare:
		// Compute signature share

		sigShare, err := c.threshCrypto.SignShare(e.SignShare.Data)
		if err != nil {
			return nil, err
		}

		origin := e.SignShare.Origin
		return events.ListOf(
			tcEvents.SignShareResult(origin.Module, sigShare, origin),
		), nil

	case *threshcryptopbtypes.Event_VerifyShare:
		// Verify signature share

		err := c.threshCrypto.VerifyShare(e.VerifyShare.Data, e.VerifyShare.SignatureShare, e.VerifyShare.NodeId)

		ok := err == nil
		var errStr string
		if err != nil {
			errStr = err.Error()
		}

		origin := e.VerifyShare.Origin
		return events.ListOf(
			tcEvents.VerifyShareResult(origin.Module, ok, errStr, origin),
		), nil

	case *threshcryptopbtypes.Event_VerifyFull:
		// Verify full signature

		err := c.threshCrypto.VerifyFull(e.VerifyFull.Data, e.VerifyFull.FullSignature)

		ok := err == nil
		var errStr string
		if err != nil {
			errStr = err.Error()
		}

		origin := e.VerifyFull.Origin
		return events.ListOf(
			tcEvents.VerifyFullResult(origin.Module, ok, errStr, origin),
		), nil

	case *threshcryptopbtypes.Event_Recover:
		// Recover full signature from shares

		fullSig, err := c.threshCrypto.Recover(e.Recover.Data, e.Recover.SignatureShares)

		ok := err == nil
		var errStr string
		if err != nil {
			errStr = err.Error()
		}

		origin := e.Recover.Origin
		return events.ListOf(
			tcEvents.RecoverResult(origin.Module, fullSig, ok, errStr, origin),
		), nil
	default:
		// Complain about all other incoming event types.
		return nil, es.Errorf("unexpected type of threshcrypto event in threshcrypto MirModule: %T", event.Type)
	}
}

// The ImplementsModule method only serves the purpose of indicating that this is a Module and must not be called.
func (c *MirModule) ImplementsModule() {}
