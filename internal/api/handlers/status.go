package handlers

import (
    "encoding/json"
    "net/http"
    "time"
    
    "github.com/Inkedup1114/dvpn/internal/api/models"
    "github.com/Inkedup1114/dvpn/pkg/p2p"
)

type StatusHandler struct {
    startTime  time.Time
    p2pNetwork *p2p.P2PNetwork
}

func NewStatusHandler(network *p2p.P2PNetwork) *StatusHandler {
    return &StatusHandler{
        startTime:  time.Now(),
        p2pNetwork: network,
    }
}

func (h *StatusHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
    var peerCount int
    if h.p2pNetwork != nil {
        peerCount = h.p2pNetwork.GetPeerCount()
    }
    
    response := models.StatusResponse{
        Status:      "running",
        BlockHeight: 1,
        PeerCount:   peerCount,
        Mining:      false,
        VPNActive:   false,
        Uptime:      int64(time.Since(h.startTime).Seconds()),
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func (h *StatusHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}
