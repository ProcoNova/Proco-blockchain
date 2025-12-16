package node

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// StartNode starts the command loop for your blockchain node
func StartNode() {
	bc, err := LoadBlockchain("blocks.json")
	if err != nil {
		fmt.Println("Error loading blockchain:", err)
		return
	}

	// Load wallets
	err = bc.LoadWallets("wallets.json")
	if err != nil {
		fmt.Println("Error loading wallets:", err)
		return
	}

	fmt.Println("🚀 Starting ProCo Node...")
	fmt.Println("✅ Node is now running. Type 'help' for commands.")

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		parts := strings.Split(input, " ")

		if len(parts) == 0 {
			continue
		}

		command := parts[0]

		switch command {

		// ---------------- HELP ----------------
		case "help":
			fmt.Println("\nAvailable commands:")
			fmt.Println(" show_chain")
			fmt.Println(" add_block <data>")
			fmt.Println(" validate")
			fmt.Println(" create_wallet <initial_balance>")
			fmt.Println(" list_wallets")
			fmt.Println(" send <from_address> <to_address> <amount>")
			fmt.Println(" balance <wallet_address>")
			fmt.Println(" exit")

		// ---------------- SHOW CHAIN ----------------
		case "show_chain":
			bc.ShowChain()

		// ---------------- ADD BLOCK ----------------
		case "add_block":
			if len(parts) < 2 {
				fmt.Println("Usage: add_block <data>")
				continue
			}
			data := strings.Join(parts[1:], " ")
			bc.AddBlock(data)
			fmt.Println("✅ Block added and blockchain saved.")

		// ---------------- VALIDATE ----------------
		case "validate":
			bc.ValidateChain()

		// ---------------- CREATE WALLET ----------------
		case "create_wallet":
			if len(parts) != 2 {
				fmt.Println("Usage: create_wallet <initial_balance>")
				continue
			}
			balance, err := strconv.Atoi(parts[1])
			if err != nil {
				fmt.Println("❌ Enter a valid number for balance")
				continue
			}
			wallet := NewWallet(balance)
			bc.Wallets = append(bc.Wallets, wallet)
			err = bc.SaveWallets("wallets.json")
			if err != nil {
				fmt.Println("Error saving wallet:", err)
				continue
			}

			fmt.Println("\n✅ Wallet created successfully!")
			fmt.Println(" Address :", wallet.Address)
			fmt.Println(" Balance :", wallet.Balance)

		// ---------------- LIST WALLETS ----------------
		case "list_wallets":
			if len(bc.Wallets) == 0 {
				fmt.Println("No wallets found.")
				continue
			}
			fmt.Println("\n💼 Wallets:")
			for i, w := range bc.Wallets {
				fmt.Printf("%d) %s | Balance: %d\n", i+1, w.Address, w.Balance)
			}

		// ---------------- SEND COINS ----------------
		case "send":
			if len(parts) != 4 {
				fmt.Println("Usage: send <from_address> <to_address> <amount>")
				continue
			}
			from := parts[1]
			to := parts[2]
			amount, err := strconv.Atoi(parts[3])
			if err != nil {
				fmt.Println("❌ Invalid amount")
				continue
			}
			SendCoins(bc, from, to, amount)

		// ---------------- BALANCE ----------------
		case "balance":
			if len(parts) != 2 {
				fmt.Println("Usage: balance <wallet_address>")
				continue
			}
			address := parts[1]
			balance := GetBalance(bc, address)
			fmt.Printf("💰 Wallet %s Balance: %d\n", address, balance)

		// ---------------- EXIT ----------------
		case "exit":
			fmt.Println("👋 Shutting down ProCo Node...")
			return

		default:
			fmt.Println("Unknown command. Type 'help' for commands.")
		}
	}
}
