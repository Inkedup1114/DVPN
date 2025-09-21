#!/bin/bash

set -e

echo "Building CUDA mining kernels..."

# Check if CUDA is available
if ! command -v nvcc &> /dev/null; then
    echo "CUDA compiler (nvcc) not found. Please install CUDA toolkit."
    exit 1
fi

# Compile CUDA kernel
cd pkg/mining/cuda/kernels
nvcc -c progpow.cu -o progpow.o \
    --compiler-options -fPIC \
    --gpu-architecture=sm_75 \
    --gpu-code=compute_75,sm_75,sm_80,sm_86

# Create static library
ar rcs libprogpow_cuda.a progpow.o

# Move to lib directory
mkdir -p ../../../../lib
mv libprogpow_cuda.a ../../../../lib/

echo "CUDA mining kernels built successfully!"
