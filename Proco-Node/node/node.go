package node

import (
	"time" // <-- ADD THIS
)

// node.go - Node structure and methods

// Node uses Config from config.go
type Node struct {
	Config     *Config
	Blockchain *Blockchain
}

// NewNode creates a node with default genesis
func NewNode(configPath string) (*Node, error) {
	// For now we ignore configPath for simplicity
	cfg := &Config{
		ChainID:          "proco-testnet",
		EpochDurationSec: 0,
		InitialSupply:    1000000,
		Timestamp:        time.Now(), // time.Now() is fine now because we imported "time"
	}
	bc := NewBlockchain()
	return &Node{
		Config:     cfg,
		Blockchain: bc,
	}, nil
}

// Start is a placeholder
func (n *Node) Start() error {
	// nothing fancy yet
	return nil
}
