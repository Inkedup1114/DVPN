package wireguard

import (
	"context"
	"errors"
	"testing"

	"github.com/Inkedup1114/dvpn/pkg/vpn"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// failingWgClient always returns an error from ConfigureDevice and Device
type failingWgClient struct{}

func (f *failingWgClient) ConfigureDevice(iface string, cfg wgtypes.Config) error {
	return errors.New("configure failed")
}

func (f *failingWgClient) Device(iface string) (*wgtypes.Device, error) {
	return nil, errors.New("device failed")
}

func (f *failingWgClient) Close() error { return nil }

func TestAddClientBubblesWgError(t *testing.T) {
	cfg := &vpn.Config{Interface: "dvpn-test"}
	fr := &fakeRunner{}
	fw := &failingWgClient{}

	s, err := NewWireGuardServerWithDeps(cfg, fr, fw)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// key: 32 zero bytes base64
	keyStr := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="

	if err := s.AddClient(keyStr, "10.0.0.5"); err == nil {
		t.Fatalf("expected error when wgclient.ConfigureDevice fails, got nil")
	}
}

func TestStartBubblesWgError(t *testing.T) {
	// Use a valid-looking private key: 32 zero bytes base64
	keyStr := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="

	cfg := &vpn.Config{Interface: "dvpn-test", PrivateKey: keyStr, ListenPort: 51820}
	fr := &fakeRunner{}
	fw := &failingWgClient{}

	s, err := NewWireGuardServerWithDeps(cfg, fr, fw)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	if err := s.Start(context.TODO()); err == nil {
		t.Fatalf("expected error from Start when wgclient.ConfigureDevice fails, got nil")
	}
}
