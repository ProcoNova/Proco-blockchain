package main

import (
	"fmt"
	"log"
	"time"

	"proco-node/node"
)

// Startup banner
func printStartupBanner(nodeID string, chainID string) {
	fmt.Println("=======================================")
	fmt.Println("           ProCo Node Started           ")
	fmt.Println("=======================================")
	fmt.Printf("Node ID  : %s\n", nodeID)
	fmt.Printf("Network  : %s\n", chainID)
	fmt.Println("=======================================")
	fmt.Println()
}

// ✅ FIXED: accepts *node.Block (pointer)
func printBlock(block *node.Block) {
	hash := block.Hash
	prev := block.PrevHash

	if len(hash) > 8 {
		hash = hash[:8]
	}
	if len(prev) > 8 {
		prev = prev[:8]
	}

	if block.Index == 0 {
		fmt.Printf("[GENESIS] Block %d | Hash: %s | Prev: %s\n", block.Index, hash, prev)
	} else {
		fmt.Printf("[BLOCK]   Block %d | Hash: %s | Prev: %s\n", block.Index, hash, prev)
	}
}

func main() {
	// --- CHANGE THIS LINE FOR NODE ID ---
	nodeID := "NODE-001" // Use "NODE-002" for Node2

	// Initialize node
	n, err := node.NewNode("configs/genesis.json")
	if err != nil {
		log.Fatalf("node init failed: %v", err)
	}

	printStartupBanner(nodeID, n.Config.ChainID)

	// Start node
	if err := n.Start(); err != nil {
		log.Fatalf("node start failed: %v", err)
	}

	// Print existing blocks
	for _, block := range n.Blockchain.Blocks {
		printBlock(block)
	}

	// Auto-generate new blocks
	blockCount := len(n.Blockchain.Blocks)
	for {
		time.Sleep(5 * time.Second)
		data := fmt.Sprintf("Block number %d", blockCount)
		n.Blockchain.AddBlock(data)

		last := n.Blockchain.Blocks[len(n.Blockchain.Blocks)-1]
		printBlock(last)

		blockCount++
	}
}
