#!/bin/bash

set -e

echo "Building DVPN system..."

# Set up Go build environment
export CGO_ENABLED=1
export GOOS=linux
export GOARCH=amd64

# Build directories
mkdir -p bin

# Build blockchain node
echo "Building blockchain node..."
go build -o bin/dvpnd ./cmd/dvpnd

# Build miner (with CUDA support if available)
echo "Building mining software..."
if command -v nvcc &> /dev/null; then
    echo "CUDA found, building with GPU support..."
    ./scripts/build-cuda.sh
    go build -tags cuda -o bin/dvpn-miner ./cmd/dvpn-miner
else
    echo "CUDA not found, building CPU-only version..."
    go build -o bin/dvpn-miner ./cmd/dvpn-miner
fi

# Build CLI tools
echo "Building CLI tools..."
go build -o bin/dvpn-cli ./cmd/dvpn-cli
go build -o bin/dvpn-wallet ./cmd/dvpn-wallet

echo "Build complete!"
echo "Binaries available in ./bin/"
