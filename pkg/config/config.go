package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type Config struct {
	Network NetworkConfig `mapstructure:"network"`
	Mining  MiningConfig  `mapstructure:"mining"`
	VPN     VPNConfig     `mapstructure:"vpn"`
	API     APIConfig     `mapstructure:"api"`
	Logging LoggingConfig `mapstructure:"logging"`
}

type NetworkConfig struct {
	ListenAddress string `mapstructure:"listen_address"`
	MaxPeers      int    `mapstructure:"max_peers"`
}

type MiningConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	GPUBackend string `mapstructure:"gpu_backend"`
	Address    string `mapstructure:"miner_address"`
}

type VPNConfig struct {
	Enabled        bool   `mapstructure:"enabled"`
	Interface      string `mapstructure:"interface"`
	ListenPort     int    `mapstructure:"listen_port"`
	PrivateKey     string `mapstructure:"private_key"`
	PrivateKeyFile string `mapstructure:"private_key_file"`
}

type APIConfig struct {
	Enabled       bool   `mapstructure:"enabled"`
	ListenAddress string `mapstructure:"listen_address"`
	CorsEnabled   bool   `mapstructure:"cors_enabled"`
}

type LoggingConfig struct {
	Level string `mapstructure:"level"`
	File  string `mapstructure:"file"`
}

func LoadConfig(filename string) (*Config, error) {
	viper.SetConfigFile(filename)
	viper.SetEnvPrefix("DVPN")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %v", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	return &config, nil
}

func (c *Config) GetVPNPrivateKey() (string, error) {
	if c.VPN.PrivateKey != "" {
		// Clean the key - remove quotes and whitespace
		key := strings.TrimSpace(c.VPN.PrivateKey)
		key = strings.Trim(key, "\"'")
		return key, nil
	}

	if c.VPN.PrivateKeyFile != "" {
		data, err := os.ReadFile(c.VPN.PrivateKeyFile)
		if err != nil {
			return "", fmt.Errorf("failed to read private key file: %v", err)
		}
		return strings.TrimSpace(string(data)), nil
	}

	return "", fmt.Errorf("no VPN private key configured")
}

// ValidateVPN validates VPN-related configuration fields. It checks interface name,
// listen port range, and private key presence/format when VPN is enabled.
func (c *Config) ValidateVPN() error {
	if !c.VPN.Enabled {
		return nil
	}

	if c.VPN.Interface == "" {
		return fmt.Errorf("vpn.interface is required when VPN is enabled")
	}

	if c.VPN.ListenPort <= 0 || c.VPN.ListenPort > 65535 {
		return fmt.Errorf("vpn.listen_port must be between 1 and 65535")
	}

	// Ensure a private key is available (either inline or from file)
	key, err := c.GetVPNPrivateKey()
	if err != nil {
		return fmt.Errorf("vpn private key: %v", err)
	}

	// Validate wireguard private key format
	if _, err := wgtypes.ParseKey(key); err != nil {
		return fmt.Errorf("invalid vpn private key: %v", err)
	}

	return nil
}
