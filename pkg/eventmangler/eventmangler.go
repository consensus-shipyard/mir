package eventmangler

import (
	"math/rand"
	"time"

	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/modules"
	eventpbdsl "github.com/filecoin-project/mir/pkg/pb/eventpb/dsl"
	eventpbtypes "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	"github.com/filecoin-project/mir/pkg/timer/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

type ModuleConfig struct {
	Self  t.ModuleID
	Dest  t.ModuleID
	Timer t.ModuleID
}

type ModuleParams struct {

	// All events that are not dropped are delayed at least by MinDelay.
	MinDelay time.Duration

	// All events that are not dropped are delayed by a time chosen uniformly at random between MinDelay and MaxDelay.
	MaxDelay time.Duration

	// Number between 0 and 1 indicating the probability with which an incoming event is dropped.
	DropRate float32

	// Source of randomness used for random decisions.
	// If set to nil, a new pseudo-random generator will be used.
	Rand *rand.Rand
}

func DefaultParams() *ModuleParams {
	return &ModuleParams{
		MinDelay: 0,
		MaxDelay: time.Second,
		DropRate: 0.05,
		Rand:     nil,
	}
}

func CheckParams(p *ModuleParams) error {
	if p.MinDelay < 0 {
		return es.Errorf("MinDelay must be non-negative, given: %v", p.MinDelay)
	}

	if p.MaxDelay < 0 {
		return es.Errorf("MaxDelay must be non-negative, given: %v", p.MaxDelay)
	}

	if p.MinDelay > p.MaxDelay {
		return es.Errorf("MinDelay (%v) must be smaller than MaxDelay (%v)", p.MinDelay, p.MaxDelay)
	}

	if p.DropRate < 0 {
		return es.Errorf("DropRate must be non-negative, given: %f", p.DropRate)
	}

	return nil
}

// NewModule returns a new instance of the event mangler module.
// The event mangler probabilistically drops or delays all incoming events.
// The drop rate and interval of possible delays is determined by the params argument.
func NewModule(mc ModuleConfig, params *ModuleParams) (modules.PassiveModule, error) {
	// Check whether parameters are valid.
	if err := CheckParams(params); err != nil {
		return nil, es.Errorf("invalid event mangler parameters: %w", err)
	}

	// Initialize randomness source.
	var r *rand.Rand
	if params.Rand != nil {
		// If a source is given in the configuration, use it.
		r = params.Rand
	} else {
		// If no randomness source is configured, create a new one using current time.
		r = rand.New(rand.NewSource(time.Now().UnixNano())) // nolint:gosec
	}

	// Create DSL module.
	m := dsl.NewModule(mc.Self)

	// Register only a single handler for all events, dropping and / or delaying them as configured.
	dsl.UponOtherEvent(m, func(ev *eventpbtypes.Event) error {

		// Drop event completely with probability params.DropRate
		if r.Float32() < params.DropRate {
			return nil
		}

		// Compute the delay to apply to this event.
		var delay time.Duration
		if params.MinDelay == params.MaxDelay {
			// If a fixed delay is specified, use it.
			delay = params.MinDelay
		} else {
			// Otherwise, generate a random delay within the specified interval.
			delay = params.MinDelay +
				time.Duration(r.Int63n((params.MaxDelay-params.MinDelay).Nanoseconds()))*time.Nanosecond
		}

		// Delay the event by the computed random time.
		if delay > 0 {
			eventpbdsl.TimerDelay(
				m,
				mc.Timer,
				[]*eventpbtypes.Event{events.Redirect(ev, mc.Dest)},
				types.Duration(delay),
			)
		} else {
			dsl.EmitEvent(m, events.Redirect(ev, mc.Dest))
		}

		return nil
	})

	return m, nil
}
