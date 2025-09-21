package blockchain

import (
	"crypto/sha256"
	"encoding/binary"
	"time"
)

type Block struct {
	Version          uint32         `json:"version"`
	PrevBlockHash    []byte         `json:"prev_block_hash"`
	MerkleRoot       []byte         `json:"merkle_root"`
	Timestamp        int64          `json:"timestamp"`
	DifficultyTarget uint32         `json:"difficulty_target"`
	Nonce            uint64         `json:"nonce"`
	MinerAddress     string         `json:"miner_address"`
	ExtraData        []byte         `json:"extra_data"`
	Transactions     []*Transaction `json:"transactions"`
}

type Transaction struct {
	Version   uint32          `json:"version"`
	Type      TransactionType `json:"type"`
	Inputs    []TxInput       `json:"inputs"`
	Outputs   []TxOutput      `json:"outputs"`
	Signature []byte          `json:"signature"`
	Hash      []byte          `json:"hash"`
}

type TransactionType uint8

const (
	TxTypeCoinbase TransactionType = iota
	TxTypeTransfer
	TxTypeNodeAd
	TxTypeBandwidthProof
	TxTypeMicropayment
)

type TxInput struct {
	PrevTxHash []byte `json:"prev_tx_hash"`
	Index      uint32 `json:"index"`
	ScriptSig  []byte `json:"script_sig"`
}

type TxOutput struct {
	Value        uint64 `json:"value"`
	ScriptPubKey []byte `json:"script_pub_key"`
}

func (b *Block) Hash() []byte {
	// Create a simple hash without potential circular references
	h := sha256.New()

	// Write fixed-size fields
	_ = binary.Write(h, binary.LittleEndian, b.Version)
	h.Write(b.PrevBlockHash)
	_ = binary.Write(h, binary.LittleEndian, b.Timestamp)
	_ = binary.Write(h, binary.LittleEndian, b.DifficultyTarget)
	_ = binary.Write(h, binary.LittleEndian, b.Nonce)
	h.Write([]byte(b.MinerAddress))

	// Simple transaction hash
	_ = binary.Write(h, binary.LittleEndian, uint32(len(b.Transactions)))

	hash := h.Sum(nil)
	return hash
}

func NewBlock(prevHash []byte, transactions []*Transaction, minerAddr string) *Block {
	return &Block{
		Version:          1,
		PrevBlockHash:    prevHash,
		Timestamp:        time.Now().Unix(),
		DifficultyTarget: 0x1e0ffff0, // Set a reasonable default difficulty
		MinerAddress:     minerAddr,
		Transactions:     transactions,
	}
}
