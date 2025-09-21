# DVPN - Decentralized VPN + GPU Mining Protocol
# DVPN - Decentralized VPN + GPU Mining Protocol

A blockchain-based decentralized VPN network where participants run nodes to earn cryptocurrency while providing VPN services. Thi
s system combines proof-of-work mining with bandwidth relay services to create economic incentives for network participation.     
## Project Status

**Current Version**: v0.1.0-alpha  
**Status**: Working Prototype  

### What's Working
- ✅ Blockchain consensus with PoW mining
- ✅ CPU mining with 800+ H/s performance
- ✅ REST API for node management
- ✅ P2P networking foundation
- ✅ Configuration management
- ✅ CLI tools and status monitoring

### Known Limitations
- ⚠️ P2P handshake issues (connection debugging in progress)
- ⚠️ VPN key configuration needs refinement
- ⚠️ No blockchain persistence (data lost on restart)
- ⚠️ Missing payment channel implementation
- ⚠️ No bandwidth proof verification

## Architecture Overview

```
┌─────────────────────────────────────────────────────────┐
│                    DVPN Node                            │
├─────────────────┬─────────────────┬─────────────────────┤
│   REST API      │   P2P Network   │   Mining Engine     │
│   Port: 8334    │   Port: 8333    │   CPU/GPU Based     │
├─────────────────┼─────────────────┼─────────────────────┤
│              Blockchain Layer                           │
│        (PoW Consensus + Node Registry)                  │
├─────────────────────────────────────────────────────────┤
│              VPN Layer (WireGuard)                      │
│         (Bandwidth Relay + Payment Channels)           │
└─────────────────────────────────────────────────────────┘
```

### Key Components

- **Blockchain Core**: ProgPoW consensus with node advertisements
- **Mining Engine**: CPU-based mining (GPU support planned)
- **VPN Layer**: WireGuard-based relay with bandwidth tracking
- **P2P Network**: Node discovery and blockchain synchronization
- **Payment System**: Micropayment channels for bandwidth compensation
- **REST API**: Node management and monitoring interface

## Installation

### System Requirements

- **OS**: Ubuntu 24.04 LTS (other Linux distributions may work)
- **RAM**: Minimum 4GB, recommended 8GB+
- **Storage**: 10GB+ free space
- **Network**: Broadband connection for VPN relay
- **CPU**: Multi-core processor (GPU optional for mining)

### Dependencies

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install essential build tools
sudo apt install -y \
    build-essential \
    cmake \
    git \
    curl \
    wget \
    pkg-config \
    libssl-dev \
    libboost-all-dev

# Install GPU development tools (optional)
sudo apt install -y \
    opencl-headers \
    opencl-dev \
    clinfo

# NVIDIA CUDA (if using NVIDIA GPUs)
# wget https://developer.download.nvidia.com/compute/cuda/repos/ubuntu2404/x86_64/cuda-keyring_1.1-1_all.deb
# sudo dpkg -i cuda-keyring_1.1-1_all.deb
# sudo apt update && sudo apt install -y cuda-toolkit-12-6

# Install programming languages
# Go 1.21+
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Additional libraries
sudo apt install -y \
    libleveldb-dev \
    libzmq3-dev \
    libsodium-dev \
    protobuf-compiler \
    wireguard \
    wireguard-tools
```

### Build from Source

```bash
# Clone the repository
git clone https://github.com/yourusername/dvpn-project.git
cd dvpn-project

# Download dependencies
go mod tidy

# Build all components
./scripts/build.sh

# Verify installation
./bin/dvpn-cli version
```

## Quick Start

### 1. Configure Your Node

```bash
# Generate WireGuard keys
wg genkey > ~/.dvpn/private.key
chmod 600 ~/.dvpn/private.key

# Edit node configuration
cp configs/node.yaml ~/.dvpn/node.yaml
nano ~/.dvpn/node.yaml
```

Example configuration:
```yaml
network:
  listen_address: "0.0.0.0:8333"
  max_peers: 50

mining:
  enabled: true
  gpu_backend: "cpu"
  miner_address: "your-dvpn-address"

vpn:
  enabled: true
  interface: "dvpn0"
  listen_port: 51820
  private_key_file: "~/.dvpn/private.key"

api:
  enabled: true
  listen_address: "127.0.0.1:8334"
```

### 2. Start Your Node

```bash
# Start the DVPN node
./bin/dvpnd -config ~/.dvpn/node.yaml

# In another terminal, check status
./bin/dvpn-cli status
```

### 3. Start Mining (Optional)

```bash
# Start mining via API
curl -X POST http://127.0.0.1:8334/api/mining/start

# Check mining status
curl http://127.0.0.1:8334/api/mining/status
```

### 4. Connect to Network

```bash
# Connect to another node
curl -X POST http://127.0.0.1:8334/api/peers/connect \
     -H "Content-Type: application/json" \
     -d '{"address":"peer-ip:8333"}'

# Check peer connections
curl http://127.0.0.1:8334/api/peers
```

## API Documentation

### Node Status
```http
GET /api/status
```
