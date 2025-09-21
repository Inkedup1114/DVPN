package unit

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/Inkedup1114/dvpn/pkg/blockchain"
)

func TestBlockchainCreation(t *testing.T) {
    bc := blockchain.NewBlockchain()
    assert.NotNil(t, bc)
    
    latest := bc.GetLatestBlock()
    assert.Equal(t, "genesis", latest.MinerAddress)
}

func TestBlockAddition(t *testing.T) {
    bc := blockchain.NewBlockchain()
    
    // Create a new block with very easy difficulty for testing
    prevBlock := bc.GetLatestBlock()
    newBlock := blockchain.NewBlock(
        prevBlock.Hash(),
        []*blockchain.Transaction{},
        "test-miner",
    )
    
    // Set an easy difficulty target for testing (1 = easiest)
    newBlock.DifficultyTarget = 1
    
    err := bc.AddBlock(newBlock)
    assert.NoError(t, err)
    
    latest := bc.GetLatestBlock()
    assert.Equal(t, "test-miner", latest.MinerAddress)
}
