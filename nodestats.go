package mir

import (
	"fmt"
	"sync"
	"time"

	"github.com/filecoin-project/mir/pkg/logging"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
)

// ==============================================================================================================
// Event dispatching statistics
// ==============================================================================================================

// StatsConfig configures the generation of event processing statistics.
type StatsConfig struct {

	// Logger to send event processing statistics to.
	Logger logging.Logger

	// Level with which to log event processing statistics.
	LogLevel logging.LogLevel

	// If not zero, the Node will emit a log entry every period containing statistics about event processing.
	Period time.Duration
}

// eventDispatchStats saves statistical information about the dispatching of events between modules,
// such as the numbers of events dispatched for each module.
type eventDispatchStats struct {
	dispatchCounts map[t.ModuleID]int
	eventCounts    map[t.ModuleID]int
	numDispatches  int
	lastDispatch   time.Time
}

// newDispatchStats returns a new eventDispatchStats object with all counters set to 0
// and the last dispatch time set to the current time.
func newDispatchStats(moduleIDs []t.ModuleID) eventDispatchStats {
	stats := eventDispatchStats{
		dispatchCounts: make(map[t.ModuleID]int),
		eventCounts:    make(map[t.ModuleID]int),
		numDispatches:  0,
		lastDispatch:   time.Now(),
	}
	for _, moduleID := range moduleIDs {
		stats.dispatchCounts[moduleID] = 0
		stats.eventCounts[moduleID] = 0
	}
	return stats
}

// AddDispatch registers the dispatching of an event list between two modules to the statistics.
func (ds *eventDispatchStats) AddDispatch(mID t.ModuleID, numEvents int) {
	ds.numDispatches++
	ds.dispatchCounts[mID]++
	ds.eventCounts[mID] += numEvents
}

// CombinedStats returns the dispatching statistics combined with information about event buffer occupancy
// in a format that can directly be passed to the logger.
// For each module, it includes the number of event lists dispatched to that module (d),
// the total number of events in those lists (e),
// and the number of events still buffered at the input of that module (b).
func (ds *eventDispatchStats) CombinedStats(
	bufferStats map[t.ModuleID]int,
	processingTimes map[t.ModuleID]time.Duration,
	period time.Duration,
) []interface{} {

	// Calculate the column width for the module IDs.
	maxModuleIDLength := 0
	for mID := range ds.eventCounts {
		if len(string(mID)) > maxModuleIDLength {
			maxModuleIDLength = len(string(mID))
		}
	}

	//
	logVals := make([]interface{}, 0, len(ds.eventCounts)+2)
	totalEventsDispatched := 0
	totalEventsBuffered := 0
	maputil.IterateSortedCustom(ds.dispatchCounts, func(mID t.ModuleID, cnt int) (cont bool) {
		// TODO: Pass these as separate values somehow ahd have the logger figure out what to do with them.
		logVals = append(logVals, fmt.Sprintf("%"+fmt.Sprint(maxModuleIDLength)+"s", mID),
			fmt.Sprintf("d(%10d)-e(%10d)-b(%10d)-t(%6.2f%%)\n",
				cnt*int(time.Second)/int(period),
				ds.eventCounts[mID]*int(time.Second)/int(period),
				bufferStats[mID],
				100*float32(processingTimes[mID])/float32(period)),
		)
		totalEventsDispatched += ds.eventCounts[mID]
		totalEventsBuffered += bufferStats[mID]
		return true
	}, func(mID1 t.ModuleID, mID2 t.ModuleID) bool {
		// For now, sort by processing time.
		return processingTimes[mID1] >= processingTimes[mID2]
	})
	logVals = append(logVals,
		"numDispatches/s", ds.numDispatches*int(time.Second)/int(period),
		"totalEventsDispatched/s", totalEventsDispatched*int(time.Second)/int(period),
		"totalEventsBuffered", totalEventsBuffered,
	)
	return logVals
}

// ==============================================================================================================
// Additional methods of Node that deal with stats.
// ==============================================================================================================

// monitorStats prints and resets the dispatching statistics every given time interval, until the node is stopped.
func (n *Node) monitorStats(interval time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(interval)

	for {
		select {
		case <-n.workErrNotifier.ExitC():
			ticker.Stop()
			n.Config.Logger.Log(logging.LevelInfo, "Event processing monitoring finished.")
			n.flushStats()
			return
		case <-ticker.C:
			n.flushStats()
		}
	}
}

func (n *Node) flushStats() {
	n.statsLock.Lock()
	defer n.statsLock.Unlock()

	eventBufferStats := n.pendingEvents.Stats()
	stats := n.dispatchStats.CombinedStats(eventBufferStats, n.resetStopwatches(), n.Config.Stats.Period)

	if n.inputIsPaused() {
		n.Config.Stats.Logger.Log(n.Config.Stats.LogLevel, "External event processing paused.\n", stats...)
	} else {
		n.Config.Stats.Logger.Log(n.Config.Stats.LogLevel, "External event processing running.\n", stats...)
	}

	n.dispatchStats = newDispatchStats(maputil.GetSortedKeys(n.modules))
}

// resetStopwatches resets all the nodes stopwatches tracking the processing time of each module.
// It returns the durations read from the stopwatches just before resetting.
func (n *Node) resetStopwatches() map[t.ModuleID]time.Duration {
	processingTimes := make(map[t.ModuleID]time.Duration, len(n.stopwatches))
	for mID, stopwatch := range n.stopwatches {
		processingTimes[mID] = stopwatch.Reset()
	}
	return processingTimes
}
