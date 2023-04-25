// Copyright Contributors to the Mir project
//
// SPDX-License-Identifier: Apache-2.0

package stats

import (
	"encoding/csv"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/filecoin-project/mir/pkg/pb/trantorpb"
)

type Stats struct {
	lock                sync.RWMutex
	reqTimestamps       map[reqKey]time.Time
	avgLatency          float64
	timestampedRequests int
	deliveredRequests   int
}

type reqKey struct {
	ClientID string
	ReqNo    uint64
}

func NewStats() *Stats {
	return &Stats{
		reqTimestamps: make(map[reqKey]time.Time),
	}
}

func (s *Stats) NewRequest(req *trantorpb.Transaction) {
	s.lock.Lock()
	k := reqKey{req.ClientId, req.TxNo}
	s.reqTimestamps[k] = time.Now()
	s.lock.Unlock()
}

func (s *Stats) Delivered(req *trantorpb.Transaction) {
	s.lock.Lock()
	s.deliveredRequests++
	k := reqKey{req.ClientId, req.TxNo}
	if t, ok := s.reqTimestamps[k]; ok {
		delete(s.reqTimestamps, k)
		s.timestampedRequests++
		d := time.Since(t)

		// $CA_{n+1} = CA_n + {x_{n+1} - CA_n \over n + 1}$
		s.avgLatency += (float64(d) - s.avgLatency) / float64(s.timestampedRequests)
	}
	s.lock.Unlock()
}

func (s *Stats) AvgLatency() time.Duration {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return time.Duration(s.avgLatency)
}

func (s *Stats) DeliveredRequests() int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.deliveredRequests
}

func (s *Stats) Reset() {
	s.lock.Lock()
	s.avgLatency = 0
	s.timestampedRequests = 0
	s.deliveredRequests = 0
	s.lock.Unlock()
}

func (s *Stats) WriteCSVHeader(w *csv.Writer) {
	record := []string{
		"nrDelivered",
		"tps",
		"avgLatency",
	}
	_ = w.Write(record)
}

func (s *Stats) WriteCSVRecord(w *csv.Writer, d time.Duration) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	tps := float64(s.deliveredRequests) / (float64(d) / float64(time.Second))
	record := []string{
		strconv.Itoa(s.deliveredRequests),
		strconv.Itoa(int(tps)),
		fmt.Sprintf("%.2f", time.Duration(s.avgLatency).Seconds()),
	}
	_ = w.Write(record)
}
