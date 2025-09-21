package wireguard

import (
	"testing"

	"github.com/Inkedup1114/dvpn/pkg/vpn"
)

func TestServerConstructorValidatesConfig(t *testing.T) {
	_, err := NewWireGuardServerWithDeps(nil, &fakeRunner{}, &fakeWgClient{})
	if err == nil {
		t.Fatalf("expected error for nil config, got nil")
	}

	_, err = NewWireGuardServerWithDeps(&vpn.Config{}, &fakeRunner{}, &fakeWgClient{})
	if err == nil {
		t.Fatalf("expected error for empty config fields, got nil")
	}
}

func TestClientConstructorValidatesConfig(t *testing.T) {
	_, err := NewWireGuardClientWithRunner(nil, &fakeRunner{})
	if err == nil {
		t.Fatalf("expected error for nil config, got nil")
	}

	_, err = NewWireGuardClientWithRunner(&vpn.Config{}, &fakeRunner{})
	if err == nil {
		t.Fatalf("expected error for empty interface, got nil")
	}
}
