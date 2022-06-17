package helpers

import (
	"math/rand"
	"time"
)

type Interval struct {
	duration       time.Duration
	maxRandomDelay time.Duration

	function         func()
	isExecuting      bool
	allowOverlapping bool

	isRunning bool
	stop      chan int
}

func NewInterval(d time.Duration, f func()) *Interval {
	interval := new(Interval)
	interval.duration = d
	interval.maxRandomDelay = 0
	interval.function = f
	interval.isExecuting = false
	interval.allowOverlapping = false
	interval.isRunning = false
	interval.stop = make(chan int)

	return interval
}

func (interval *Interval) Overlapping(allow bool) *Interval {
	interval.allowOverlapping = allow
	return interval
}

func (interval *Interval) RandomDelay(maxRandomDelay time.Duration) *Interval {
	interval.maxRandomDelay = maxRandomDelay
	return interval
}

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
				if interval.isExecuting && !interval.allowOverlapping {
					continue
				}
				interval.isExecuting = true

				go func() {
					interval.function()
					interval.isExecuting = false
				}()
			}
		}
	}()
}

func (interval *Interval) Stop() {
	if !interval.isRunning {
		return
	}

	interval.stop <- 1
	interval.isRunning = false
}
