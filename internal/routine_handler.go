package internal

import (
	"sync"
	"time"
)

var (
	// Ensure thread safety while creating a new GoRoutineHandler.
	create = sync.Once{}

	// Global instance for the handler.
	handler *GoRoutineHandler
)

// GoRoutineHandler is responsible for handling goroutines.
// This is used so go routines do not leak and are spawned without any control.
// Using the handler to spawn new routines will guarantee that any routine that
//is not controller careful will be known when the application finishes.
type GoRoutineHandler struct {
	mutex *sync.Mutex

	working bool

	group *sync.WaitGroup
}

// Create a singleton instance for the GoRoutineHandler struct.
// This is a singleton to ensure that throughout the application exists
// only one single point where go routines are spawned, thus avoiding a leak.
func NewRoutineHandler() *GoRoutineHandler {
	create.Do(func() {
		handler = &GoRoutineHandler{
			mutex:   &sync.Mutex{},
			working: true,
			group:   &sync.WaitGroup{},
		}
	})
	return handler
}

// This method will increase the size of the group count and spawn
// the new go routine. After the routine is done, the group will be decreased.
//
// This method will panic if the handler is already closed.
func (h *GoRoutineHandler) Spawn(f func()) {
	h.mutex.Lock()
	if !h.working {
		panic("go routine handler closed!")
	}
	h.mutex.Unlock()

	h.group.Add(1)
	go func() {
		defer h.group.Done()
		f()
	}()
}

func (h *GoRoutineHandler) WithTimeout(f func(), timeout time.Duration) bool {
	done := make(chan bool)
	h.Spawn(func() {
		f()
		done <- true
	})
	select {
	case <-done:
		return true
	case <-time.After(timeout):
		return false
	}
}

// Blocks while waiting for go routines to stop. This will set the
// working mode to off, so after this is called any spawned go routine will panic.
func (h *GoRoutineHandler) Close() {
	h.mutex.Lock()
	h.working = false
	h.mutex.Unlock()
	h.group.Wait()
}
