#include <cuda_runtime.h>
#include <device_launch_parameters.h>
#include <stdint.h>

__device__ uint32_t cuda_rotl32(uint32_t x, int r) {
    return (x << r) | (x >> (32 - r));
}

__global__ void progpow_kernel(
    uint8_t* header_hash,
    uint64_t start_nonce,
    uint32_t* target,
    uint64_t* result_nonce,
    uint8_t* result_hash,
    uint32_t num_threads
) {
    uint32_t thread_id = blockIdx.x * blockDim.x + threadIdx.x;
    if (thread_id >= num_threads) return;
    
    uint64_t nonce = start_nonce + thread_id;
    
    // Simplified ProgPoW computation
    uint32_t mix[8];
    
    // Initialize mix with header hash
    for (int i = 0; i < 8; i++) {
        mix[i] = ((uint32_t*)header_hash)[i % 8];
    }
    
    // Add nonce
    mix[0] ^= (uint32_t)(nonce & 0xFFFFFFFF);
    mix[1] ^= (uint32_t)(nonce >> 32);
    
    // Simplified mixing rounds
    for (int round = 0; round < 18; round++) {
        for (int i = 0; i < 8; i++) {
            mix[i] = cuda_rotl32(mix[i] ^ (round * 0x9e3779b9), 7);
        }
    }
    
    // Check if result meets target
    bool meets_target = true;
    for (int i = 0; i < 8; i++) {
        if (mix[i] > target[i]) {
            meets_target = false;
            break;
        } else if (mix[i] < target[i]) {
            break;
        }
    }
    
    if (meets_target) {
        // Atomic operation to set result
        uint64_t old = atomicCAS((unsigned long long*)result_nonce, 0ULL, nonce);
        if (old == 0) {
            // Copy result hash
            for (int i = 0; i < 32; i++) {
                result_hash[i] = ((uint8_t*)mix)[i % 32];
            }
        }
    }
}

extern "C" {
    int cuda_progpow_mine(
        uint8_t* header_hash,
        uint64_t start_nonce,
        uint32_t* target,
        uint64_t* result_nonce,
        uint8_t* result_hash,
        uint32_t threads_per_block,
        uint32_t num_blocks
    ) {
        uint8_t* d_header_hash;
        uint32_t* d_target;
        uint64_t* d_result_nonce;
        uint8_t* d_result_hash;
        
        // Allocate GPU memory
        cudaMalloc(&d_header_hash, 32);
        cudaMalloc(&d_target, 32);
        cudaMalloc(&d_result_nonce, sizeof(uint64_t));
        cudaMalloc(&d_result_hash, 32);
        
        // Copy data to GPU
        cudaMemcpy(d_header_hash, header_hash, 32, cudaMemcpyHostToDevice);
        cudaMemcpy(d_target, target, 32, cudaMemcpyHostToDevice);
        cudaMemset(d_result_nonce, 0, sizeof(uint64_t));
        
        // Launch kernel
        uint32_t total_threads = threads_per_block * num_blocks;
        progpow_kernel<<<num_blocks, threads_per_block>>>(
            d_header_hash, start_nonce, d_target,
            d_result_nonce, d_result_hash, total_threads
        );
        
        // Wait for completion
        cudaDeviceSynchronize();
        
        // Copy results back
        cudaMemcpy(result_nonce, d_result_nonce, sizeof(uint64_t), cudaMemcpyDeviceToHost);
        cudaMemcpy(result_hash, d_result_hash, 32, cudaMemcpyDeviceToHost);
        
        // Cleanup
        cudaFree(d_header_hash);
        cudaFree(d_target);
        cudaFree(d_result_nonce);
        cudaFree(d_result_hash);
        
        return (*result_nonce != 0) ? 1 : 0;
    }
}
