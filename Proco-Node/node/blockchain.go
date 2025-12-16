package node

import (
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "time"
)

type Block struct {
    Index     int
    Timestamp time.Time
    Data      string
    PrevHash  string
    Hash      string
}

type Blockchain struct {
    Blocks []*Block
}

// Helper to calculate SHA256 hash of a block
func calculateHash(block *Block) string {
    record := fmt.Sprintf("%d%s%s%s", block.Index, block.Timestamp, block.Data, block.PrevHash)
    h := sha256.New()
    h.Write([]byte(record))
    hashed := h.Sum(nil)
    return hex.EncodeToString(hashed)
}

// Add a new block to the blockchain
func (bc *Blockchain) AddBlock(data string) {
    var prevHash string
    index := len(bc.Blocks)
    if index > 0 {
        prevHash = bc.Blocks[index-1].Hash
    }
    block := &Block{
        Index:     index,
        Timestamp: time.Now(),
        Data:      data,
        PrevHash:  prevHash,
    }
    block.Hash = calculateHash(block) // calculate and assign hash
    bc.Blocks = append(bc.Blocks, block)
}

// Create a new blockchain with genesis block
func NewBlockchain() *Blockchain {
    bc := &Blockchain{}
    bc.AddBlock("Genesis Block")
    return bc
}
