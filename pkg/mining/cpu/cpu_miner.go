package cpu

import (
    "fmt"
    "context"
    "crypto/sha256"
    "time"
    
    "github.com/Inkedup1114/dvpn/pkg/mining"
)

type CPUMiner struct {
    baseMiner *mining.BaseMiner
}

func NewCPUMiner() *CPUMiner {
    return &CPUMiner{
        baseMiner: mining.NewBaseMiner(),
    }
}

func (m *CPUMiner) Start(ctx context.Context) error {
    m.baseMiner.SetRunning(true)
    go m.mineLoop(ctx)
    return nil
}

func (m *CPUMiner) Stop() error {
    m.baseMiner.SetRunning(false)
    return nil
}

func (m *CPUMiner) SetTarget(target []byte) {
    fmt.Printf("DEBUG: CPU miner received target: %x\n", target[:8])
    m.baseMiner.SetTarget(target)
}

func (m *CPUMiner) GetHashrate() float64 {
    return m.baseMiner.GetHashrate()
}

func (m *CPUMiner) Results() <-chan *mining.MiningResult {
    return m.baseMiner.Results()
}

func (m *CPUMiner) mineLoop(ctx context.Context) {
    nonce := uint64(0)
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()
    
    var hashCount uint64
    startTime := time.Now()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            // Update hashrate
            elapsed := time.Since(startTime).Seconds()
            if elapsed > 0 {
                m.baseMiner.SetHashrate(float64(hashCount) / elapsed)
            }
        default:
            if !m.baseMiner.IsRunning() {
                return
            }
            
            // Get current target
            target := m.baseMiner.GetTarget()
            
            if len(target) == 0 {
                time.Sleep(100 * time.Millisecond)
                continue
            }
            
            // Simple CPU mining (very basic)
            result := m.mineBatch(nonce, target)
            if result != nil {
                m.baseMiner.SendResult(result)
            }
            
            nonce++
            hashCount++
            
            // CPU mining is slow, so we don't need tight loop
            time.Sleep(time.Millisecond)
        }
    }
}

func (m *CPUMiner) mineBatch(nonce uint64, target []byte) *mining.MiningResult {
    // Simple SHA256 mining for testing
    data := make([]byte, 40)
    copy(data[:32], target) // Use target as header hash for demo
    data[32] = byte(nonce)
    data[33] = byte(nonce >> 8)
    data[34] = byte(nonce >> 16)
    data[35] = byte(nonce >> 24)
    data[36] = byte(nonce >> 32)
    data[37] = byte(nonce >> 40)
    data[38] = byte(nonce >> 48)
    data[39] = byte(nonce >> 56)
    
    hash := sha256.Sum256(data)
    
    // Much easier target for demo - just check if hash[31] < 128 (50% chance)
    if hash[31] < 128 {
        return &mining.MiningResult{
            Nonce:     nonce,
            Hash:      hash[:],
            Timestamp: time.Now(),
        }
    }
    
    return nil
}
