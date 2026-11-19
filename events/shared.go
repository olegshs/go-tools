package events

import (
	"os"
	"os/signal"
	"syscall"
)

// Shared is the application-wide dispatcher. EventExit is dispatched on it
// when the application shuts down, so that components can flush buffers
// and release resources.
var (
	Shared = New()

	EventExit = Event("Exit")
)

// DispatchExit dispatches EventExit on Shared and waits for all listeners
// to return.
func DispatchExit() {
	done := Shared.Dispatch(EventExit)
	<-done
}

// DispatchExitOnInterrupt makes the process call DispatchExit when it
// receives SIGINT or SIGTERM, and then exit with status 128 + signal
// number, as shells do.
func DispatchExitOnInterrupt() {
	// signal.Notify does not block, so a signal that arrives before the
	// goroutine is ready would be lost with an unbuffered channel.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		s := <-c

		DispatchExit()

		code := 0
		if s, ok := s.(syscall.Signal); ok {
			code = int(s)
		}
		os.Exit(128 + code)
	}()
}
