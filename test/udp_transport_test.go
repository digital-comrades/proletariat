package test

import (
	"github.com/jabolina/proletariat/internal"
	"github.com/jabolina/proletariat/pkg/proletariat"
	"net"
	"testing"
)

func TestUDPTransport_GoodAddr(t *testing.T) {
	addr := net.UDPAddr{IP: net.ParseIP("127.0.0.1")}
	_, err := internal.NewUDPTransport(&addr)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
}

func TestUDPTransport_BadAddr(t *testing.T) {
	addr := net.UDPAddr{IP: net.ParseIP("0.0.0.0")}
	_, err := internal.NewUDPTransport(&addr)
	if err != proletariat.ErrInvalidAddr {
		t.Fatalf("failed: %v", err)
	}
}
