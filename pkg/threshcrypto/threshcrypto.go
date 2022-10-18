// Package threshcrypto provides an implementation of the MirModule module.
// It supports TBLS signatures.
package threshcrypto

import (
	"fmt"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/threshcryptopb"
	tcEvents "github.com/filecoin-project/mir/pkg/threshcrypto/events"
	t "github.com/filecoin-project/mir/pkg/types"
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

func (c *MirModule) ApplyEvent(event *eventpb.Event) (*events.EventList, error) {
	switch e := event.Type.(type) {
	case *eventpb.Event_Init:
		// no actions on init
		return events.EmptyList(), nil
	case *eventpb.Event_ThreshCrypto:
		return c.applyTCEvent(e.ThreshCrypto)
	default:
		// Complain about all other incoming event types.
		return nil, fmt.Errorf("unexpected type of event in threshcrypto MirModule: %T", event.Type)
	}
}

// apply a thresholdcryptopb.Event
func (c *MirModule) applyTCEvent(event *threshcryptopb.Event) (*events.EventList, error) {
	switch e := event.Type.(type) {
	case *threshcryptopb.Event_SignShare:
		// Compute signature share

		sigShare, err := c.threshCrypto.SignShare(e.SignShare.Data)
		if err != nil {
			return nil, err
		}

		return events.ListOf(
			tcEvents.SignShareResult(t.ModuleID(e.SignShare.Origin.Module), sigShare, e.SignShare.Origin),
		), nil

	case *threshcryptopb.Event_VerifyShare:
		// Verify signature share

		err := c.threshCrypto.VerifyShare(e.VerifyShare.Data, e.VerifyShare.SignatureShare, t.NodeID(e.VerifyShare.NodeId))
		ok := err == nil

		return events.ListOf(
			tcEvents.VerifyShareResult(t.ModuleID(e.VerifyShare.Origin.Module), ok, err.Error(), e.VerifyShare.Origin),
		), nil

	case *threshcryptopb.Event_VerifyFull:
		// Verify full signature

		err := c.threshCrypto.VerifyFull(e.VerifyFull.Data, e.VerifyFull.FullSignature)
		ok := err == nil

		return events.ListOf(
			tcEvents.VerifyFullResult(t.ModuleID(e.VerifyFull.Origin.Module), ok, err.Error(), e.VerifyFull.Origin),
		), nil

	case *threshcryptopb.Event_Recover:
		// Recover full signature from shares

		fullSig, err := c.threshCrypto.Recover(e.Recover.Data, e.Recover.SignatureShares)
		ok := err == nil

		return events.ListOf(
			tcEvents.RecoverResult(t.ModuleID(e.Recover.Origin.Module), fullSig, ok, err.Error(), e.Recover.Origin),
		), nil
	default:
		// Complain about all other incoming event types.
		return nil, fmt.Errorf("unexpected type of threshcrypto event in threshcrypto MirModule: %T", event.Type)
	}
}

// The ImplementsModule method only serves the purpose of indicating that this is a Module and must not be called.
func (c *MirModule) ImplementsModule() {}
