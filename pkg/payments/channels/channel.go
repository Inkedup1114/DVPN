package channels

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type PaymentChannel struct {
	ID            []byte            `json:"id"`
	Sender        ed25519.PublicKey `json:"sender"`
	Recipient     ed25519.PublicKey `json:"recipient"`
	InitialAmount uint64            `json:"initial_amount"`
	CurrentAmount uint64            `json:"current_amount"`
	SequenceNum   uint64            `json:"sequence_num"`
	ExpiryBlock   uint64            `json:"expiry_block"`
	State         ChannelState      `json:"state"`
	mutex         sync.RWMutex
}

type ChannelState uint8

const (
	StateOpen ChannelState = iota
	StateClosed
	StateDisputed
)

type PaymentUpdate struct {
	ChannelID    []byte `json:"channel_id"`
	SequenceNum  uint64 `json:"sequence_num"`
	Amount       uint64 `json:"amount"`
	RecipientSig []byte `json:"recipient_sig"`
	SenderSig    []byte `json:"sender_sig"`
	Timestamp    int64  `json:"timestamp"`
}

type ChannelManager struct {
	channels map[string]*PaymentChannel
	mutex    sync.RWMutex
}

func NewChannelManager() *ChannelManager {
	return &ChannelManager{
		channels: make(map[string]*PaymentChannel),
	}
}

func (cm *ChannelManager) OpenChannel(
	id []byte,
	sender, recipient ed25519.PublicKey,
	amount uint64,
	expiryBlock uint64,
) (*PaymentChannel, error) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	channelKey := string(id)
	if _, exists := cm.channels[channelKey]; exists {
		return nil, fmt.Errorf("channel already exists")
	}

	channel := &PaymentChannel{
		ID:            id,
		Sender:        sender,
		Recipient:     recipient,
		InitialAmount: amount,
		CurrentAmount: amount,
		SequenceNum:   0,
		ExpiryBlock:   expiryBlock,
		State:         StateOpen,
	}

	cm.channels[channelKey] = channel
	return channel, nil
}

func (cm *ChannelManager) GetChannel(id []byte) (*PaymentChannel, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	channel, exists := cm.channels[string(id)]
	if !exists {
		return nil, fmt.Errorf("channel not found")
	}

	return channel, nil
}

func (cm *ChannelManager) UpdateChannel(update *PaymentUpdate) error {
	channel, err := cm.GetChannel(update.ChannelID)
	if err != nil {
		return err
	}

	channel.mutex.Lock()
	defer channel.mutex.Unlock()

	// Validate sequence number
	if update.SequenceNum <= channel.SequenceNum {
		return fmt.Errorf("invalid sequence number")
	}

	// Validate amount
	if update.Amount > channel.InitialAmount {
		return fmt.Errorf("amount exceeds channel capacity")
	}

	// Validate signatures (simplified)
	if !cm.validateSignatures(channel, update) {
		return fmt.Errorf("invalid signatures")
	}

	// Update channel state
	channel.CurrentAmount = update.Amount
	channel.SequenceNum = update.SequenceNum

	return nil
}

func (cm *ChannelManager) CloseChannel(channelID []byte) error {
	channel, err := cm.GetChannel(channelID)
	if err != nil {
		return err
	}

	channel.mutex.Lock()
	defer channel.mutex.Unlock()

	if channel.State != StateOpen {
		return fmt.Errorf("channel not open")
	}

	channel.State = StateClosed
	return nil
}

func (cm *ChannelManager) validateSignatures(channel *PaymentChannel, update *PaymentUpdate) bool {
	// Create message to verify
	message := struct {
		ChannelID   []byte `json:"channel_id"`
		SequenceNum uint64 `json:"sequence_num"`
		Amount      uint64 `json:"amount"`
	}{
		ChannelID:   update.ChannelID,
		SequenceNum: update.SequenceNum,
		Amount:      update.Amount,
	}

	msgBytes, _ := json.Marshal(message)

	// Verify sender signature
	if !ed25519.Verify(channel.Sender, msgBytes, update.SenderSig) {
		return false
	}

	// Verify recipient signature
	if !ed25519.Verify(channel.Recipient, msgBytes, update.RecipientSig) {
		return false
	}

	return true
}

func (c *PaymentChannel) CreateUpdate(amount uint64, senderKey ed25519.PrivateKey) (*PaymentUpdate, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.State != StateOpen {
		return nil, fmt.Errorf("channel not open")
	}

	if amount > c.InitialAmount {
		return nil, fmt.Errorf("amount exceeds channel capacity")
	}

	newSeqNum := c.SequenceNum + 1

	// Create update message
	message := struct {
		ChannelID   []byte `json:"channel_id"`
		SequenceNum uint64 `json:"sequence_num"`
		Amount      uint64 `json:"amount"`
	}{
		ChannelID:   c.ID,
		SequenceNum: newSeqNum,
		Amount:      amount,
	}

	msgBytes, _ := json.Marshal(message)

	// Sign with sender key
	senderSig := ed25519.Sign(senderKey, msgBytes)

	return &PaymentUpdate{
		ChannelID:   c.ID,
		SequenceNum: newSeqNum,
		Amount:      amount,
		SenderSig:   senderSig,
		Timestamp:   time.Now().Unix(),
	}, nil
}
