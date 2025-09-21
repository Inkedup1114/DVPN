package wireguard

import (
	"errors"
	"testing"

	"github.com/Inkedup1114/dvpn/pkg/vpn"
)

// failingRunner simulates command execution failures
type failingRunner struct{}

func (r *failingRunner) Run(name string, args ...string) error {
	return errors.New("exec failed")
}

func TestServerStopBubblesRunnerError(t *testing.T) {
	cfg := &vpn.Config{Interface: "dvpn-test"}
	fr := &failingRunner{}
	rf := &recordingFakeWgClient{}

	s, err := NewWireGuardServerWithDeps(cfg, fr, rf)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// mark server as running so Stop will attempt to destroy the interface
	s.running = true

	// call destroyInterface via Stop; Stop should return exec error
	if err := s.Stop(); err == nil {
		t.Fatalf("expected error from Stop when runner fails, got nil")
	}
}

func TestClientDisconnectBubblesRunnerError(t *testing.T) {
	cfg := &vpn.Config{Interface: "dvpn-test"}
	fr := &failingRunner{}

	c, err := NewWireGuardClientWithRunner(cfg, fr)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// mark connected to allow Disconnect to attempt stopConnection
	c.connected = true

	if err := c.Disconnect(); err == nil {
		t.Fatalf("expected error from Disconnect when runner fails, got nil")
	}
}
