package stdevents

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/filecoin-project/mir/stdtypes"
)

type fakeMessage struct {
	Content string
}

func (fm *fakeMessage) ToBytes() ([]byte, error) {
	return json.Marshal(fm)
}

func TestSerialization_Raw(t *testing.T) {
	e := NewRawWithSrc("srcModule", "destModule", []byte{0, 1, 2})

	data, err := e.ToBytes()
	assert.NoError(t, err)

	restoredEvent, err := Deserialize(data)
	assert.NoError(t, err)

	assert.Equal(t, e, restoredEvent)
}

func TestSerialization_Init(t *testing.T) {
	e := NewInitWithSrc("srcModule", "destModule")

	data, err := e.ToBytes()
	assert.NoError(t, err)

	restoredEvent, err := Deserialize(data)
	assert.NoError(t, err)

	assert.Equal(t, e, restoredEvent)
}

func TestSerialization_SendMessage(t *testing.T) {
	dummyMessage := fakeMessage{Content: "dummy message content"}
	e := NewSendMessageWithSrc(
		"srcModule",
		&dummyMessage,
		"localDestModule",
		"remoteDestModule",
		"some", "node", "IDs",
	)

	data, err := e.ToBytes()
	assert.NoError(t, err)

	restoredEvent, err := Deserialize(data)
	assert.NoError(t, err)

	dummyMsgBytes, err := dummyMessage.ToBytes()
	assert.NoError(t, err)

	expected := *e
	expected.Payload = stdtypes.RawMessage(dummyMsgBytes)
	assert.Equal(t, &expected, restoredEvent)
}

func TestSerialization_MessageReceived(t *testing.T) {
	dummyMessage := fakeMessage{Content: "dummy message content"}
	e := NewMessageReceivedWithSrc("srcModule", "destModule", "dummyNodeID", &dummyMessage)

	data, err := e.ToBytes()
	assert.NoError(t, err)

	restoredEvent, err := Deserialize(data)
	assert.NoError(t, err)

	dummyMsgBytes, err := dummyMessage.ToBytes()
	assert.NoError(t, err)

	expected := *e
	expected.Payload = stdtypes.RawMessage(dummyMsgBytes)
	assert.Equal(t, &expected, restoredEvent)
}

func TestSerialization_TimerDelay(t *testing.T) {
	dummyEvent1 := NewInitWithSrc("srcModule1", "destModule1")
	dummyEvent2 := NewRawWithSrc("srcModule2", "destModule2", []byte{0, 1, 2})

	e := NewTimerDelayWithSrc("srcModule", "destModule", time.Second, dummyEvent1, dummyEvent2)

	data, err := e.ToBytes()
	assert.NoError(t, err)

	restoredEvent, err := Deserialize(data)
	assert.NoError(t, err)

	dummyEvent1Raw, err := WrapInRaw(dummyEvent1)
	assert.NoError(t, err)
	dummyEvent2Raw, err := WrapInRaw(dummyEvent2)
	assert.NoError(t, err)

	expected := *e
	expected.Events[0] = dummyEvent1Raw
	expected.Events[1] = dummyEvent2Raw
	assert.Equal(t, &expected, restoredEvent)
}

func TestSerialization_TimerRepeat(t *testing.T) {
	dummyEvent1 := NewInitWithSrc("srcModule1", "destModule1")
	dummyEvent2 := NewRawWithSrc("srcModule2", "destModule2", []byte{0, 1, 2})

	e := NewTimerRepeatWithSrc("srcModule", "destModule", time.Second, 0, dummyEvent1, dummyEvent2)

	data, err := e.ToBytes()
	assert.NoError(t, err)

	restoredEvent, err := Deserialize(data)
	assert.NoError(t, err)

	dummyEvent1Raw, err := WrapInRaw(dummyEvent1)
	assert.NoError(t, err)
	dummyEvent2Raw, err := WrapInRaw(dummyEvent2)
	assert.NoError(t, err)

	expected := *e
	expected.Events[0] = dummyEvent1Raw
	expected.Events[1] = dummyEvent2Raw
	assert.Equal(t, &expected, restoredEvent)
}

func TestSerialization_TestString(t *testing.T) {
	e := NewTestStringWithSrc("srcModule", "destModule", "Hello!")

	data, err := e.ToBytes()
	assert.NoError(t, err)

	restoredEvent, err := Deserialize(data)
	assert.NoError(t, err)

	assert.Equal(t, e, restoredEvent)
}

func TestSerialization_TestUint64(t *testing.T) {
	e := NewTestUint64WithSrc("srcModule", "destModule", 123)

	data, err := e.ToBytes()
	assert.NoError(t, err)

	restoredEvent, err := Deserialize(data)
	assert.NoError(t, err)

	assert.Equal(t, e, restoredEvent)
}
