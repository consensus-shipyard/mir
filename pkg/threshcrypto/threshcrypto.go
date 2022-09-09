// Package crypto provides an implementation of the MirModule module.
// It supports RSA and ECDSA signatures.
package threshcrypto

import (
	"fmt"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
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

	case *eventpb.Event_ThreshSign:
		// Compute signature share

		if sigShare, err := c.threshCrypto.SignShare(e.ThreshSign.Data); err != nil {
			return nil, err
		} else {
			return events.ListOf(
				events.ThreshSignResult(t.ModuleID(e.ThreshSign.Origin.Module), sigShare, e.ThreshSign.Origin),
			), nil
		}

	case *eventpb.Event_ThreshVerShare:
		// Verify signature share

		err := c.threshCrypto.VerifyShare(e.ThreshVerShare.Data, e.ThreshVerShare.SignatureShare)
		ok := err == nil

		return events.ListOf(
			events.ThreshVerShareResult(t.ModuleID(e.ThreshVerShare.Origin.Module), ok, err.Error(), e.ThreshVerShare.Origin),
		), nil

	case *eventpb.Event_ThreshVerFull:
		// Verify full signature

		err := c.threshCrypto.VerifyFull(e.ThreshVerFull.Data, e.ThreshVerFull.FullSignature)
		ok := err == nil

		return events.ListOf(
			events.ThreshVerFullResult(t.ModuleID(e.ThreshVerFull.Origin.Module), ok, err.Error(), e.ThreshVerFull.Origin),
		), nil

	case *eventpb.Event_ThreshRecover:
		// Recover full signature from shares

		fullSig, err := c.threshCrypto.Recover(e.ThreshRecover.Data, e.ThreshRecover.SignatureShares)
		ok := err == nil

		return events.ListOf(
			events.ThreshRecoverResult(t.ModuleID(e.ThreshRecover.Origin.Module), fullSig, ok, err.Error(), e.ThreshRecover.Origin),
		), nil

	default:
		// Complain about all other incoming event types.
		return nil, fmt.Errorf("unexpected type of MirModule event: %T", event.Type)
	}
}

// The ImplementsModule method only serves the purpose of indicating that this is a Module and must not be called.
func (c *MirModule) ImplementsModule() {}
