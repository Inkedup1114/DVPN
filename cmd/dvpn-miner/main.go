package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Inkedup1114/dvpn/pkg/mining/cpu"
)

func main() {
	var backend = flag.String("backend", "cpu", "Mining backend (cpu)")
	flag.Parse()

	fmt.Printf("DVPN Miner starting with %s backend...\n", *backend)

	// Initialize CPU miner (for now)
	miner := cpu.NewCPUMiner()

	// Start mining
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := miner.Start(ctx); err != nil {
		log.Fatalf("Failed to start miner: %v", err)
	}

	fmt.Println("Miner started successfully!")
	fmt.Println("Press Ctrl+C to stop...")

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("Stopping miner...")
	if err := miner.Stop(); err != nil {
		log.Printf("Miner stop error: %v", err)
	}
	fmt.Println("Miner stopped")
}
