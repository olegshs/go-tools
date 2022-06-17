package events

import (
	"os"
	"os/signal"
	"syscall"
)

var (
	Shared = New()

	EventExit = Event("Exit")
)

func DispatchExit() {
	done := Shared.Dispatch(EventExit)
	<-done
}

func DispatchExitOnInterrupt() {
	c := make(chan os.Signal)
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
