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
	lock                    sync.RWMutex
	txTimestamps            map[txKey]time.Time
	avgLatency              float64
	timestampedTransactions int
	deliveredTransactions   int
}

type txKey struct {
	ClientID string
	TxNo     uint64
}

func NewStats() *Stats {
	return &Stats{
		txTimestamps: make(map[txKey]time.Time),
	}
}

func (s *Stats) NewTX(tx *trantorpb.Transaction) {
	s.lock.Lock()
	k := txKey{tx.ClientId, tx.TxNo}
	s.txTimestamps[k] = time.Now()
	s.lock.Unlock()
}

func (s *Stats) Delivered(tx *trantorpb.Transaction) {
	s.lock.Lock()
	s.deliveredTransactions++
	k := txKey{tx.ClientId, tx.TxNo}
	if t, ok := s.txTimestamps[k]; ok {
		delete(s.txTimestamps, k)
		s.timestampedTransactions++
		d := time.Since(t)

		// $CA_{n+1} = CA_n + {x_{n+1} - CA_n \over n + 1}$
		s.avgLatency += (float64(d) - s.avgLatency) / float64(s.timestampedTransactions)
	}
	s.lock.Unlock()
}

func (s *Stats) AvgLatency() time.Duration {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return time.Duration(s.avgLatency)
}

func (s *Stats) DeliveredTransactions() int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.deliveredTransactions
}

func (s *Stats) Reset() {
	s.lock.Lock()
	s.avgLatency = 0
	s.timestampedTransactions = 0
	s.deliveredTransactions = 0
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

	tps := float64(s.deliveredTransactions) / (float64(d) / float64(time.Second))
	record := []string{
		strconv.Itoa(s.deliveredTransactions),
		strconv.Itoa(int(tps)),
		fmt.Sprintf("%.2f", time.Duration(s.avgLatency).Seconds()),
	}
	_ = w.Write(record)
}
