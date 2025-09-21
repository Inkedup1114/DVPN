package mining

import (
    "context"
    "sync"
    "time"
)

type Miner interface {
    Start(ctx context.Context) error
    Stop() error
    SetTarget(target []byte)
    GetHashrate() float64
    Results() <-chan *MiningResult
}

type MiningResult struct {
    Nonce     uint64
    Hash      []byte
    Timestamp time.Time
}

type BaseMiner struct {
    target   []byte
    hashrate float64
    running  bool
    mutex    sync.RWMutex
    results  chan *MiningResult
}

func NewBaseMiner() *BaseMiner {
    return &BaseMiner{
        results: make(chan *MiningResult, 100),
    }
}

func (m *BaseMiner) SetTarget(target []byte) {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    m.target = make([]byte, len(target))
    copy(m.target, target)
}

func (m *BaseMiner) GetHashrate() float64 {
    m.mutex.RLock()
    defer m.mutex.RUnlock()
    return m.hashrate
}

func (m *BaseMiner) IsRunning() bool {
    m.mutex.RLock()
    defer m.mutex.RUnlock()
    return m.running
}

func (m *BaseMiner) Results() <-chan *MiningResult {
    return m.results
}

// SendResult allows derived miners to send results
func (m *BaseMiner) SendResult(result *MiningResult) {
    select {
    case m.results <- result:
    default:
        // Drop result if channel is full
    }
}

func (m *BaseMiner) SetRunning(running bool) {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    m.running = running
}

func (m *BaseMiner) SetHashrate(hashrate float64) {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    m.hashrate = hashrate
}

func (m *BaseMiner) GetTarget() []byte {
    m.mutex.RLock()
    defer m.mutex.RUnlock()
    target := make([]byte, len(m.target))
    copy(target, m.target)
    return target
}
