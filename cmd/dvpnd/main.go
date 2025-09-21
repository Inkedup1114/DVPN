package main

import (
	"context"
	"crypto/ed25519"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Inkedup1114/dvpn/internal/api"
	"github.com/Inkedup1114/dvpn/pkg/blockchain"
	"github.com/Inkedup1114/dvpn/pkg/config"
	"github.com/Inkedup1114/dvpn/pkg/mining"
	"github.com/Inkedup1114/dvpn/pkg/mining/cpu"
	"github.com/Inkedup1114/dvpn/pkg/p2p"
	"github.com/Inkedup1114/dvpn/pkg/vpn"
	"github.com/Inkedup1114/dvpn/pkg/vpn/wireguard"
)

func main() {
	var configFile = flag.String("config", "configs/node.yaml", "Configuration file path")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Printf("Failed to load config: %v, using defaults", err)
		cfg = getDefaultConfig()
	}

	// Initialize blockchain
	bc := blockchain.NewBlockchain()
	latestBlock := bc.GetLatestBlock()
	log.Printf("Blockchain initialized with genesis block from miner: %s", latestBlock.MinerAddress)

	// Generate a temporary private key for demo
	_, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		log.Fatalf("Failed to generate private key: %v", err)
	}

	// Initialize P2P network
	p2pNet, err := p2p.NewP2PNetwork(cfg.Network.ListenAddress, privateKey)
	if err != nil {
		log.Fatalf("Failed to create P2P network: %v", err)
	}

	// Initialize VPN server if enabled
	var vpnServer vpn.VPNServer
	if cfg.VPN.Enabled {
		privateKey, err := cfg.GetVPNPrivateKey()
		if err != nil {
			log.Printf("VPN private key error: %v", err)
		} else {
			vpnConfig := &vpn.Config{
				Interface:  cfg.VPN.Interface,
				Port:       cfg.VPN.ListenPort,
				PrivateKey: privateKey,
				ListenPort: cfg.VPN.ListenPort,
			}
			var err error
			vpnServer, err = wireguard.NewWireGuardServer(vpnConfig)
			if err != nil {
				log.Printf("Failed to initialize WireGuard server: %v", err)
				vpnServer = nil
			}
		}
	}

	// Initialize miner if enabled - FIXED: Use proper mining.Miner type
	var miner mining.Miner
	if cfg.Mining.Enabled {
		switch cfg.Mining.GPUBackend {
		case "cpu":
			miner = cpu.NewCPUMiner()
			log.Println("CPU miner initialized")
		case "cuda":
			log.Println("CUDA mining not available in this build")
		case "opencl":
			log.Println("OpenCL mining not yet implemented")
		default:
			miner = cpu.NewCPUMiner()
			log.Println("CPU miner initialized (default)")
		}
	}

	// Initialize API server if enabled - FIXED: Pass actual miner
	var apiServer *api.Server
	if cfg.API.Enabled {
		apiServer = api.NewServer(cfg.API.ListenAddress, miner, p2pNet)
		go func() {
			log.Printf("API server starting on %s", cfg.API.ListenAddress)
			if err := apiServer.Start(); err != nil && err != http.ErrServerClosed {
				log.Printf("API server error: %v", err)
			}
		}()
	}

	// Start services
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := p2pNet.Start(); err != nil {
		log.Fatalf("Failed to start P2P network: %v", err)
	}
	log.Println("P2P network started")

	if vpnServer != nil {
		if err := vpnServer.Start(ctx); err != nil {
			log.Printf("Failed to start VPN server: %v", err)
		} else {
			log.Println("VPN server started")
		}
	}

	if miner != nil {
		log.Println("Miner configured (start/stop via API)")
	}

	fmt.Println("DVPN node started successfully!")
	fmt.Printf("Blockchain: genesis block by '%s'\n", latestBlock.MinerAddress)
	fmt.Printf("P2P: listening on %s\n", cfg.Network.ListenAddress)
	if cfg.VPN.Enabled {
		fmt.Printf("VPN: interface %s\n", cfg.VPN.Interface)
	}
	if cfg.API.Enabled {
		fmt.Printf("API: listening on %s\n", cfg.API.ListenAddress)
	}

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("Shutting down...")

	// Graceful shutdown
	if apiServer != nil {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		apiServer.Stop(shutdownCtx)
	}
	if vpnServer != nil {
		vpnServer.Stop()
	}
	p2pNet.Stop()

	fmt.Printf("Final blockchain state: %d blocks\n", 1)
	fmt.Println("Shutdown complete")
}

// getDefaultConfig returns a default configuration
func getDefaultConfig() *config.Config {
	return &config.Config{
		Network: config.NetworkConfig{
			ListenAddress: "0.0.0.0:8333",
			MaxPeers:      50,
		},
		Mining: config.MiningConfig{
			Enabled:    false,
			GPUBackend: "cpu",
			Address:    "DVPNtest123...",
		},
		VPN: config.VPNConfig{
			Enabled:    false,
			Interface:  "dvpn0",
			ListenPort: 51820,
			PrivateKey: "DEMO_KEY",
		},
		API: config.APIConfig{
			Enabled:       true,
			ListenAddress: "127.0.0.1:8334",
			CorsEnabled:   true,
		},
		Logging: config.LoggingConfig{
			Level: "info",
			File:  "~/.dvpn/logs/node.log",
		},
	}
}
