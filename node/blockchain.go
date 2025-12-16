package node

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// ---------------- BLOCK STRUCT ----------------
type Block struct {
	Index     int    `json:"Index"`
	Timestamp string `json:"Timestamp"`
	Data      string `json:"Data"`
	PrevHash  string `json:"PrevHash"`
	Hash      string `json:"Hash"`
}

// ---------------- BLOCKCHAIN STRUCT ----------------
type Blockchain struct {
	Blocks  []Block   `json:"Blocks"`
	Wallets []*Wallet `json:"Wallets"`
}

// ---------------- HASH FUNCTION ----------------
func CalculateHash(block Block) string {
	record := fmt.Sprintf("%d%s%s%s",
		block.Index,
		block.Timestamp,
		block.Data,
		block.PrevHash,
	)
	h := sha256.New()
	h.Write([]byte(record))
	return hex.EncodeToString(h.Sum(nil))
}

// ---------------- GENESIS BLOCK ----------------
func NewGenesisBlock() Block {
	block := Block{
		Index:     0,
		Timestamp: time.Now().Format(time.RFC3339),
		Data:      "Genesis Block",
		PrevHash:  "",
	}
	block.Hash = CalculateHash(block)
	return block
}

// ---------------- ADD BLOCK ----------------
func (bc *Blockchain) AddBlock(data string) {
	if len(bc.Blocks) == 0 {
		genesis := NewGenesisBlock()
		bc.Blocks = append(bc.Blocks, genesis)
		bc.Save("blocks.json")
		return
	}

	prevBlock := bc.Blocks[len(bc.Blocks)-1]
	newBlock := Block{
		Index:     prevBlock.Index + 1,
		Timestamp: time.Now().Format(time.RFC3339),
		Data:      data,
		PrevHash:  prevBlock.Hash,
	}
	newBlock.Hash = CalculateHash(newBlock)
	bc.Blocks = append(bc.Blocks, newBlock)
	bc.Save("blocks.json")
}

// ---------------- SAVE BLOCKCHAIN ----------------
func (bc *Blockchain) Save(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(bc)
}

// ---------------- LOAD BLOCKCHAIN ----------------
func LoadBlockchain(filename string) (*Blockchain, error) {
	file, err := os.Open(filename)
	if err != nil {
		genesis := NewGenesisBlock()
		bc := &Blockchain{Blocks: []Block{genesis}}
		bc.Save(filename)
		return bc, nil
	}
	defer file.Close()

	var bc Blockchain
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&bc)
	if err != nil {
		return nil, err
	}
	return &bc, nil
}

// ---------------- SHOW BLOCKS ----------------
func (bc *Blockchain) ShowChain() {
	fmt.Println("\n📦 Blockchain:")
	out, _ := json.MarshalIndent(bc.Blocks, "", "  ")
	fmt.Println(string(out))
}

// ---------------- VALIDATE BLOCKCHAIN ----------------
func (bc *Blockchain) ValidateChain() {
	fmt.Println("\n🔍 Validating Blockchain...")
	for i := 1; i < len(bc.Blocks); i++ {
		current := bc.Blocks[i]
		previous := bc.Blocks[i-1]

		if current.PrevHash != previous.Hash {
			fmt.Println("❌ Chain broken at block", current.Index)
			fmt.Println("Expected:", previous.Hash)
			fmt.Println("Got     :", current.PrevHash)
			return
		}

		validHash := CalculateHash(current)
		if current.Hash != validHash {
			fmt.Println("❌ Invalid hash at block", current.Index)
			return
		}
	}
	fmt.Println("✅ Blockchain is valid and secure.")
}
