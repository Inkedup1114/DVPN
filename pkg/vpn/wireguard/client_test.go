package wireguard

import (
	"testing"

	"github.com/Inkedup1114/dvpn/pkg/vpn"
)

// fakeRunner provided by testhelpers_test.go

func TestClientStartStopUsesRunner(t *testing.T) {
	cfg := &vpn.Config{Interface: "dvpn-test"}
	fr := &fakeRunner{}

	c, err := NewWireGuardClientWithRunner(cfg, fr)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// call Connect (startConnection will be invoked)
	if err := c.Connect("serverkey", "1.2.3.4:51820"); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Expect wg-quick up call
	foundUp := false
	for _, call := range fr.calls {
		if len(call) >= 3 && call[0] == "sudo" && call[1] == "wg-quick" && call[2] == "up" {
			foundUp = true
		}
	}
	if !foundUp {
		t.Fatalf("expected wg-quick up call, got: %v", fr.calls)
	}

	// Disconnect should call wg-quick down
	if err := c.Disconnect(); err != nil {
		t.Fatalf("Disconnect failed: %v", err)
	}

	foundDown := false
	for _, call := range fr.calls {
		if len(call) >= 3 && call[0] == "sudo" && call[1] == "wg-quick" && call[2] == "down" {
			foundDown = true
		}
	}
	if !foundDown {
		t.Fatalf("expected wg-quick down call, got: %v", fr.calls)
	}
}
