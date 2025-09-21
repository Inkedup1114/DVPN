package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Inkedup1114/dvpn/pkg/p2p"
)

type PeersHandler struct {
	p2pNetwork *p2p.P2PNetwork
}

func NewPeersHandler(network *p2p.P2PNetwork) *PeersHandler {
	return &PeersHandler{
		p2pNetwork: network,
	}
}

type ConnectPeerRequest struct {
	Address string `json:"address"`
}

func (h *PeersHandler) ConnectPeer(w http.ResponseWriter, r *http.Request) {
	var req ConnectPeerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.p2pNetwork.ConnectToPeer(req.Address); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"status": "connected",
		"peer":   req.Address,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *PeersHandler) GetPeers(w http.ResponseWriter, r *http.Request) {
	peers := h.p2pNetwork.GetConnectedPeers()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"peers": peers,
		"count": len(peers),
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
