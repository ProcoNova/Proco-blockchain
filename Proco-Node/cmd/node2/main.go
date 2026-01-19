package main

import (
	"fmt"

	"proco-node/node"
)

func main() {
	fmt.Println("Starting Node 2...")

	n, err := node.NewNode("../configs/genesis.json")
	if err != nil {
		panic(err)
	}

	n.ListenPort = "3002"
	n.Peers = []string{"127.0.0.1:3001"}

	go n.Listen()

	select {} // keep running
}
