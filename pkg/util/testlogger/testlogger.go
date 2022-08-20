package testlogger

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/filecoin-project/mir/pkg/logging"
)

type testLogEntry struct {
	level logging.LogLevel
	text  string
	args  []interface{}
}

type TestLogger struct {
	entries    []*testLogEntry
	entryIndex map[string]int
}

func NewTestLogger() *TestLogger {
	return &TestLogger{
		entries:    make([]*testLogEntry, 0),
		entryIndex: make(map[string]int),
	}
}

func (tl *TestLogger) Log(level logging.LogLevel, text string, args ...interface{}) {
	newEntry := testLogEntry{
		level: level,
		text:  text,
		args:  args,
	}
	// TODO: With this implementation, duplicate entries are overwritten. Make duplicates possible.
	tl.entryIndex[fmt.Sprintf("%v", newEntry)] = len(tl.entries)
	tl.entries = append(tl.entries, &newEntry)
}

func (tl *TestLogger) IsConcurrent() bool {
	return false
}

func (tl *TestLogger) CheckEmpty(t *testing.T) {
	assert.Equal(t, 0, len(tl.entries), "log not empty, entries left: %d", len(tl.entries))
}

func (tl *TestLogger) CheckFirstEntry(t *testing.T, level logging.LogLevel, text string, args ...interface{}) {
	tl.checkAndRemoveEntry(t, &testLogEntry{
		level: level,
		text:  text,
		args:  args,
	}, 0)
}

func (tl *TestLogger) CheckAnyEntry(t *testing.T, level logging.LogLevel, text string, args ...interface{}) {
	entry := testLogEntry{
		level: level,
		text:  text,
		args:  args,
	}
	entryStr := fmt.Sprintf("%v", entry)
	idx, ok := tl.entryIndex[entryStr]
	if assert.True(t, ok, "entry not in log") {
		tl.checkAndRemoveEntry(t, &entry, idx)
		delete(tl.entryIndex, entryStr)
	}
}

func (tl *TestLogger) checkAndRemoveEntry(t *testing.T, refEntry *testLogEntry, idx int) {
	assert.Less(t, idx, len(tl.entries))
	entry := tl.entries[idx]

	assert.Equal(t, refEntry.level, entry.level)
	assert.Equal(t, refEntry.text, entry.text)
	assert.Equal(t, len(refEntry.args), len(entry.args))
	for i, arg := range refEntry.args {
		assert.Equal(t, arg, entry.args[i])
	}

	newEntries := tl.entries[0:idx]
	newEntries = append(newEntries, tl.entries[idx+1:]...)
	tl.entries = newEntries
}
