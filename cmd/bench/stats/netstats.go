package stats

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/filecoin-project/mir/pkg/util/maputil"
)

type NetStats struct {
	sync.Mutex

	// A set of all labels that ever occurred.
	labels map[string]struct{}

	// Data.
	// key: Time since start of measurement (in multiples of SamplingPeriod)
	// value: map associating a string label with the number of transactions sent/received/dropped
	//        under the label in the time slot
	SentData          map[string]map[time.Duration]int `json:"Sent"`
	ReceivedData      map[string]map[time.Duration]int `json:"Received"`
	DroppedData       map[string]map[time.Duration]int `json:"Dropped"`
	TotalSentData     map[time.Duration]int            `json:"TotalSent"`
	TotalReceivedData map[time.Duration]int            `json:"TotalReceived"`
	TotalDroppedData  map[time.Duration]int            `json:"TotalDropped"`

	// Time of the start of measurements.
	started   bool
	startTime time.Time

	// Duration of the measurement.
	duration time.Duration

	SamplingPeriod time.Duration
}

func NewNetStats(
	samplingPeriod time.Duration,
) *NetStats {
	return &NetStats{
		labels:            make(map[string]struct{}),
		SentData:          make(map[string]map[time.Duration]int),
		ReceivedData:      make(map[string]map[time.Duration]int),
		DroppedData:       make(map[string]map[time.Duration]int),
		TotalSentData:     make(map[time.Duration]int),
		TotalReceivedData: make(map[time.Duration]int),
		TotalDroppedData:  make(map[time.Duration]int),
		SamplingPeriod:    samplingPeriod,
		started:           false,
	}
}

func (ns *NetStats) ToJSON() ([]byte, error) {
	return json.Marshal(ns)
}

func (ns *NetStats) Start() {
	ns.Lock()
	defer ns.Unlock()

	now := time.Now()
	ns.startTime = now
	ns.duration = ns.SamplingPeriod
	ns.started = true
}

func (ns *NetStats) Sent(nBytes int, label string) {
	ns.recordData(ns.SentData, ns.TotalSentData, nBytes, label)
}

func (ns *NetStats) Received(nBytes int, label string) {
	ns.recordData(ns.ReceivedData, ns.TotalReceivedData, nBytes, label)
}

func (ns *NetStats) Dropped(nBytes int, label string) {
	ns.recordData(ns.DroppedData, ns.TotalDroppedData, nBytes, label)
}

func (ns *NetStats) recordData(
	data map[string]map[time.Duration]int,
	totalData map[time.Duration]int,
	nBytes int,
	label string,
) {
	ns.Lock()
	defer ns.Unlock()

	ns.createLabel(label)

	// Get time since start of measurement and round it to the next lower step.
	// Messages that arrive before Start() is called on NetStats are considered to have arrived at time zero.
	t := time.Duration(0)
	if ns.started {
		t = time.Since(ns.startTime)
		t = (t / ns.SamplingPeriod) * ns.SamplingPeriod
	}

	// Fill periods with no delivered transactions with explicit zeroes.
	// This is redundant, as it can always be inferred from the rest of the data,
	// but makes it a bit more convenient to work with when iterating over the (sorted) items.
	ns.fillAtDuration(t)
	data[label][t] += nBytes
	totalData[t] += nBytes

}

func (ns *NetStats) createLabel(label string) {
	_, ok := ns.labels[label]
	if !ok {
		fmt.Printf("Creating label: %s\n", label)
		ns.SentData[label] = make(map[time.Duration]int)
		ns.ReceivedData[label] = make(map[time.Duration]int)
		ns.DroppedData[label] = make(map[time.Duration]int)
		ns.labels[label] = struct{}{}

		for t := time.Duration(0); t <= ns.duration; t += ns.SamplingPeriod {
			ns.SentData[label][t] = 0
			ns.ReceivedData[label][t] = 0
			ns.DroppedData[label][t] = 0
		}
	}
}

