package mir

import (
	"sync"
	"time"
)

type Stopwatch struct {
	sync.Mutex

	t       time.Duration
	started time.Time
	running bool
}

func (s *Stopwatch) Start() {
	s.Lock()
	defer s.Unlock()

	if s.running {
		panic("stopwatch already running")
	}
	s.running = true
	s.started = time.Now()
}

func (s *Stopwatch) Stop() {
	s.Lock()
	defer s.Unlock()

	if s.running {
		s.t += time.Since(s.started)
		s.running = false
	}
}

func (s *Stopwatch) Reset() time.Duration {
	s.Lock()
	defer s.Unlock()

	// This code is identical to the one in s.Stop().
	// We cannot simply call stop here, because we are already holding the lock.
	// Calling stop before acquiring the lock would, on the other hand, make Reset not atomic.
	if s.running {
		s.t += time.Since(s.started)
		s.running = false
	}

	t := s.t
	s.t = 0
	return t
}

func (s *Stopwatch) Read() time.Duration {
	s.Lock()
	defer s.Unlock()

	return s.t
}
