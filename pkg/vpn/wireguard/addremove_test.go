package wireguard

import (
	"encoding/base64"
	"net"
	"testing"

	"github.com/Inkedup1114/dvpn/pkg/vpn"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// recordingFakeWgClient provided by testhelpers_test.go

func TestAddRemoveClientConfiguresPeers(t *testing.T) {
	cfg := &vpn.Config{Interface: "dvpn-test"}
	fr := &fakeRunner{}
	rf := &recordingFakeWgClient{}

	s, err := NewWireGuardServerWithDeps(cfg, fr, rf)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// create a valid wireguard key string (32 zero bytes base64)
	keyBytes := make([]byte, 32)
	keyStr := base64.StdEncoding.EncodeToString(keyBytes)

	// Add client
	if err := s.AddClient(keyStr, "10.0.0.5"); err != nil {
		t.Fatalf("AddClient failed: %v", err)
	}

	if rf.calls == 0 {
		t.Fatalf("expected ConfigureDevice to be called on AddClient")
	}
	if len(rf.lastCfg.Peers) != 1 {
		t.Fatalf("expected 1 peer in config, got %d", len(rf.lastCfg.Peers))
	}

	peer := rf.lastCfg.Peers[0]
	parsedKey, err := wgtypes.ParseKey(keyStr)
	if err != nil {
		t.Fatalf("failed to parse key in test: %v", err)
	}
	if peer.PublicKey != parsedKey {
		t.Fatalf("peer public key mismatch: got %v want %v", peer.PublicKey, parsedKey)
	}
	if len(peer.AllowedIPs) != 1 || peer.AllowedIPs[0].String() != "10.0.0.5/32" {
		t.Fatalf("unexpected AllowedIPs: %v", peer.AllowedIPs)
	}

	// Remove client
	if err := s.RemoveClient(keyStr); err != nil {
		t.Fatalf("RemoveClient failed: %v", err)
	}

	if rf.calls < 2 {
		t.Fatalf("expected ConfigureDevice to be called on RemoveClient as well, calls=%d", rf.calls)
	}
	if len(rf.lastCfg.Peers) != 1 {
		t.Fatalf("expected 1 peer in config after remove, got %d", len(rf.lastCfg.Peers))
	}
	peer = rf.lastCfg.Peers[0]
	if !peer.Remove {
		t.Fatalf("expected peer Remove=true on removal, got %#v", peer)
	}
	if peer.PublicKey != parsedKey {
		t.Fatalf("peer public key mismatch on removal: got %v want %v", peer.PublicKey, parsedKey)
	}

	// Also validate the parse step used by addPeer
	if _, _, err := net.ParseCIDR("10.0.0.5/32"); err != nil {
		t.Fatalf("test internal: failed to parse CIDR: %v", err)
	}
}

func TestAddClientInvalidKey(t *testing.T) {
	cfg := &vpn.Config{Interface: "dvpn-test"}
	fr := &fakeRunner{}
	rf := &recordingFakeWgClient{}

	s, err := NewWireGuardServerWithDeps(cfg, fr, rf)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// attempt to add invalid key
	if err := s.AddClient("not-a-valid-key", "10.0.0.6"); err == nil {
		t.Fatalf("expected error for invalid key, got nil")
	}
}

func TestAddRemoveClientConcurrency(t *testing.T) {
	cfg := &vpn.Config{Interface: "dvpn-test"}
	fr := &fakeRunner{}
	rf := &recordingFakeWgClient{}

	s, err := NewWireGuardServerWithDeps(cfg, fr, rf)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	keyBytes := make([]byte, 32)
	keyStr := base64.StdEncoding.EncodeToString(keyBytes)

	// run multiple goroutines adding/removing the same client
	done := make(chan struct{})
	for i := 0; i < 20; i++ {
		go func() {
			for j := 0; j < 50; j++ {
				_ = s.AddClient(keyStr, "10.0.0.7")
				_ = s.RemoveClient(keyStr)
			}
			done <- struct{}{}
		}()
	}

	// wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// ensure no panics and that ConfigureDevice was called at least once
	if rf.calls == 0 {
		t.Fatalf("expected ConfigureDevice to be called during concurrent ops")
	}
}

func TestMultiKeyConcurrency(t *testing.T) {
	cfg := &vpn.Config{Interface: "dvpn-test"}
	fr := &fakeRunner{}
	rf := &recordingFakeWgClient{}

	s, err := NewWireGuardServerWithDeps(cfg, fr, rf)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// generate multiple distinct keys
	keys := make([]string, 50)
	for i := 0; i < 50; i++ {
		k := make([]byte, 32)
		k[0] = byte(i)
		keys[i] = base64.StdEncoding.EncodeToString(k)
	}

	// concurrently add/remove many keys
	done := make(chan struct{})
	for i := 0; i < 50; i++ {
		key := keys[i]
		go func() {
			for j := 0; j < 20; j++ {
				_ = s.AddClient(key, "10.0.1.1")
				_ = s.RemoveClient(key)
			}
			done <- struct{}{}
		}()
	}

	for i := 0; i < 50; i++ {
		<-done
	}

	// final server clients map should be empty or at least stable; check length
	s.mutex.RLock()
	remaining := len(s.clients)
	s.mutex.RUnlock()

	if remaining != 0 {
		t.Fatalf("expected 0 clients after concurrent ops, got %d", remaining)
	}
}
