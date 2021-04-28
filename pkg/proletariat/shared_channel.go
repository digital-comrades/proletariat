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
	"sync"
)

// SharedChannel is structure that holds a channel that can be
// shared across multiple goroutines, without danger of publishing
// to a closed channel nor data race while publishing and closing.
type SharedChannel struct {
	mutex *sync.Mutex
	flag  Flag
	sc    chan Datagram
}

func NewSharedChannel() *SharedChannel {
	return &SharedChannel{
		mutex: &sync.Mutex{},
		flag:  Flag{},
		sc:    make(chan Datagram, 1024),
	}
}

// Publish will try to publish the message, if successful will return `true`,
// if the channel is closed or the context is done, returns `false`.
func (s *SharedChannel) Publish(ctx context.Context, datagram Datagram) bool {
	if s.flag.IsActive() {
		s.mutex.Lock()
		defer s.mutex.Unlock()
		select {
		case <-ctx.Done():
		case s.sc <- datagram:
			return true
		}
	}
	return false
}

// Consume will return the channel to listen to messages.
func (s *SharedChannel) Consume() <-chan Datagram {
	return s.sc
}

// Close will close the channel.
func (s *SharedChannel) Close() error {
	if s.flag.Inactivate() {
		s.mutex.Lock()
		defer s.mutex.Unlock()
		close(s.sc)
	}
	return nil
}
