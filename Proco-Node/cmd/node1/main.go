package main

import (
	"fmt"
	"time"

	"proco-node/node"
)

func main() {
	fmt.Println("Starting Node 1...")

	n, err := node.NewNode("../configs/genesis.json")
	if err != nil {
		panic(err)
	}

	n.ListenPort = "3001"
	n.Peers = []string{"127.0.0.1:3002"}

	go n.Listen()

	time.Sleep(2 * time.Second)

	// create some transactions and mine
	n.AddTransaction("NODE1", "NODE2", 10)
	block := n.MinePendingTxs()
	if block != nil {
		n.SendBlock(block)
	}

	select {} // keep running
}
