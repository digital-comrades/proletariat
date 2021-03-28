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
	"go.uber.org/goleak"
	"sync"
	"testing"
	"time"
)

func TestTCPCommunication_CreateAndSend(t *testing.T) {
	defer goleak.VerifyNone(t)
	tcpOneAddr := proletariat.Address("127.0.0.1:11111")
	tcpTwoAddr := proletariat.Address("127.0.0.1:22222")
	ctxOne, cancelOne := context.WithCancel(context.TODO())
	ctxTwo, cancelTwo := context.WithCancel(context.TODO())
	testTimeout := 500 * time.Millisecond

	commOne, err := proletariat.NewCommunication(proletariat.Configuration{
		Address: tcpOneAddr,
		Timeout: 0,
		Ctx:     ctxOne,
	})
	if err != nil {
		t.Fatalf("failed tcp one: %v", err)
	}

	commTwo, err := proletariat.NewCommunication(proletariat.Configuration{
		Address: tcpTwoAddr,
		Timeout: 0,
		Ctx:     ctxTwo,
	})
	if err != nil {
		t.Fatalf("failed tcp two: %v", err)
	}

	go commOne.Start()
	go commTwo.Start()

	content := []byte("Ola, Mundo!")

	groupSync := sync.WaitGroup{}
	groupSync.Add(1)
	go func() {
		defer groupSync.Done()
		msg := <-commOne.Receive()
		if msg.Err != nil {
			t.Errorf("should not receive error. %#v", msg.Err)
		}
		if string(content) != msg.Data.String() {
			t.Errorf("failed. should be %s found %s", string(content), msg.Data.String())
		}
	}()

	err = commTwo.Send(tcpOneAddr, content)
	if err != nil {
		t.Fatalf("failed sending. %v", err)
	}

	time.Sleep(testTimeout)
	cancelOne()
	cancelTwo()

	if !WaitThisOrTimeout(func() {
		err = commOne.Close()
		if err != nil {
			t.Errorf("failed closing one. %v", err)
		}
	}, testTimeout) {
		PrintStackTrace(t)
		t.Errorf("took to long closing one")
	}

	if !WaitThisOrTimeout(func() {
		err = commTwo.Close()
		if err != nil {
			t.Errorf("failed closing two. %v", err)
		}
	}, testTimeout) {
		PrintStackTrace(t)
		t.Errorf("took to long closing two")
	}

	if !WaitThisOrTimeout(groupSync.Wait, time.Second) {
		t.Errorf("took to long closing goroutine")
	}
}
