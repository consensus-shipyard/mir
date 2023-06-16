package testlogger

import (
	"fmt"
	"sync"
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
	entries        map[int]*testLogEntry
	entryIndex     map[string]int
	internalMutex  sync.Mutex
	entriesCounter int
}

func New() *TestLogger {
	return &TestLogger{
		entries:       make(map[int]*testLogEntry, 0),
		entryIndex:    make(map[string]int),
		internalMutex: sync.Mutex{},
	}
}

func (tl *TestLogger) Log(level logging.LogLevel, text string, args ...interface{}) {
	newEntry := testLogEntry{
		level: level,
		text:  text,
		args:  args,
	}
	tl.internalMutex.Lock()
	defer tl.internalMutex.Unlock()
	// TODO: With this implementation, duplicate entries are overwritten. Make duplicates possible.
	tl.entryIndex[fmt.Sprintf("%v", newEntry)] = tl.entriesCounter
	tl.entries[tl.entriesCounter] = &newEntry
	tl.entriesCounter++
}

func (tl *TestLogger) MinLevel() logging.LogLevel {
	return logging.LevelTrace
}

func (tl *TestLogger) IsConcurrent() bool {
	return false
}

func (tl *TestLogger) CheckEmpty(t *testing.T) {
	assert.Equal(t, 0, len(tl.entries), "log not empty, entries left: %d", len(tl.entries))
}

func (tl *TestLogger) CheckFirstEntry(t *testing.T, level logging.LogLevel, text string, args ...interface{}) {
	entry := testLogEntry{
		level: level,
		text:  text,
		args:  args,
	}
	tl.checkAndRemoveEntry(t, &entry, 0)
}

func (tl *TestLogger) CheckAnyEntry(t *testing.T, level logging.LogLevel, text string, args ...interface{}) {
	entry := testLogEntry{
		level: level,
		text:  text,
		args:  args,
	}
	idx, ok := tl.entryIndex[fmt.Sprintf("%v", entry)]
	if assert.True(t, ok, "entry not in log") {
		tl.checkAndRemoveEntry(t, &entry, idx)
	}
}

func (tl *TestLogger) checkAndRemoveEntry(t *testing.T, refEntry *testLogEntry, idx int) {

	// Mark function as test helper
	t.Helper()

	// Check that the log contains enough entries.
	entry, ok := tl.entries[idx]

	assert.True(t, ok, "log does not contain entry at index %d", idx)

	// Check the content of the log entry.
	assert.Equal(t, refEntry.level, entry.level, "unexpected log level")
	assert.Equal(t, refEntry.text, entry.text, "unexpected log message")
	assert.Equal(t, len(refEntry.args), len(entry.args), "unexpected number of log message parameters")
	for i, arg := range refEntry.args {
		assert.Equal(t, arg, entry.args[i], "unexpected log message parameter at offset %d", i)
	}

	// Remove entry from the log.
	delete(tl.entries, idx)
	delete(tl.entryIndex, fmt.Sprintf("%v", *entry))

}
