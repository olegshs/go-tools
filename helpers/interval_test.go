package helpers

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestInterval_NoOverlap(t *testing.T) {
	var running, maxRunning, calls atomic.Int32

	interval := NewInterval(5*time.Millisecond, func() {
		n := running.Add(1)
		defer running.Add(-1)

		if n > maxRunning.Load() {
			maxRunning.Store(n)
		}
		calls.Add(1)

		time.Sleep(12 * time.Millisecond)
	})

	interval.Start()
	time.Sleep(60 * time.Millisecond)
	interval.Stop()

	if calls.Load() == 0 {
		t.Error("function was not called")
	}
	if maxRunning.Load() > 1 {
		t.Errorf("%d calls overlapped", maxRunning.Load())
	}
}
