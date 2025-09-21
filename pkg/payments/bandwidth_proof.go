package payments

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type BandwidthSession struct {
	ID           []byte            `json:"id"`
	NodePubKey   ed25519.PublicKey `json:"node_pubkey"`
	ClientPubKey ed25519.PublicKey `json:"client_pubkey"`
	StartTime    int64             `json:"start_time"`
	EndTime      int64             `json:"end_time"`
	BytesSent    uint64            `json:"bytes_sent"`
	BytesRecv    uint64            `json:"bytes_recv"`
	ProofHash    []byte            `json:"proof_hash"`
	Audits       []*AuditResult    `json:"audits"`
	mutex        sync.RWMutex
}

type AuditResult struct {
	AuditorPubKey ed25519.PublicKey `json:"auditor_pubkey"`
	MeasuredSpeed uint64            `json:"measured_speed_mbps"`
	Latency       uint64            `json:"latency_ms"`
	Timestamp     int64             `json:"timestamp"`
	Signature     []byte            `json:"signature"`
	Valid         bool              `json:"valid"`
}

type BandwidthProofManager struct {
	sessions map[string]*BandwidthSession
	auditors []ed25519.PublicKey
	mutex    sync.RWMutex
}

func NewBandwidthProofManager() *BandwidthProofManager {
	return &BandwidthProofManager{
		sessions: make(map[string]*BandwidthSession),
		auditors: make([]ed25519.PublicKey, 0),
	}
}

func (bpm *BandwidthProofManager) StartSession(
	nodePubKey, clientPubKey ed25519.PublicKey,
) (*BandwidthSession, error) {
	// Generate session ID
	sessionID := make([]byte, 32)
	if _, err := rand.Read(sessionID); err != nil {
		return nil, err
	}

	session := &BandwidthSession{
		ID:           sessionID,
		NodePubKey:   nodePubKey,
		ClientPubKey: clientPubKey,
		StartTime:    time.Now().Unix(),
		BytesSent:    0,
		BytesRecv:    0,
		Audits:       make([]*AuditResult, 0),
	}

	bpm.mutex.Lock()
	bpm.sessions[string(sessionID)] = session
	bpm.mutex.Unlock()

	return session, nil
}

func (bpm *BandwidthProofManager) EndSession(sessionID []byte) (*BandwidthSession, error) {
	bpm.mutex.Lock()
	defer bpm.mutex.Unlock()

	session, exists := bpm.sessions[string(sessionID)]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	session.mutex.Lock()
	session.EndTime = time.Now().Unix()
	session.ProofHash = bpm.calculateProofHash(session)
	session.mutex.Unlock()

	return session, nil
}

func (bpm *BandwidthProofManager) UpdateSessionStats(
	sessionID []byte,
	bytesSent, bytesRecv uint64,
) error {
	bpm.mutex.RLock()
	session, exists := bpm.sessions[string(sessionID)]
	bpm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("session not found")
	}

	session.mutex.Lock()
	session.BytesSent += bytesSent
	session.BytesRecv += bytesRecv
	session.mutex.Unlock()

	return nil
}

func (bpm *BandwidthProofManager) AddAuditResult(
	sessionID []byte,
	audit *AuditResult,
) error {
	bpm.mutex.RLock()
	session, exists := bpm.sessions[string(sessionID)]
	bpm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("session not found")
	}

	// Validate audit signature
	if !bpm.validateAuditSignature(audit) {
		return fmt.Errorf("invalid audit signature")
	}

	session.mutex.Lock()
	session.Audits = append(session.Audits, audit)
	session.mutex.Unlock()

	return nil
}

func (bpm *BandwidthProofManager) calculateProofHash(session *BandwidthSession) []byte {
	// Create deterministic hash of session data
	data := struct {
		ID           []byte `json:"id"`
		NodePubKey   []byte `json:"node_pubkey"`
		ClientPubKey []byte `json:"client_pubkey"`
		StartTime    int64  `json:"start_time"`
		EndTime      int64  `json:"end_time"`
		BytesSent    uint64 `json:"bytes_sent"`
		BytesRecv    uint64 `json:"bytes_recv"`
	}{
		ID:           session.ID,
		NodePubKey:   session.NodePubKey,
		ClientPubKey: session.ClientPubKey,
		StartTime:    session.StartTime,
		EndTime:      session.EndTime,
		BytesSent:    session.BytesSent,
		BytesRecv:    session.BytesRecv,
	}

	jsonData, _ := json.Marshal(data)
	hash := sha256.Sum256(jsonData)
	return hash[:]
}

func (bpm *BandwidthProofManager) validateAuditSignature(audit *AuditResult) bool {
	// Create message to verify
	message := struct {
		MeasuredSpeed uint64 `json:"measured_speed_mbps"`
		Latency       uint64 `json:"latency_ms"`
		Timestamp     int64  `json:"timestamp"`
	}{
		MeasuredSpeed: audit.MeasuredSpeed,
		Latency:       audit.Latency,
		Timestamp:     audit.Timestamp,
	}

	msgBytes, _ := json.Marshal(message)
	return ed25519.Verify(audit.AuditorPubKey, msgBytes, audit.Signature)
}

func (bpm *BandwidthProofManager) CreateProof(
	sessionID []byte,
	nodePrivateKey ed25519.PrivateKey,
) (*BandwidthProofMessage, error) {
	session, err := bpm.EndSession(sessionID)
	if err != nil {
		return nil, err
	}

	// Create proof message
	proof := &BandwidthProofMessage{
		SessionIDHash: session.ProofHash,
		BytesSent:     session.BytesSent,
		BytesReceived: session.BytesRecv,
		AuditSigs:     make([]AuditSignature, len(session.Audits)),
	}

	// Convert audit results to signatures
	for i, audit := range session.Audits {
		proof.AuditSigs[i] = AuditSignature{
			AuditorPubKey: audit.AuditorPubKey,
			Signature:     audit.Signature,
			Timestamp:     audit.Timestamp,
		}
	}

	return proof, nil
}

type BandwidthProofMessage struct {
	SessionIDHash []byte           `json:"session_id_hash"`
	BytesSent     uint64           `json:"bytes_sent"`
	BytesReceived uint64           `json:"bytes_received"`
	AuditSigs     []AuditSignature `json:"audit_sigs"`
}

type AuditSignature struct {
	AuditorPubKey ed25519.PublicKey `json:"auditor_pubkey"`
	Signature     []byte            `json:"signature"`
	Timestamp     int64             `json:"timestamp"`
}
