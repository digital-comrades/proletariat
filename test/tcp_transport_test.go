package test

import (
	"context"
	"github.com/digital-comrades/proletariat/pkg/proletariat"
	"testing"
)

func TestTCPTransport_GoodAddress(t *testing.T) {
	_, err := proletariat.NewTCPTransport(context.TODO(), "localhost:0")
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
}

func TestTCPTransport_BadAddress(t *testing.T) {
	_, err := proletariat.NewTCPTransport(context.TODO(), "0.0.0.0:0")
	if err != proletariat.ErrInvalidAddr {
		t.Fatalf("failed: %v", err)
	}
}

func TestTCPTransport_EmptyAddr(t *testing.T) {
	_, err := proletariat.NewTCPTransport(context.TODO(), ":0")
	if err != proletariat.ErrInvalidAddr {
		t.Fatalf("failed: %v", err)
	}
}
