#!/bin/bash

set -e

echo "Building DVPN Docker images..."

# Build base image
docker build -t dvpn-base -f docker/Dockerfile.base .

# Build node image
docker build -t dvpn-node -f docker/Dockerfile.node .

# Build miner image
docker build -t dvpn-miner -f docker/Dockerfile.miner .

echo "Docker images built successfully!"
