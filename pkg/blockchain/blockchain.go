package blockchain

import (
	"bytes"
	"encoding/binary"
	"errors"
	"sync"
)

type Blockchain struct {
	blocks     []*Block
	utxoSet    map[string]*TxOutput
	difficulty uint32
	mutex      sync.RWMutex
}

func NewBlockchain() *Blockchain {
	// Create genesis block
	genesis := &Block{
		Version:          1,
		PrevBlockHash:    make([]byte, 32),
		Timestamp:        1640995200, // Jan 1, 2022
		DifficultyTarget: 0x1e0ffff0,
		MinerAddress:     "genesis",
		Transactions:     []*Transaction{},
	}

	return &Blockchain{
		blocks:     []*Block{genesis},
		utxoSet:    make(map[string]*TxOutput),
		difficulty: 0x1e0ffff0,
	}
}

func (bc *Blockchain) AddBlock(block *Block) error {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	// Validate block using internal method (no additional locking)
	if !bc.isValidBlockInternal(block) {
		return errors.New("invalid block")
	}

	bc.blocks = append(bc.blocks, block)
	bc.updateUTXOSetInternal(block)

	return nil
}

func (bc *Blockchain) GetLatestBlock() *Block {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	return bc.getLatestBlockInternal()
}

// Internal method without locking
func (bc *Blockchain) getLatestBlockInternal() *Block {
	return bc.blocks[len(bc.blocks)-1]
}

// Internal method without locking
func (bc *Blockchain) isValidBlockInternal(block *Block) bool {
	latest := bc.getLatestBlockInternal() // Use internal method

	// Check if prev hash matches
	if !bytes.Equal(block.PrevBlockHash, latest.Hash()) {
		return false
	}

	// Check proof of work
	return bc.isValidProofOfWorkInternal(block)
}

// Public method with locking
// Note: public validation entrypoint is AddBlock which performs validation under lock.
// The internal helper `isValidBlockInternal` is used by AddBlock.

func (bc *Blockchain) isValidProofOfWorkInternal(block *Block) bool {
	// Simplified PoW validation - implement ProgPoW here
	hash := block.Hash()
	target := bc.calculateTargetInternal(block.DifficultyTarget)

	return binary.BigEndian.Uint32(hash[:4]) < target
}

func (bc *Blockchain) calculateTargetInternal(difficulty uint32) uint32 {
	// Safety check to prevent division by zero
	if difficulty == 0 {
		difficulty = 1 // Use minimum difficulty
	}

	// Simplified target calculation
	return 0xFFFFFFFF / difficulty
}

// Internal method without locking
func (bc *Blockchain) updateUTXOSetInternal(block *Block) {
	// Update UTXO set with new transactions
	for _, tx := range block.Transactions {
		// Remove spent outputs
		for _, input := range tx.Inputs {
			key := string(input.PrevTxHash) + string(rune(input.Index))
			delete(bc.utxoSet, key)
		}

		// Add new outputs
		for i, output := range tx.Outputs {
			key := string(tx.Hash) + string(rune(i))
			bc.utxoSet[key] = &output
		}
	}
}

// Public method with locking
func (bc *Blockchain) UpdateUTXOSet(block *Block) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	bc.updateUTXOSetInternal(block)
}

// GetMiningTarget returns the current difficulty target for mining
func (bc *Blockchain) GetMiningTarget() []byte {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	// Create an easy target for testing (4 zero bytes = very easy)
	target := make([]byte, 32)
	for i := 28; i < 32; i++ {
		target[i] = 0xFF
	}
	return target
}

// GetBlockTemplate creates a template for miners
func (bc *Blockchain) GetBlockTemplate(minerAddress string) *Block {
	bc.mutex.RLock()
	latest := bc.getLatestBlockInternal()
	bc.mutex.RUnlock()

	return NewBlock(latest.Hash(), []*Transaction{}, minerAddress)
}