// Fill adds padding to DeliveredTxs. In case no transactions have been delivered for some time,
// Fill adds zero values for these time slots.
// Fill should be called once more at the very end of data collection,
// especially if no transactions have been delivered at all. In such a case, DeliveredTxs would otherwise stay empty
// and not represent the true result of data collection (namely zeroes for all the duration.)
func (ns *NetStats) Fill() {
	ns.Lock()
	defer ns.Unlock()

	ns.fillAtDuration(time.Since(ns.startTime))
}

func (ns *NetStats) fillAtDuration(t time.Duration) {
	for t >= ns.duration+ns.SamplingPeriod {
		ns.duration += ns.SamplingPeriod
		ns.TotalSentData[ns.duration] = 0
		ns.TotalReceivedData[ns.duration] = 0
		ns.TotalDroppedData[ns.duration] = 0
		for _, label := range maputil.GetKeys(ns.labels) {
			ns.SentData[label][ns.duration] = 0
			ns.ReceivedData[label][ns.duration] = 0
			ns.DroppedData[label][ns.duration] = 0
		}
	}
}

func (ns *NetStats) AvgSendRate(label string) int {
	return ns.avgDataRate(ns.SentData[label])
}

func (ns *NetStats) AvgReceiveRate(label string) int {
	return ns.avgDataRate(ns.ReceivedData[label])
}

func (ns *NetStats) AvgDropRate(label string) int {
	return ns.avgDataRate(ns.DroppedData[label])
}

func (ns *NetStats) AvgTotalSendRate() int {
	return ns.avgDataRate(ns.TotalSentData)
}

func (ns *NetStats) AvgTotalReceiveRate() int {
	return ns.avgDataRate(ns.TotalReceivedData)
}

func (ns *NetStats) AvgTotalDropRate() int {
	return ns.avgDataRate(ns.TotalDroppedData)
}

func (ns *NetStats) avgDataRate(data map[time.Duration]int) int {
	if ns.duration == 0 {
		return 0
	}

	totalData := 0
	for _, d := range data {
		totalData += d
	}
	return totalData / int(ns.duration.Seconds())
}

func (ns *NetStats) AvgSendRateAll() int {
	return ns.avgDataRateAll(ns.SentData)
}

func (ns *NetStats) AvgReceiveRateAll() int {
	return ns.avgDataRateAll(ns.ReceivedData)
}

func (ns *NetStats) AvgDropRateAll() int {
	return ns.avgDataRateAll(ns.DroppedData)
}

func (ns *NetStats) avgDataRateAll(data map[string]map[time.Duration]int) int {
	if len(ns.labels) == 0 {
		return 0
	}

	// Compute the average of the average data rates for all labels.
	// Note that we can afford doing this, as the duration is the same for all labels,
	// and thus all the averaged values have the same "weight".
	total := 0
	for label := range ns.labels {
		total += ns.avgDataRate(data[label])
	}
	return total
}

func (ns *NetStats) WriteCSVHeader(w *csv.Writer) error {
	record := []string{
		fmt.Sprintf(fmt.Sprintf("%%%ds", width), "duration"),
		fmt.Sprintf(fmt.Sprintf("%%%ds", width), "upload-MiB/s"),
		fmt.Sprintf(fmt.Sprintf("%%%ds", width), "download-MiB/s"),
		fmt.Sprintf(fmt.Sprintf("%%%ds", width), "loss-MiB/s"),
	}
	return w.Write(record)
}

func (ns *NetStats) WriteCSVRecord(w *csv.Writer) error {

	record := []string{
		fmt.Sprintf(fmt.Sprintf("%%%ds", width), ns.duration),
		fmt.Sprintf(fmt.Sprintf("%%%d.3f", width), float64(ns.AvgTotalSendRate())/(1024*1024)),
		fmt.Sprintf(fmt.Sprintf("%%%d.3f", width), float64(ns.AvgTotalReceiveRate())/(1024*1024)),
		fmt.Sprintf(fmt.Sprintf("%%%d.3f", width), float64(ns.AvgTotalDropRate())/(1024*1024)),
	}
	return w.Write(record)
}
