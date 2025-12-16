package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"
)

// ---------------- Blockchain structs ----------------
type Block struct {
	Index     int    `json:"index"`
	Timestamp string `json:"timestamp"`
	Data      string `json:"data"`
	PrevHash  string `json:"prev_hash"`
	Hash      string `json:"hash"`
}

type Blockchain struct {
	Blocks []Block
	mu     sync.Mutex
}

func LoadBlockchain(filename string) (*Blockchain, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return &Blockchain{}, nil
	}
	var blocks []Block
	err = json.Unmarshal(file, &blocks)
	if err != nil {
		return nil, err
	}
	return &Blockchain{Blocks: blocks}, nil
}

func (bc *Blockchain) SaveToFile(filename string) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	data, err := json.MarshalIndent(bc.Blocks, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0644)
}

func (bc *Blockchain) AddBlock(block Block) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	for _, b := range bc.Blocks {
		if b.Hash == block.Hash {
			return
		}
	}
	bc.Blocks = append(bc.Blocks, block)
}

func (bc *Blockchain) GetBlocks() []Block {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	return bc.Blocks
}

// ---------------- Wallet ----------------
type Wallet struct {
	Node         string   `json:"node"`
	Address      string   `json:"address"`
	PrivateKey   string   `json:"private_key"`
	PublicKey    string   `json:"public_key"`
	Balance      int      `json:"balance"`
	Transactions []string `json:"transactions"`
}

// Safe wallet loader and creator
func LoadOrCreateWallet(filename string) *Wallet {
	wallet := &Wallet{} // always initialize

	data, err := os.ReadFile(filename)
	if err != nil || len(data) == 0 {
		fmt.Println("Wallet file missing or empty. Creating new wallet...")
		wallet = &Wallet{
			Node:         "Node2",
			Address:      "PROCO2-" + fmt.Sprintf("%x", time.Now().UnixNano()),
			PrivateKey:   fmt.Sprintf("%x", time.Now().UnixNano()*7),
			PublicKey:    fmt.Sprintf("%x", time.Now().UnixNano()*13),
			Balance:      1000,
			Transactions: []string{},
		}
	} else {
		err = json.Unmarshal(data, wallet)
		if err != nil || wallet.Address == "" {
			fmt.Println("Wallet file corrupted. Creating new wallet...")
			wallet = &Wallet{
				Node:         "Node2",
				Address:      "PROCO2-" + fmt.Sprintf("%x", time.Now().UnixNano()),
				PrivateKey:   fmt.Sprintf("%x", time.Now().UnixNano()*7),
				PublicKey:    fmt.Sprintf("%x", time.Now().UnixNano()*13),
				Balance:      1000,
				Transactions: []string{},
			}
		}
	}

	// Save wallet immediately
	walletData, _ := json.MarshalIndent(wallet, "", "  ")
	_ = ioutil.WriteFile(filename, walletData, 0644)
	return wallet
}

// ---------------- Sync Node1 ----------------
func SyncWithNode1(bc *Blockchain) {
	resp, err := http.Get("http://localhost:8080/getBlockchain")
	if err != nil {
		fmt.Println("❌ Error fetching Node1 blockchain:", err)
		return
	}
	defer resp.Body.Close()

	var node1Blocks []Block
	err = json.NewDecoder(resp.Body).Decode(&node1Blocks)
	if err != nil {
		fmt.Println("❌ Error decoding Node1 blockchain:", err)
		return
	}

	for _, b := range node1Blocks {
		bc.AddBlock(b)
	}
	bc.SaveToFile("blocks.json")
	fmt.Println("✅ Synced with Node1")
}

// ---------------- Main ----------------
func main() {
	bc, _ := LoadBlockchain("blocks.json")
	wallet := LoadOrCreateWallet("wallet.json")

	if len(bc.Blocks) == 0 {
		genesis := Block{
			Index:     0,
			Timestamp: time.Now().Format(time.RFC3339),
			Data:      "Genesis Block from Node2",
			PrevHash:  "",
			Hash:      fmt.Sprintf("%x", time.Now().UnixNano()),
		}
		bc.AddBlock(genesis)
		bc.SaveToFile("blocks.json")
		fmt.Println("Genesis block created")
	}

	fmt.Printf("✅ Node2 started\n")
	fmt.Printf("✅ Wallet Address: %s\n", wallet.Address)
	fmt.Printf("✅ Listening on port 8081\n")

	SyncWithNode1(bc)

	http.HandleFunc("/getBlockchain", func(w http.ResponseWriter, r *http.Request) {
		data, _ := json.MarshalIndent(bc.GetBlocks(), "", "  ")
		w.Write(data)
	})

	http.HandleFunc("/addBlock", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var block Block
		err := json.NewDecoder(r.Body).Decode(&block)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		bc.AddBlock(block)
		bc.SaveToFile("blocks.json")
		w.WriteHeader(http.StatusCreated)
		fmt.Println("New block added:", block.Index)
	})

	http.ListenAndServe(":8081", nil)
}
