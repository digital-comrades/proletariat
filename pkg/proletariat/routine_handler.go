// Copyright (C) 2020-2021 digital-comrades and others.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//         http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package proletariat

import (
	"sync"
	"time"
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
	return &GoRoutineHandler{
		mutex:   &sync.Mutex{},
		working: true,
		group:   &sync.WaitGroup{},
	}
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
