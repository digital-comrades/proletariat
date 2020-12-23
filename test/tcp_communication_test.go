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
	tcpOne, err := internal.NewTCPTransport(tcpOneAddr)
	if err != nil {
		t.Fatalf("failed tcp one: %v", err)
	}

	ctxTwo, cancelTwo := context.WithCancel(context.TODO())
	tcpTwo, err := internal.NewTCPTransport(tcpTwoAddr)
	if err != nil {
		t.Fatalf("failed tcp two: %v", err)
	}

	commOne := internal.NewCommunication(proletariat.CommunicationConfiguration{
		Address:   tcpOneAddr,
		Timeout:   time.Second,
		Transport: tcpOne,
		Ctx:       ctxOne,
	})
	commTwo := internal.NewCommunication(proletariat.CommunicationConfiguration{
		Address:   tcpTwoAddr,
		Timeout:   time.Second,
		Transport: tcpTwo,
		Ctx:       ctxTwo,
	})
	commOne.Start()
	commTwo.Start()

	content := []byte("Ola, Mundo!")

	groupSync := sync.WaitGroup{}
	groupSync.Add(1)
	go func() {
		defer groupSync.Done()
		msg := <-commOne.Receive()
		if string(content) != string(msg) {
			t.Fatalf("failed. should be %s found %s", string(content), string(msg))
		}
	}()

	err = commTwo.Send(tcpOneAddr, content)
	if err != nil {
		t.Fatalf("failed sending. %v", err)
	}

	time.Sleep(time.Second)
	cancelOne()
	cancelTwo()

	if !WaitThisOrTimeout(func() {
		err = commOne.Close()
		if err != nil {
			t.Fatalf("failed closing one. %v", err)
		}
	}, time.Second) {
		t.Fatalf("took to long closing one")
	}

	if !WaitThisOrTimeout(func() {
		err = commTwo.Close()
		if err != nil {
			t.Fatalf("failed closing two. %v", err)
		}
	}, time.Second) {
		t.Fatalf("took to long closing two")
	}

	if !WaitThisOrTimeout(groupSync.Wait, time.Second) {
		t.Fatalf("took to long closing goroutine")
	}
}
