package helpers

import (
	"math/rand"
	"sync/atomic"
	"time"
)

// Interval runs a function periodically in the background, similar to
// time.Ticker but with optional random jitter and protection against
// overlapping runs.
type Interval struct {
	duration       time.Duration
	maxRandomDelay time.Duration

	function         func()
	isExecuting      atomic.Bool // set while a call runs; read by the timer goroutine
	allowOverlapping bool

	isRunning bool
	stop      chan int
}

// NewInterval creates a stopped Interval that calls f every d.
// Call Start to begin.
func NewInterval(d time.Duration, f func()) *Interval {
	interval := new(Interval)
	interval.duration = d
	interval.maxRandomDelay = 0
	interval.function = f
	interval.allowOverlapping = false
	interval.isRunning = false
	interval.stop = make(chan int)

	return interval
}

// Overlapping sets whether a new call may start while the previous one is
// still running. By default overlapping calls are skipped.
func (interval *Interval) Overlapping(allow bool) *Interval {
	interval.allowOverlapping = allow
	return interval
}

// RandomDelay adds a random delay in [0, maxRandomDelay) to every tick, so
// that many instances started at the same time do not fire together.
func (interval *Interval) RandomDelay(maxRandomDelay time.Duration) *Interval {
	interval.maxRandomDelay = maxRandomDelay
	return interval
}

// Start begins calling the function in a background goroutine. Each call
// runs in its own goroutine. Ticks are scheduled from the start time rather
// than from the end of the previous call, so the period does not drift.
// Start does nothing if the interval is already running or its duration
// is not positive.
func (interval *Interval) Start() {
	if interval.isRunning || interval.duration <= 0 {
		return
	}
	interval.isRunning = true

	t := time.Now()
	r := rand.New(rand.NewSource(t.UnixNano() + rand.Int63()))

	go func() {
		for {
			t = t.Add(interval.duration)

			var d time.Duration
			if interval.maxRandomDelay > 0 {
				d = time.Duration(r.Int63n(int64(interval.maxRandomDelay)))
			}

			time.Sleep(time.Until(t) + d)

			select {
			case <-interval.stop:
				return
			default:
				if interval.isExecuting.Load() && !interval.allowOverlapping {
					continue
				}
				interval.isExecuting.Store(true)

				// The function runs in its own goroutine so that a slow call does not
				// delay the schedule.
				go func() {
					interval.function()
					interval.isExecuting.Store(false)
				}()
			}
		}
	}()
}

// Stop stops the interval. It blocks until the background goroutine wakes
// for its next tick, which can take up to one full period. A call that is
// already running is not interrupted.
func (interval *Interval) Stop() {
	if !interval.isRunning {
		return
	}

	interval.stop <- 1
	interval.isRunning = false
}
