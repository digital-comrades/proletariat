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
	"bytes"
	"context"
	"github.com/digital-comrades/proletariat/pkg/proletariat"
	"go.uber.org/goleak"
	"sync"
	"testing"
	"time"
)

func Test_FIFOQueue(t *testing.T) {
	defer goleak.VerifyNone(t)
	q := proletariat.NewQueue()
	var wg sync.WaitGroup
	wg.Add(1)

	for i := 0; i < 1000; i++ {
		q.Append(i)
	}

	go func() {
		defer wg.Done()
		i := 0
		e := q.Pop()
		for ; e != nil; e = q.Pop() {
			if e != i {
				t.Errorf("expected %d found %d", i, e)
			}
			i++
		}
	}()

	wg.Wait()
}

func Test_ShouldAddAndReadLater(t *testing.T) {
	defer goleak.VerifyNone(t)
	testSize := 1024
	read := make(chan proletariat.Datagram)
	ctx, cancel := context.WithCancel(context.TODO())
	r := proletariat.NewReceiver(read, 0, ctx)

	go r.Start()

	for i := 0; i < testSize; i++ {
		data := []byte{byte(i)}
		r.AddResponse(proletariat.Datagram{Data: bytes.NewBuffer(data)})
	}

	for i := 0; i < testSize; i++ {
		expected := []byte{byte(i)}
		select {
		case <-time.After(time.Second):
			t.Error("did not read messages")
			goto C
		case d := <-read:
			if !bytes.Equal(expected, d.Data.Bytes()) {
				t.Errorf("expected %#v and found %#v", expected, d.Data.Bytes())
			}
		}
	}

C:
	cancel()
	if err := r.Close(); err != nil {
		t.Errorf("failed closing. %v", err)
	}
}
