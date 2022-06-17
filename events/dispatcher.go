// Пакет events реализует диспетчер событий.
package events

import (
	"reflect"
	"sync"

	"github.com/olegshs/go-tools/helpers/call"
)

type Event string

type Callback interface{}

type Dispatcher struct {
	events map[Event][]reflect.Value
	mutex  sync.RWMutex
}

func New() *Dispatcher {
	d := Dispatcher{
		map[Event][]reflect.Value{},
		sync.RWMutex{},
	}
	return &d
}

func (d *Dispatcher) AddListener(e Event, f Callback) {
	rv := d.value(f)

	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.events[e] = append(d.events[e], rv)
}

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
