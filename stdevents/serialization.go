package stdevents

import (
	"encoding/json"
	"fmt"

	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/stdtypes"
)

// eventType represents the event types in serialized events.
// We use descriptive strings so that the serialized values are more human-readable.
// To save space (and bandwidth), change this to small integers (type byte).
type eventType string

const (
	RawEvent             = "Raw"
	InitEvent            = "Init"
	SendMessageEvent     = "SendMessage"
	MessageReceivedEvent = "MessageReceived"
	TimerDelayEvent      = "TimerDelay"
	TimerRepeatEvent     = "TimerRepeat"
	TestStringEvent      = "TestString"
	TestUint64Event      = "TestUint"
)

type serializedEvent struct {
	EvType eventType
	EvData json.RawMessage
}

func serialize(e any) ([]byte, error) {

	var evType eventType
	switch e.(type) {
	case *Raw:
		evType = RawEvent
	case *Init:
		evType = InitEvent
	case *serializableSendMessage:
		evType = SendMessageEvent
	case *serializableMessageReceived:
		evType = MessageReceivedEvent
	case *serializableTimerDelay:
		evType = TimerDelayEvent
	case *serializableTimerRepeat:
		evType = TimerRepeatEvent
	case *TestString:
		evType = TestStringEvent
	case *TestUint64:
		evType = TestUint64Event
	default:
		return nil, es.Errorf("unknown event type: %T", e)
	}

	evtData, err := json.Marshal(e)
	if err != nil {
		return nil, es.Errorf("could not marshal event: %w", err)
	}

	se := serializedEvent{
		EvType: evType,
		EvData: evtData,
	}

	data, err := json.Marshal(se)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func Deserialize(data []byte) (stdtypes.Event, error) {
	var se serializedEvent
	err := json.Unmarshal(data, &se)
	if err != nil {
		return nil, err
	}
	switch se.EvType {
	case RawEvent:
		var e Raw
		if err := json.Unmarshal(se.EvData, &e); err != nil {
			return nil, err
		}
		return &e, nil
	case InitEvent:
		var e Init
		if err := json.Unmarshal(se.EvData, &e); err != nil {
			return nil, err
		}
		return &e, nil
	case SendMessageEvent:
		var e serializableSendMessage
		if err := json.Unmarshal(se.EvData, &e); err != nil {
			return nil, err
		}
		return e.SendMessage(), nil
	case MessageReceivedEvent:
		var e serializableMessageReceived
		if err := json.Unmarshal(se.EvData, &e); err != nil {
			return nil, err
		}
		return e.MessageReceived(), nil
	case TimerDelayEvent:
		var e serializableTimerDelay
		if err := json.Unmarshal(se.EvData, &e); err != nil {
			return nil, err
		}
		return e.TimerDelay(), nil
	case TimerRepeatEvent:
		var e serializableTimerRepeat
		if err := json.Unmarshal(se.EvData, &e); err != nil {
			return nil, err
		}
		return e.TimerRepeat(), nil
	case TestStringEvent:
		var e TestString
		if err := json.Unmarshal(se.EvData, &e); err != nil {
			return nil, err
		}
		return &e, nil
	case TestUint64Event:
		var e TestUint64
		if err := json.Unmarshal(se.EvData, &e); err != nil {
			return nil, err
		}
		return &e, nil
	default:
		return nil, fmt.Errorf("unknown event type: %s", se.EvType)
	}
}
