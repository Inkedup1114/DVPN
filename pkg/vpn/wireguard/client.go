package wireguard

import (
	"fmt"
	"os"
	"sync"

	"golang.zx2c4.com/wireguard/wgctrl"

	"github.com/Inkedup1114/dvpn/pkg/vpn"
)

type WireGuardClient struct {
	config         *vpn.Config
	serverKey      string
	serverEndpoint string
	connected      bool
	mutex          sync.RWMutex
	stats          *vpn.Stats
	wgClient       *wgctrl.Client
	runner         ExecRunner
}

func NewWireGuardClient(config *vpn.Config) (*WireGuardClient, error) {
	return NewWireGuardClientWithRunner(config, &DefaultExecRunner{})
}

func NewWireGuardClientWithRunner(config *vpn.Config, runner ExecRunner) (*WireGuardClient, error) {
	wgClient, err := wgctrl.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create WireGuard client: %v", err)
	}

	return &WireGuardClient{
		config:   config,
		stats:    &vpn.Stats{},
		wgClient: wgClient,
		runner:   runner,
	}, nil
}

func (c *WireGuardClient) Connect(serverKey, serverEndpoint string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.connected {
		return fmt.Errorf("already connected")
	}

	c.serverKey = serverKey
	c.serverEndpoint = serverEndpoint

	// Create client configuration
	err := c.createClientConfig()
	if err != nil {
		return fmt.Errorf("failed to create config: %v", err)
	}

	// Start WireGuard connection
	err = c.startConnection()
	if err != nil {
		return fmt.Errorf("failed to start connection: %v", err)
	}

	c.connected = true
	return nil
}

func (c *WireGuardClient) Disconnect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.connected {
		return nil
	}

	err := c.stopConnection()
	if err != nil {
		return fmt.Errorf("failed to stop connection: %v", err)
	}

	c.connected = false
	return nil
}

func (c *WireGuardClient) IsConnected() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.connected
}

func (c *WireGuardClient) GetStats() *vpn.Stats {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return &vpn.Stats{
		BytesSent:     c.stats.BytesSent,
		BytesReceived: c.stats.BytesReceived,
		PacketsSent:   c.stats.PacketsSent,
		PacketsRecv:   c.stats.PacketsRecv,
		Uptime:        c.stats.Uptime,
	}
}

func (c *WireGuardClient) createClientConfig() error {
	config := fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = 10.0.0.2/24

[Peer]
PublicKey = %s
Endpoint = %s
AllowedIPs = 0.0.0.0/0
`,
		c.config.PrivateKey,
		c.serverKey,
		c.serverEndpoint,
	)

	// Write config to temporary file
	configPath := fmt.Sprintf("/tmp/dvpn-client-%s.conf", c.config.Interface)
	return os.WriteFile(configPath, []byte(config), 0600)
}

func (c *WireGuardClient) startConnection() error {
	configPath := fmt.Sprintf("/tmp/dvpn-client-%s.conf", c.config.Interface)
	return c.runner.Run("sudo", "wg-quick", "up", configPath)
}

func (c *WireGuardClient) stopConnection() error {
	configPath := fmt.Sprintf("/tmp/dvpn-client-%s.conf", c.config.Interface)
	return c.runner.Run("sudo", "wg-quick", "down", configPath)
}
