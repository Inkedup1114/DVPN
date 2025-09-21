package wireguard

import (
	"context"
	"testing"
	"time"

	"github.com/Inkedup1114/dvpn/pkg/vpn"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// fakeWgClient provided by testhelpers_test.go

func TestServerCreateDestroyUsesRunnerAndWgClient(t *testing.T) {
	cfg := &vpn.Config{Interface: "dvpn-test", ListenPort: 51820, PrivateKey: "DEMO"}
	fr := &fakeRunner{}
	fw := &fakeWgClient{}

	s, err := NewWireGuardServerWithDeps(cfg, fr, fw)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// Start should create interface and call ConfigureDevice
	if err := s.Start(context.TODO()); err == nil {
		t.Fatalf("expected error due to invalid private key, got nil")
	}

	// Use a valid private key string for ParseKey by creating a random key-like string
	// For test simplicity, bypass Start and call createInterface directly with a zeroed key
	var key wgtypes.Key
	if err := s.createInterface(key); err != nil {
		t.Fatalf("createInterface failed: %v", err)
	}

	if !fw.configured {
		t.Fatalf("expected wg client to be configured")
	}

	// destroyInterface should call runner
	if err := s.destroyInterface(); err != nil {
		t.Fatalf("destroyInterface failed: %v", err)
	}

	// ensure runner saw a delete call
	foundDel := false
	for _, call := range fr.calls {
		if len(call) >= 4 && call[0] == "sudo" && call[1] == "ip" && call[2] == "link" && call[3] == "del" {
			foundDel = true
		}
	}
	if !foundDel {
		t.Fatalf("expected ip link del call, got: %v", fr.calls)
	}

	// small sleep to allow any goroutines to settle (defensive)
	time.Sleep(10 * time.Millisecond)
}
