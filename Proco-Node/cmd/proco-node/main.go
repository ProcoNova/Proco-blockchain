package main

import (
    "fmt"
    "log"
    "time"

    "github.com/yourname/proco-node/node"
)

func main() {
    // Initialize the node with genesis.json
    n, err := node.NewNode("config/genesis.json")
    if err != nil {
        log.Fatalf("node init: %v", err)
    }

    // Print the loaded config
    fmt.Printf("Node started with config: %+v\n", n.Config)

    // Start the node
    if err := n.Start(); err != nil {
        log.Fatalf("node start: %v", err)
    }

    // Add a new block manually
    n.Blockchain.AddBlock("First real block")

    // Print all blocks with hash
    for _, block := range n.Blockchain.Blocks {
        fmt.Printf("Block %d: %s | Hash: %s | PrevHash: %s\n", block.Index, block.Data, block.Hash, block.PrevHash)
    }

    // Automatically add a new block every 10 seconds
    blockCount := len(n.Blockchain.Blocks) // start from current number of blocks

    for {
        time.Sleep(10 * time.Second)
        newData := fmt.Sprintf("Block number %d", blockCount)
        n.Blockchain.AddBlock(newData)

        // Print the new block
        lastBlock := n.Blockchain.Blocks[len(n.Blockchain.Blocks)-1]
        fmt.Printf("Block %d added: %s | Hash: %s | PrevHash: %s\n", lastBlock.Index, lastBlock.Data, lastBlock.Hash, lastBlock.PrevHash)

        blockCount++
    }
}
