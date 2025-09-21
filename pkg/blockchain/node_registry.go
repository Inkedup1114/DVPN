package blockchain

import (
    "crypto/ed25519"
    "encoding/json"
    "errors"
    "sync"
)

type NodeAdvertisement struct {
    PubKey              ed25519.PublicKey `json:"pubkey"`
    IPAddress           string            `json:"ip_address"`
    Port                uint16            `json:"port"`
    LocationHash        string            `json:"location_hash"`
    AvailableBandwidth  uint64            `json:"available_bandwidth_mbps"`
    MinPricePerMB       uint64            `json:"min_price_per_mb"`
    NodeType            string            `json:"node_type"`
    ExpiryBlock         uint64            `json:"expiry_block"`
    Signature           []byte            `json:"signature"`
}

type NodeRegistry struct {
    nodes map[string]*NodeAdvertisement
    mutex sync.RWMutex
}

func NewNodeRegistry() *NodeRegistry {
    return &NodeRegistry{
        nodes: make(map[string]*NodeAdvertisement),
    }
}

func (nr *NodeRegistry) RegisterNode(ad *NodeAdvertisement) error {
    nr.mutex.Lock()
    defer nr.mutex.Unlock()
    
    // Validate signature
    if !nr.validateNodeSignature(ad) {
        return errors.New("invalid node signature")
    }
    
    key := string(ad.PubKey)
    nr.nodes[key] = ad
    
    return nil
}

func (nr *NodeRegistry) GetActiveNodes() []*NodeAdvertisement {
    nr.mutex.RLock()
    defer nr.mutex.RUnlock()
    
    var active []*NodeAdvertisement
    currentBlock := getCurrentBlockHeight() // Implement this
    
    for _, node := range nr.nodes {
        if node.ExpiryBlock > currentBlock {
            active = append(active, node)
        }
    }
    
    return active
}

func (nr *NodeRegistry) validateNodeSignature(ad *NodeAdvertisement) bool {
    // Create message to verify
    data, _ := json.Marshal(struct {
        PubKey             ed25519.PublicKey `json:"pubkey"`
        IPAddress          string            `json:"ip_address"`
        Port               uint16            `json:"port"`
        LocationHash       string            `json:"location_hash"`
        AvailableBandwidth uint64            `json:"available_bandwidth_mbps"`
        MinPricePerMB      uint64            `json:"min_price_per_mb"`
        NodeType           string            `json:"node_type"`
    }{
        PubKey:             ad.PubKey,
        IPAddress:          ad.IPAddress,
        Port:               ad.Port,
        LocationHash:       ad.LocationHash,
        AvailableBandwidth: ad.AvailableBandwidth,
        MinPricePerMB:      ad.MinPricePerMB,
        NodeType:           ad.NodeType,
    })
    
    return ed25519.Verify(ad.PubKey, data, ad.Signature)
}

func getCurrentBlockHeight() uint64 {
    // Implement: return current blockchain height
    return 0
}
