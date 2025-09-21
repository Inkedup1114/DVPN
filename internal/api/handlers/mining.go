package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Inkedup1114/dvpn/internal/api/models"
	"github.com/Inkedup1114/dvpn/pkg/mining"
)

type MiningHandler struct {
	miner   mining.Miner
	ctx     context.Context
	cancel  context.CancelFunc
	running bool
}

func NewMiningHandler(miner mining.Miner) *MiningHandler {
	return &MiningHandler{
		miner: miner,
	}
}

func (h *MiningHandler) StartMining(w http.ResponseWriter, r *http.Request) {
	if h.miner == nil {
		http.Error(w, "No miner configured", http.StatusBadRequest)
		return
	}

	if h.running {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "already running"}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Set an easy target for testing
	target := make([]byte, 32)
	target[28] = 0x00
	target[29] = 0x00
	target[30] = 0x00
	target[31] = 0xFF

	h.miner.SetTarget(target)

	h.ctx, h.cancel = context.WithCancel(context.Background())
	if err := h.miner.Start(h.ctx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.running = true
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "mining started"}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *MiningHandler) StopMining(w http.ResponseWriter, r *http.Request) {
	if !h.running {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "not running"}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	h.cancel()
	if err := h.miner.Stop(); err != nil {
		// Log error but continue to return response
		// (in real app, use structured logging)
		_ = err
	}
	h.running = false

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "mining stopped"}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *MiningHandler) GetMiningStatus(w http.ResponseWriter, r *http.Request) {
	var hashrate float64
	if h.miner != nil {
		hashrate = h.miner.GetHashrate()
	}

	status := models.MiningStatus{
		Enabled:  h.running,
		Hashrate: hashrate,
		Target:   "0x1e0ffff0",
		Backend:  "cpu",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
