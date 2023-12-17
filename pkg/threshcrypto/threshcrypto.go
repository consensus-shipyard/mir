// Package threshcrypto provides an implementation of the MirModule module.
// It supports TBLS signatures.
package threshcrypto

import (
	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/stdtypes"

	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/threshcryptopb"
	tcEvents "github.com/filecoin-project/mir/pkg/threshcrypto/events"
)

type MirModule struct {
	threshCrypto ThreshCrypto
}

func New(threshCrypto ThreshCrypto) *MirModule {
	return &MirModule{threshCrypto: threshCrypto}
}

func (c *MirModule) ApplyEvents(eventsIn *stdtypes.EventList) (*stdtypes.EventList, error) {
	return modules.ApplyEventsConcurrently(eventsIn, c.ApplyEvent)
}

func (c *MirModule) ApplyEvent(event stdtypes.Event) (*stdtypes.EventList, error) {

	// We only support proto events.
	pbevent, ok := event.(*eventpb.Event)
	if !ok {
		return nil, es.Errorf("The threshold crypto module only supports proto events, received %T", event)
	}

	switch e := pbevent.Type.(type) {
	case *eventpb.Event_Init:
		// no actions on init
		return stdtypes.EmptyList(), nil
	case *eventpb.Event_ThreshCrypto:
		return c.applyTCEvent(e.ThreshCrypto)
	default:
		// Complain about all other incoming event types.
		return nil, es.Errorf("unexpected type of event in threshcrypto MirModule: %T", pbevent.Type)
	}
}

// apply a thresholdcryptopb.Event
func (c *MirModule) applyTCEvent(event *threshcryptopb.Event) (*stdtypes.EventList, error) {
	switch e := event.Type.(type) {
	case *threshcryptopb.Event_SignShare:
		// Compute signature share

		sigShare, err := c.threshCrypto.SignShare(e.SignShare.Data)
		if err != nil {
			return nil, err
		}

		return stdtypes.ListOf(
			tcEvents.SignShareResult(stdtypes.ModuleID(e.SignShare.Origin.Module), sigShare, e.SignShare.Origin),
		), nil

	case *threshcryptopb.Event_VerifyShare:
		// Verify signature share

		err := c.threshCrypto.VerifyShare(e.VerifyShare.Data, e.VerifyShare.SignatureShare, stdtypes.NodeID(e.VerifyShare.NodeId))

		ok := err == nil
		var errStr string
		if err != nil {
			errStr = err.Error()
		}

		return stdtypes.ListOf(
			tcEvents.VerifyShareResult(stdtypes.ModuleID(e.VerifyShare.Origin.Module), ok, errStr, e.VerifyShare.Origin),
		), nil

	case *threshcryptopb.Event_VerifyFull:
		// Verify full signature

		err := c.threshCrypto.VerifyFull(e.VerifyFull.Data, e.VerifyFull.FullSignature)

		ok := err == nil
		var errStr string
		if err != nil {
			errStr = err.Error()
		}

		return stdtypes.ListOf(
			tcEvents.VerifyFullResult(stdtypes.ModuleID(e.VerifyFull.Origin.Module), ok, errStr, e.VerifyFull.Origin),
		), nil

	case *threshcryptopb.Event_Recover:
		// Recover full signature from shares

		fullSig, err := c.threshCrypto.Recover(e.Recover.Data, e.Recover.SignatureShares)

		ok := err == nil
		var errStr string
		if err != nil {
			errStr = err.Error()
		}

		return stdtypes.ListOf(
			tcEvents.RecoverResult(stdtypes.ModuleID(e.Recover.Origin.Module), fullSig, ok, errStr, e.Recover.Origin),
		), nil
	default:
		// Complain about all other incoming event types.
		return nil, es.Errorf("unexpected type of threshcrypto event in threshcrypto MirModule: %T", event.Type)
	}
}

// The ImplementsModule method only serves the purpose of indicating that this is a Module and must not be called.
func (c *MirModule) ImplementsModule() {}
