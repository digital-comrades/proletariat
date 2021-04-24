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
	"math/rand"
	"sync"
)

// Queue interface to create simple functions for a queue.
type Queue interface {
	// Append insert a new element to the queue.
	Append(interface{})

	// Peek verify the head of the queue, without removing.
	Peek() interface{}

	// Pop remove the fist element, nil if empty.
	Pop() interface{}

	// HasElements channel to be notified when new elements are added.
	HasElements() <-chan interface{}
}

// Buffer will hold received messages in a FIFO queue.
type buffer struct {
	// Holds the identifiers for each element present on the
	// buffer. The elements are stored in the same order they
	// were added.
	buf []int64

	// Holds the actual elements. The buf holds the identifier
	// to locate the element in the map.
	data map[int64]interface{}

	// ReadWrite mutex to safely execute actions.
	mutex *sync.RWMutex

	// Channel to notify about elements.
	hasElements chan interface{}
}

// NewQueue create a new queue.
func NewQueue() Queue {
	return &buffer{
		data:        make(map[int64]interface{}),
		mutex:       &sync.RWMutex{},
		hasElements: make(chan interface{}),
	}
}

// Generate a new identifier that is no currently being used.
func (b *buffer) newId() int64 {
	for {
		id := rand.Int63()
		_, ok := b.data[id]
		if id != 0 && !ok {
			return id
		}
	}
}

// Sends a notification through the channel if the buffer have elements.
func (b *buffer) notify() {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	if len(b.buf) > 0 {
		data := b.data[b.buf[0]]
		select {
		case b.hasElements <- data:
		default:
		}
	}
}

// Append insert a new element to the end of the buffer.
func (b *buffer) Append(i interface{}) {
	defer b.notify()
	b.mutex.Lock()
	defer b.mutex.Unlock()

	id := b.newId()
	b.data[id] = i
	b.buf = append(b.buf, id)
}

// Peek verify the head of the queue. If the queue does
// not have elements, nil is returned.
func (b *buffer) Peek() interface{} {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	if len(b.buf) == 0 {
		return nil
	}
	return b.data[b.buf[0]]
}

// HasElements listen for the channel.
func (b *buffer) HasElements() <-chan interface{} {
	return b.hasElements
}

// Pop remove the head of the queue. If there is not element,
// nil is returned.
func (b *buffer) Pop() interface{} {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if len(b.buf) == 0 {
		return nil
	}

	id := b.buf[0]
	data := b.data[id]
	delete(b.data, id)
	// Is this destructuring expensive?
	b.buf = append(b.buf[:0], b.buf[1:]...)
	return data
}
