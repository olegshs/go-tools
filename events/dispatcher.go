// Package events implements a simple event dispatcher.
//
// Listeners are functions of any signature. When an event is dispatched,
// its arguments are converted to the listener's parameter types with
// helpers/call, and extra arguments are dropped, so a listener can accept
// only the leading arguments it needs:
//
//	d := events.New()
//	d.AddListener("Saved", func(id int64) { ... })
//	d.DispatchSync("Saved", id, time.Now())
//
// A listener whose parameters cannot be filled from the arguments panics
// when the event is dispatched.
package events

import (
	"reflect"
	"sync"

	"github.com/olegshs/go-tools/helpers/call"
)

// Event is the name of an event.
type Event string

// Callback is a listener function. It may have any signature.
type Callback interface{}

// Dispatcher keeps a list of listeners per event. It is safe for
// concurrent use.
type Dispatcher struct {
	events map[Event][]reflect.Value
	mutex  sync.RWMutex
}

// New creates a dispatcher without listeners.
func New() *Dispatcher {
	d := Dispatcher{
		map[Event][]reflect.Value{},
		sync.RWMutex{},
	}
	return &d
}

// AddListener adds f as a listener for e. It panics if f is not a function.
func (d *Dispatcher) AddListener(e Event, f Callback) {
	rv := d.value(f)

	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.events[e] = append(d.events[e], rv)
}

// RemoveListener removes f from the listeners of e. f must be the same
// function value that was passed to AddListener. A method value such as
// obj.Method is a new function each time it is evaluated, so store it in a
// variable and pass that variable to both calls.
func (d *Dispatcher) RemoveListener(e Event, f Callback) {
	rv := d.value(f)

	d.mutex.Lock()
	defer d.mutex.Unlock()

	a := make([]reflect.Value, 0, len(d.events[e]))
	for _, f := range d.events[e] {
		if f != rv {
			a = append(a, f)
		}
	}

	if len(a) == 0 {
		delete(d.events, e)
		return
	}

	d.events[e] = a
}

// Dispatch calls every listener of e with args, each in its own goroutine,
// and returns immediately. The returned channel receives the number of
// listeners once all of them have returned. Receiving from it is optional.
//
// A listener that cannot be called with args panics in its goroutine,
// which crashes the program.
func (d *Dispatcher) Dispatch(e Event, args ...interface{}) chan int {
	a := d.listeners(e)
	n := len(a)

	c := make(chan int, n)
	for i, f := range a {
		go func(i int, f reflect.Value) {
			_, err := call.Func(f, args...)
			if err != nil {
				panic(err)
			}

			c <- i
		}(i, f)
	}

	done := make(chan int)
	go func() {
		for i := 0; i < n; i++ {
			<-c
		}
		done <- n
	}()

	return done
}

// DispatchSync calls every listener of e with args, one after another in
// the calling goroutine, and returns the number of listeners. It panics if
// a listener cannot be called with args.
func (d *Dispatcher) DispatchSync(e Event, args ...interface{}) int {
	a := d.listeners(e)
	n := len(a)

	for _, f := range a {
		_, err := call.Func(f, args...)
		if err != nil {
			panic(err)
		}
	}

	return n
}

// listeners returns a copy of the listeners of e, so they can be called
// without holding the lock.
func (d *Dispatcher) listeners(e Event) []reflect.Value {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	n := len(d.events[e])
	a := make([]reflect.Value, n)

	for i, f := range d.events[e] {
		a[i] = f
	}

	return a
}

// value returns the reflect.Value of the function f, dereferencing
// pointers. It panics if f is not a function.
func (d *Dispatcher) value(f Callback) reflect.Value {
	rv := reflect.ValueOf(f)
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Func {
		panic(rv.Kind().String() + " is not a function")
	}

	return rv
}
