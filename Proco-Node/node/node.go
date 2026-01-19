package node

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

// --------------------
// BLOCK & BLOCKCHAIN
// --------------------

type Block struct {
	Index    int
	Hash     string
	PrevHash string
	Data     string
	Tx       []Transaction
}

type Blockchain struct {
	Blocks []*Block
	mu     sync.Mutex
}

func NewBlockchain() *Blockchain {
	genesis := &Block{
		Index:    0,
		Hash:     "GENESIS_HASH",
		PrevHash: "",
		Data:     "Genesis Block",
		Tx:       []Transaction{},
	}
	return &Blockchain{
		Blocks: []*Block{genesis},
	}
}

func (bc *Blockchain) AddBlock(data string, txs []Transaction) *Block {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	last := bc.Blocks[len(bc.Blocks)-1]

	newBlock := &Block{
		Index:    last.Index + 1,
		PrevHash: last.Hash,
		Hash:     fmt.Sprintf("%x", last.Index+1),
		Data:     data,
		Tx:       txs,
	}

	bc.Blocks = append(bc.Blocks, newBlock)
	return newBlock
}

// --------------------
// TRANSACTION
// --------------------

type Transaction struct {
	Sender    string
	Recipient string
	Amount    int
	Timestamp time.Time
}

// --------------------
// NODE
// --------------------

type Node struct {
	Config     *Config
	Blockchain *Blockchain
	Mempool    []Transaction
	mu         sync.Mutex
	ListenPort string
	Peers      []string
}

type Config struct {
	ChainID string
}

func NewNode(_ string) (*Node, error) {
	return &Node{
		Config: &Config{
			ChainID: "proco-testnet",
		},
		Blockchain: NewBlockchain(),
		Mempool:    []Transaction{},
		Peers:      []string{},
	}, nil
}

// --------------------
// TRANSACTIONS & MINING
// --------------------

func (n *Node) AddTransaction(sender, recipient string, amount int) {
	n.mu.Lock()
	defer n.mu.Unlock()

	tx := Transaction{
		Sender:    sender,
		Recipient: recipient,
		Amount:    amount,
		Timestamp: time.Now(),
	}

	n.Mempool = append(n.Mempool, tx)
	fmt.Printf("Transaction added: %+v\n", tx)
}

func (n *Node) MinePendingTxs() *Block {
	n.mu.Lock()
	if len(n.Mempool) == 0 {
		n.mu.Unlock()
		fmt.Println("No transactions to mine")
		return nil
	}

	txs := make([]Transaction, len(n.Mempool))
	copy(txs, n.Mempool)
	n.Mempool = []Transaction{}
	n.mu.Unlock()

	block := n.Blockchain.AddBlock(
		fmt.Sprintf("Block with %d TXs", len(txs)),
		txs,
	)

	fmt.Printf("⛏️  Mined new block #%d with %d transactions\n", block.Index, len(txs))

	// 🔥 THIS WAS MISSING
	n.SendBlock(block)

	return block
}

// --------------------
// NETWORKING
// --------------------

func (n *Node) Listen() {
	ln, err := net.Listen("tcp", ":"+n.ListenPort)
	if err != nil {
		fmt.Println("Listen error:", err)
		return
	}
	defer ln.Close()

	fmt.Println("🌐 Node listening on port", n.ListenPort)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Accept error:", err)
			continue
		}

		fmt.Println("📥 Incoming connection from", conn.RemoteAddr())
		go n.handleConnection(conn)
	}
}

func (n *Node) handleConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 4096)
	nBytes, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Read error:", err)
		return
	}

	var block Block
	if err := json.Unmarshal(buf[:nBytes], &block); err != nil {
		fmt.Println("JSON decode error:", err)
		return
	}

	n.Blockchain.mu.Lock()
	defer n.Blockchain.mu.Unlock()

	last := n.Blockchain.Blocks[len(n.Blockchain.Blocks)-1]

	if block.Index == last.Index+1 {
		n.Blockchain.Blocks = append(n.Blockchain.Blocks, &block)
		fmt.Printf("✅ Received and accepted block #%d\n", block.Index)
	} else {
		fmt.Printf(
			"❌ Block rejected. Got #%d, expected #%d\n",
			block.Index,
			last.Index+1,
		)
	}
}

func (n *Node) SendBlock(block *Block) {
	data, _ := json.Marshal(block)

	for _, peer := range n.Peers {
		conn, err := net.Dial("tcp", peer)
		if err != nil {
			fmt.Println("SendBlock error:", err)
			continue
		}

		conn.Write(data)
		conn.Close()
		fmt.Println("📤 Block sent to", peer)
	}
}
