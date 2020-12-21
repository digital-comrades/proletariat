package test

import (
	"github.com/jabolina/proletariat/internal"
	"github.com/jabolina/proletariat/pkg/proletariat"
	"testing"
)

func TestTCPTransport_GoodAddress(t *testing.T) {
	_, err := internal.NewTCPTransport("localhost:0")
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
}

func TestTCPTransport_BadAddress(t *testing.T) {
	_, err := internal.NewTCPTransport("0.0.0.0:0")
	if err != proletariat.ErrInvalidAddr {
		t.Fatalf("failed: %v", err)
	}
}

func TestTCPTransport_EmptyAddr(t *testing.T) {
	_, err := internal.NewTCPTransport(":0")
	if err != proletariat.ErrInvalidAddr {
		t.Fatalf("failed: %v", err)
	}
}
