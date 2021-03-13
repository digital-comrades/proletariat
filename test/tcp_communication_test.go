package test

import (
	"context"
	"github.com/jabolina/proletariat/internal"
	"github.com/jabolina/proletariat/pkg/proletariat"
	"sync"
	"testing"
	"time"
)

func TestTCPCommunication_CreateAndSend(t *testing.T) {
	tcpOneAddr := proletariat.Address("127.0.0.1:11111")
	tcpTwoAddr := proletariat.Address("127.0.0.1:22222")
	ctxOne, cancelOne := context.WithCancel(context.TODO())
	ctxTwo, cancelTwo := context.WithCancel(context.TODO())
	testTimeout := 500 * time.Millisecond

	commOne, err := internal.NewCommunication(proletariat.CommunicationConfiguration{
		Address: tcpOneAddr,
		Timeout: 0,
		Ctx:     ctxOne,
	})
	if err != nil {
		t.Fatalf("failed tcp one: %v", err)
	}

	commTwo, err := internal.NewCommunication(proletariat.CommunicationConfiguration{
		Address: tcpTwoAddr,
		Timeout: 0,
		Ctx:     ctxTwo,
	})
	if err != nil {
		t.Fatalf("failed tcp two: %v", err)
	}

	commOne.Start()
	commTwo.Start()

	content := []byte("Ola, Mundo!")

	groupSync := sync.WaitGroup{}
	groupSync.Add(1)
	go func() {
		defer groupSync.Done()
		msg := <-commOne.Receive()
		if msg.Err != nil {
			t.Errorf("should not receive error. %#v", msg.Err)
		}
		if string(content) != string(msg.Data) {
			t.Fatalf("failed. should be %s found %s", string(content), string(msg.Data))
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
			t.Fatalf("failed closing one. %v", err)
		}
	}, testTimeout) {
		PrintStackTrace(t)
		t.Fatalf("took to long closing one")
	}

	if !WaitThisOrTimeout(func() {
		err = commTwo.Close()
		if err != nil {
			t.Fatalf("failed closing two. %v", err)
		}
	}, testTimeout) {
		PrintStackTrace(t)
		t.Fatalf("took to long closing two")
	}

	if !WaitThisOrTimeout(groupSync.Wait, time.Second) {
		t.Fatalf("took to long closing goroutine")
	}
}
