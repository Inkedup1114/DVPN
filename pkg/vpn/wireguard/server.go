package wireguard

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/Inkedup1114/dvpn/pkg/vpn"
)

type WireGuardServer struct {
	config    *vpn.Config
	clients   map[string]*ClientInfo
	mutex     sync.RWMutex
	running   bool
	stats     *vpn.Stats
	statsChan chan *vpn.Stats
	wgClient  WgClient
	runner    ExecRunner
}

type ClientInfo struct {
	PublicKey wgtypes.Key
	IPAddress string
	AddedAt   time.Time
	LastSeen  time.Time
}

func NewWireGuardServer(config *vpn.Config) (*WireGuardServer, error) {
	return NewWireGuardServerWithRunner(config, &DefaultExecRunner{})
}

func NewWireGuardServerWithRunner(config *vpn.Config, runner ExecRunner) (*WireGuardServer, error) {
	defaultWg, err := NewDefaultWgClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create WireGuard client: %v", err)
	}

	return NewWireGuardServerWithDeps(config, runner, defaultWg)
}

func NewWireGuardServerWithDeps(config *vpn.Config, runner ExecRunner, wgclient WgClient) (*WireGuardServer, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if config.Interface == "" {
		return nil, fmt.Errorf("interface name is required")
	}
	// PrivateKey may be empty for tests or when interface is created later.

	return &WireGuardServer{
		config:    config,
		clients:   make(map[string]*ClientInfo),
		stats:     &vpn.Stats{},
		statsChan: make(chan *vpn.Stats, 100),
		wgClient:  wgclient,
		runner:    runner,
	}, nil
}

func (s *WireGuardServer) Start(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.running {
		return fmt.Errorf("server already running")
	}

	// Parse private key
	privateKey, err := wgtypes.ParseKey(s.config.PrivateKey)
	if err != nil {
		return fmt.Errorf("invalid private key: %v", err)
	}

	// Create WireGuard interface
	err = s.createInterface(privateKey)
	if err != nil {
		return fmt.Errorf("failed to create interface: %v", err)
	}

	s.running = true

	// Start stats collection
	go s.statsLoop(ctx)

	return nil
}

func (s *WireGuardServer) Stop() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.running {
		return nil
	}

	// Remove WireGuard interface
	err := s.destroyInterface()
	if err != nil {
		return fmt.Errorf("failed to destroy interface: %v", err)
	}

	s.running = false
	s.wgClient.Close()
	return nil
}

func (s *WireGuardServer) AddClient(clientKeyStr, clientIP string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	clientKey, err := wgtypes.ParseKey(clientKeyStr)
	if err != nil {
		return fmt.Errorf("invalid client key: %v", err)
	}

	client := &ClientInfo{
		PublicKey: clientKey,
		IPAddress: clientIP,
		AddedAt:   time.Now(),
		LastSeen:  time.Now(),
	}

	s.clients[clientKeyStr] = client

	// Add peer to WireGuard using wgctrl
	return s.addPeer(clientKey, clientIP)
}

func (s *WireGuardServer) RemoveClient(clientKeyStr string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	clientKey, err := wgtypes.ParseKey(clientKeyStr)
	if err != nil {
		return fmt.Errorf("invalid client key: %v", err)
	}

	delete(s.clients, clientKeyStr)

	// Remove peer from WireGuard
	return s.removePeer(clientKey)
}

func (s *WireGuardServer) GetStats() *vpn.Stats {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return &vpn.Stats{
		BytesSent:     s.stats.BytesSent,
		BytesReceived: s.stats.BytesReceived,
		PacketsSent:   s.stats.PacketsSent,
		PacketsRecv:   s.stats.PacketsRecv,
		Uptime:        s.stats.Uptime,
	}
}

func (s *WireGuardServer) createInterface(privateKey wgtypes.Key) error {
	// Create WireGuard interface using ip command
	if err := s.runner.Run("sudo", "ip", "link", "add", "dev", s.config.Interface, "type", "wireguard"); err != nil {
		return fmt.Errorf("failed to create interface: %v", err)
	}

	// Configure the interface using wgctrl
	listenPort := s.config.ListenPort
	cfg := wgtypes.Config{
		PrivateKey: &privateKey,
		ListenPort: &listenPort,
	}

	if err := s.wgClient.ConfigureDevice(s.config.Interface, cfg); err != nil {
		return fmt.Errorf("failed to configure WireGuard device: %v", err)
	}

	// Set IP address on interface
	if err := s.runner.Run("sudo", "ip", "addr", "add", "10.0.0.1/24", "dev", s.config.Interface); err != nil {
		return fmt.Errorf("failed to set IP address: %v", err)
	}

	// Bring interface up
	if err := s.runner.Run("sudo", "ip", "link", "set", "up", "dev", s.config.Interface); err != nil {
		return fmt.Errorf("failed to bring interface up: %v", err)
	}

	return nil
}

func (s *WireGuardServer) destroyInterface() error {
	return s.runner.Run("sudo", "ip", "link", "del", "dev", s.config.Interface)
}

func (s *WireGuardServer) addPeer(publicKey wgtypes.Key, allowedIP string) error {
	// Parse allowed IP
	_, ipNet, err := net.ParseCIDR(allowedIP + "/32")
	if err != nil {
		return fmt.Errorf("invalid IP address: %v", err)
	}

	// Configure peer
	peer := wgtypes.PeerConfig{
		PublicKey:  publicKey,
		AllowedIPs: []net.IPNet{*ipNet},
	}

	cfg := wgtypes.Config{
		Peers: []wgtypes.PeerConfig{peer},
	}

	return s.wgClient.ConfigureDevice(s.config.Interface, cfg)
}

func (s *WireGuardServer) removePeer(publicKey wgtypes.Key) error {
	peer := wgtypes.PeerConfig{
		PublicKey: publicKey,
		Remove:    true,
	}

	cfg := wgtypes.Config{
		Peers: []wgtypes.PeerConfig{peer},
	}

	return s.wgClient.ConfigureDevice(s.config.Interface, cfg)
}

func (s *WireGuardServer) statsLoop(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.updateStats()
		}
	}
}

func (s *WireGuardServer) updateStats() {
	// Get stats from WireGuard using wgctrl
	device, err := s.wgClient.Device(s.config.Interface)
	if err != nil {
		return
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Aggregate stats from all peers
	var totalSent, totalReceived uint64
	for _, peer := range device.Peers {
		totalSent += uint64(peer.TransmitBytes)
		totalReceived += uint64(peer.ReceiveBytes)
	}

	s.stats.BytesSent = totalSent
	s.stats.BytesReceived = totalReceived
	s.stats.Uptime = time.Now().Unix()
}
