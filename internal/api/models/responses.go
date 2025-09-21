package models

type StatusResponse struct {
    Status      string `json:"status"`
    BlockHeight int64  `json:"block_height"`
    PeerCount   int    `json:"peer_count"`
    Mining      bool   `json:"mining"`
    VPNActive   bool   `json:"vpn_active"`
    Uptime      int64  `json:"uptime"`
}

type BlockchainInfo struct {
    Height     int64  `json:"height"`
    BestHash   string `json:"best_hash"`
    Difficulty uint32 `json:"difficulty"`
    TotalWork  string `json:"total_work"`
}

type MiningStatus struct {
    Enabled   bool    `json:"enabled"`
    Hashrate  float64 `json:"hashrate"`
    Target    string  `json:"target"`
    Backend   string  `json:"backend"`
}

type ErrorResponse struct {
    Error   string `json:"error"`
    Code    int    `json:"code"`
    Message string `json:"message"`
}
