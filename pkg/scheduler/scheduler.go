// pkg/scheduler/scheduler.go
// In-memory delayed task scheduler — replaces Celery countdown tasks.
// Uses time.AfterFunc under the hood; goroutine is spawned by the runtime.
//
// NOTE: tasks are lost on server restart (no persistence).
// This matches the original Celery setup which used Redis as a volatile broker.
package scheduler

import (
	"sync"
	"time"
)

// Scheduler manages named delayed tasks.
// Each task is identified by a string key; scheduling the same key again
// cancels the previous pending task (replaces celery revoke + apply_async).
type Scheduler struct {
	mu     sync.Mutex
	timers map[string]*time.Timer
}

func New() *Scheduler {
	return &Scheduler{timers: make(map[string]*time.Timer)}
}

// Schedule cancels any pending task for key, then schedules fn to run after delay.
func (s *Scheduler) Schedule(key string, delay time.Duration, fn func()) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if t, ok := s.timers[key]; ok {
		t.Stop()
	}

	s.timers[key] = time.AfterFunc(delay, func() {
		s.mu.Lock()
		delete(s.timers, key)
		s.mu.Unlock()
		fn()
	})
}

// Cancel stops a pending task for key (no-op if not found).
func (s *Scheduler) Cancel(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if t, ok := s.timers[key]; ok {
		t.Stop()
		delete(s.timers, key)
	}
}

// Pending returns true if a task is waiting for key.
func (s *Scheduler) Pending(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.timers[key]
	return ok
}
