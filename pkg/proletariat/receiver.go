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
	"context"
	"io"
	"sync"
	"time"
)

const defaultConsumptionTimeout = time.Millisecond

// IReceiver should handle all messages received by the connections.
//
// When publishing back a message to be consumed by the client,
// not necessarily it will be consumed directly, so the receiver will
// keep track of received message to publish it only once.
type IReceiver interface {
	io.Closer

	// Start will start polling to deliver messages.
	Start()

	// AddResponse insert a new message to be delivered.
	AddResponse(Datagram)
}

// A receiver implementation that will keeps messages stored into a
// FIFO queue. If the client does not consume the messages this queue
// will only increase.
//
// Since messages can be tried to be delivered indefinitely, we can be stuck
// with a message, delaying everything.
// TODO: configurable TTL for in-memory messages.
type fifoReceiver struct {
	// FIFO queue holding messages to be delivered.
	queue Queue

	// Used to bound asynchronous commands.
	timeout time.Duration

	// Channel to publish all received messages.
	deliver chan<- Datagram

	// Used to close the sending channel only once.
	once *sync.Once

	// Receiver context.
	ctx context.Context

	// Receiver cancel to stop working.
	cancel context.CancelFunc
}

func NewReceiver(deliver chan<- Datagram, timeout time.Duration, parent context.Context) IReceiver {
	ctx, cancel := context.WithCancel(parent)
	return &fifoReceiver{
		queue:   NewQueue(),
		timeout: timeout,
		once:    &sync.Once{},
		deliver: deliver,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Try to send the given datagram through the channel. If successfully consumed
// returns true, false otherwise.
//
// If the context is already closed, we will return `true`, since is not important
// to add the element to the buffer anymore.
func (f *fifoReceiver) tryDeliver(datagram Datagram) bool {
	defer func() {
		recover()
	}()

	if f.timeout <= 0 {
		select {
		case <-f.ctx.Done():
			return true
		case f.deliver <- datagram:
			return true
		default:
			return false
		}
	}

	select {
	case <-f.ctx.Done():
		return true
	case f.deliver <- datagram:
		return true
	case <-time.After(f.timeout):
		return false
	}
}

// Will try to deliver the element in the head of the queue.
// If successfully consumed it will be removed.
func (f *fifoReceiver) peekAndTryDeliver() {
	head := f.queue.Peek()
	if head == nil {
		return
	}
	if f.tryDeliver(head.(Datagram)) {
		f.queue.Pop()
	}
}

// Close the current fifo receiver.
func (f *fifoReceiver) Close() error {
	f.cancel()
	return nil
}

// Start will execute while the application is running.
// This method is responsible for delivering datagrams that were
// not sent directly and are present on the buffer.
func (f *fifoReceiver) Start() {
	for {
		select {
		case <-f.ctx.Done():
			return
		case d := <-f.queue.HasElements():
			if f.tryDeliver(d.(Datagram)) {
				f.queue.Pop()
			}
		default:
			f.peekAndTryDeliver()
		}
	}
}

// AddResponse handle a new received response.
// First it will try to directly deliver the datagram, if is not possible
// the datagram will be stored on a buffer to be delivered later.
//
// In this approach, some messages can be delivered out of order.
// TODO: do we care to fix message order?
func (f *fifoReceiver) AddResponse(datagram Datagram) {
	if !f.tryDeliver(datagram) {
		f.queue.Append(datagram)
	}
}
