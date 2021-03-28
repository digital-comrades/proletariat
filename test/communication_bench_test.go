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

package test

import (
	"context"
	"github.com/digital-comrades/proletariat/pkg/proletariat"
	"sync"
	"testing"
)

func sendMultipleMessagesBench(first, second proletariat.Communication, content []byte, testSize int, b *testing.B) {
	addr := proletariat.Address(first.Addr().String())
	for i := 0; i < testSize; i++ {
		err := second.Send(addr, content)
		if err != nil {
			b.Errorf("failed writing. %s", err.Error())
		}
	}
}

func Benchmark_CommunicationMessages(b *testing.B) {
	testSize := 1024
	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.TODO())
	first, err := proletariat.NewCommunication(proletariat.Configuration{
		Address: "127.0.0.1:0",
		Timeout: 0,
		Ctx:     ctx,
	})
	if err != nil {
		b.Fatalf("failed tcp one: %v", err)
	}

	second, err := proletariat.NewCommunication(proletariat.Configuration{
		Address:  "127.0.0.1:0",
		Timeout:  0,
		Ctx:      ctx,
		PoolSize: 10,
	})
	if err != nil {
		b.Fatalf("failed tcp two: %v", err)
	}

	go first.Start()
	go second.Start()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for d := range first.Receive() {
			if d.Err != nil && IsClosedError(d.Err) {
				return
			}
		}
	}()

	content := []byte("hello world")
	for i := 0; i < b.N; i++ {
		sendMultipleMessagesBench(first, second, content, testSize, b)
	}

	cancel()

	if err := first.Close(); err != nil {
		b.Errorf("failed closing first. %s", err.Error())
	}

	if err := second.Close(); err != nil {
		b.Errorf("failed closing second. %s", err.Error())
	}

	wg.Wait()
}

func Benchmark_CommunicationParallelMessages(b *testing.B) {
	testSize := 1024
	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.TODO())
	first, err := proletariat.NewCommunication(proletariat.Configuration{
		Address: "127.0.0.1:0",
		Timeout: 0,
		Ctx:     ctx,
	})
	if err != nil {
		b.Fatalf("failed tcp one: %v", err)
	}

	second, err := proletariat.NewCommunication(proletariat.Configuration{
		Address:  "127.0.0.1:0",
		Timeout:  0,
		Ctx:      ctx,
		PoolSize: 10,
	})
	if err != nil {
		b.Fatalf("failed tcp two: %v", err)
	}

	go first.Start()
	go second.Start()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for d := range first.Receive() {
			if d.Err != nil && IsClosedError(d.Err) {
				return
			}
		}
	}()

	content := []byte("hello world")
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sendMultipleMessagesBench(first, second, content, testSize, b)
		}
	})

	cancel()

	if err := first.Close(); err != nil {
		b.Errorf("failed closing first. %s", err.Error())
	}

	if err := second.Close(); err != nil {
		b.Errorf("failed closing second. %s", err.Error())
	}

	wg.Wait()
}
