package test

import (
	"context"
	"github.com/digital-comrades/proletariat/pkg/proletariat"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func sendMultipleMessages(first, second proletariat.Communication, testSize int, b *testing.B) {
	content := []byte("Ola, Mundo!")
	addr := proletariat.Address(first.Addr().String())
	wg := &sync.WaitGroup{}
	counter := int64(0)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-time.After(time.Second):
				return
			case <-first.Receive():
				atomic.AddInt64(&counter, 1)
			}
		}
	}()

	for i := 0; i < testSize; i++ {
		err := second.Send(addr, content)
		if err != nil {
			b.Errorf("failed writing. %s", err.Error())
		}
	}

	wg.Wait()
	if atomic.LoadInt64(&counter) != int64(testSize) {
		b.Errorf("expected %d. found %d", testSize, atomic.LoadInt64(&counter))
	}
}

func Benchmark_CommunicationMessages(b *testing.B) {
	testSize := 1024
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

	for i := 0; i < b.N; i++ {
		sendMultipleMessages(first, second, testSize, b)
	}

	cancel()

	if err := first.Close(); err != nil {
		b.Errorf("failed closing first. %s", err.Error())
	}

	if err := second.Close(); err != nil {
		b.Errorf("failed closing second. %s", err.Error())
	}
}

func Benchmark_CommunicationParallelMessages(b *testing.B) {
	testSize := 1024
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

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sendMultipleMessages(first, second, testSize, b)
		}
	})

	cancel()

	if err := first.Close(); err != nil {
		b.Errorf("failed closing first. %s", err.Error())
	}

	if err := second.Close(); err != nil {
		b.Errorf("failed closing second. %s", err.Error())
	}
}
