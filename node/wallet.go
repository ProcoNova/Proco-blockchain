package node

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
)

// ---------------- WALLET STRUCT ----------------
type Wallet struct {
	Address string `json:"Address"`
	Balance int    `json:"Balance"`
}

// ---------------- CREATE NEW WALLET ----------------
func NewWallet(initialBalance int) *Wallet {
	addr := generateAddress()
	return &Wallet{
		Address: addr,
		Balance: initialBalance,
	}
}

// ---------------- GENERATE RANDOM ADDRESS ----------------
func generateAddress() string {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}

// ---------------- FIND WALLET ----------------
func FindWallet(bc *Blockchain, address string) *Wallet {
	for _, w := range bc.Wallets {
		if w.Address == address {
			return w
		}
	}
	return nil
}

// ---------------- SEND COINS ----------------
func SendCoins(bc *Blockchain, fromAddr, toAddr string, amount int) bool {
	fromWallet := FindWallet(bc, fromAddr)
	toWallet := FindWallet(bc, toAddr)

	if fromWallet == nil || toWallet == nil {
		fmt.Println("❌ Wallet not found")
		return false
	}

	if fromWallet.Balance < amount {
		fmt.Println("❌ Insufficient balance")
		return false
	}

	fromWallet.Balance -= amount
	toWallet.Balance += amount

	// --- Automatic block creation ---
	txData := fmt.Sprintf("Sent %d ProCo from %s to %s", amount, fromAddr, toAddr)
	bc.AddBlock(txData)
	fmt.Println("✅ Transaction successful and saved to blockchain.")

	// Save wallets
	err := bc.SaveWallets("wallets.json")
	if err != nil {
		fmt.Println("Error saving wallets:", err)
	}

	return true
}

// ---------------- SAVE WALLETS ----------------
func (bc *Blockchain) SaveWallets(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(bc.Wallets)
}

// ---------------- LOAD WALLETS ----------------
func (bc *Blockchain) LoadWallets(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		bc.Wallets = []*Wallet{}
		return nil
	}
	defer file.Close()

	var wallets []*Wallet
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&wallets)
	if err != nil {
		return err
	}
	bc.Wallets = wallets
	return nil
}

// ---------------- GET BALANCE ----------------
func GetBalance(bc *Blockchain, address string) int {
	for _, w := range bc.Wallets {
		if w.Address == address {
			return w.Balance
		}
	}
	return 0
}
