package config

import (
	"os"
	"testing"
)

func TestValidateVPNDisabled(t *testing.T) {
	c := &Config{}
	// VPN disabled by default (zero value) -> no error
	if err := c.ValidateVPN(); err != nil {
		t.Fatalf("unexpected error for disabled VPN: %v", err)
	}
}

func TestValidateVPNInvalidPort(t *testing.T) {
	c := &Config{}
	c.VPN.Enabled = true
	c.VPN.Interface = "dvpn0"
	c.VPN.ListenPort = 70000

	if err := c.ValidateVPN(); err == nil {
		t.Fatalf("expected error for invalid port, got nil")
	}
}

func TestValidateVPNMissingKey(t *testing.T) {
	c := &Config{}
	c.VPN.Enabled = true
	c.VPN.Interface = "dvpn0"
	c.VPN.ListenPort = 51820

	if err := c.ValidateVPN(); err == nil {
		t.Fatalf("expected error for missing key, got nil")
	}
}

func TestValidateVPNValidInlineKey(t *testing.T) {
	c := &Config{}
	c.VPN.Enabled = true
	c.VPN.Interface = "dvpn0"
	c.VPN.ListenPort = 51820
	// 32 zero bytes base64
	c.VPN.PrivateKey = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="

	if err := c.ValidateVPN(); err != nil {
		t.Fatalf("unexpected error for valid inline key: %v", err)
	}
}

func TestValidateVPNValidKeyFromFile(t *testing.T) {
	c := &Config{}
	c.VPN.Enabled = true
	c.VPN.Interface = "dvpn0"
	c.VPN.ListenPort = 51820

	// create temp file containing key
	f, err := os.CreateTemp("", "key-*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())

	if _, err := f.WriteString("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="); err != nil {
		t.Fatalf("failed to write temp key: %v", err)
	}
	f.Close()

	c.VPN.PrivateKeyFile = f.Name()

	if err := c.ValidateVPN(); err != nil {
		t.Fatalf("unexpected error for valid key file: %v", err)
	}
}
