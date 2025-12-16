package main

import (
    "bytes"
    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/rand"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "log"
    "math/big"
    "net/http"
    "os"
    "sync"
    "time"
)

/* =======================
   DATA STRUCTURES
======================= */

type Signature struct {
    R string `json:"r"`
    S string `json:"s"`
}

type Transaction struct {
    From      string    `json:"from"`
    To        string    `json:"to"`
    Amount    int       `json:"amount"`
    Signature Signature `json:"signature"`
}

type Block struct {
    Index        int           `json:"index"`
    PreviousHash string        `json:"previous_hash"`
    Timestamp    int64         `json:"timestamp"`
    Data         []Transaction `json:"data"`
    Hash         string        `json:"hash"`
}

var Blockchain []Block
var BlockchainMutex sync.Mutex

var Mempool []Transaction
var MempoolMutex sync.Mutex

var Peers []string

var PrivateKey *ecdsa.PrivateKey
var PublicKey ecdsa.PublicKey

/* =======================
   KEYS
======================= */

func GenerateKeyPair() {
    key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    PrivateKey = key
    PublicKey = key.PublicKey
    fmt.Println("🔑 Keys generated")
}

/* =======================
   TRANSACTIONS
======================= */

func SignTransaction(tx Transaction) Signature {
    data := tx.From + tx.To + fmt.Sprint(tx.Amount)
    hash := sha256.Sum256([]byte(data))
    r, s, _ := ecdsa.Sign(rand.Reader, PrivateKey, hash[:])
    return Signature{
        R: hex.EncodeToString(r.Bytes()),
        S: hex.EncodeToString(s.Bytes()),
    }
}

func VerifyTransaction(tx Transaction) bool {
    data := tx.From + tx.To + fmt.Sprint(tx.Amount)
    hash := sha256.Sum256([]byte(data))

    rBytes, _ := hex.DecodeString(tx.Signature.R)
    sBytes, _ := hex.DecodeString(tx.Signature.S)

    var r, s big.Int
    r.SetBytes(rBytes)
    s.SetBytes(sBytes)

    return ecdsa.Verify(&PublicKey, hash[:], &r, &s)
}

/* =======================
   BLOCKS
======================= */

func CalculateHash(block Block) string {
    record := fmt.Sprintf("%d%s%d", block.Index, block.PreviousHash, block.Timestamp)
    hash := sha256.Sum256([]byte(record))
    return hex.EncodeToString(hash[:])
}

func GenerateBlock(old Block, txs []Transaction) Block {
    block := Block{
        Index:        old.Index + 1,
        PreviousHash: old.Hash,
        Timestamp:    time.Now().Unix(),
        Data:         txs,
    }
    block.Hash = CalculateHash(block)
    return block
}

/* =======================
   MEMPOOL
======================= */

func AddTransaction(tx Transaction) {
    if !VerifyTransaction(tx) {
        fmt.Println("⚠️ Invalid transaction rejected")
        return
    }

    MempoolMutex.Lock()
    defer MempoolMutex.Unlock()
    Mempool = append(Mempool, tx)
}

/* =======================
   NETWORKING
======================= */

func broadcastTx(tx Transaction) {
    for _, peer := range Peers {
        data, _ := json.Marshal(tx)
        go http.Post(peer+"/receivetx", "application/json", bytes.NewBuffer(data))
    }
}

func broadcastBlock(block Block) {
    for _, peer := range Peers {
        data, _ := json.Marshal(block)
        go http.Post(peer+"/receiveblock", "application/json", bytes.NewBuffer(data))
    }
}

/* =======================
   API
======================= */

func createTx(w http.ResponseWriter, r *http.Request) {
    tx := Transaction{
        From:   os.Getenv("NODE_ADDRESS"),
        To:     r.URL.Query().Get("to"),
        Amount: 10,
    }
    tx.Signature = SignTransaction(tx)

    AddTransaction(tx)
    broadcastTx(tx)
    json.NewEncoder(w).Encode(tx)
}

func createBlock(w http.ResponseWriter, r *http.Request) {
    MempoolMutex.Lock()
    txs := Mempool
    Mempool = nil
    MempoolMutex.Unlock()

    BlockchainMutex.Lock()
    block := GenerateBlock(Blockchain[len(Blockchain)-1], txs)
    Blockchain = append(Blockchain, block)
    BlockchainMutex.Unlock()

    broadcastBlock(block)
    json.NewEncoder(w).Encode(block)
}

func receiveBlock(w http.ResponseWriter, r *http.Request) {
    var block Block
    json.NewDecoder(r.Body).Decode(&block)

    BlockchainMutex.Lock()
    defer BlockchainMutex.Unlock()

    last := Blockchain[len(Blockchain)-1]
    if block.Index == last.Index+1 && block.PreviousHash == last.Hash {
        Blockchain = append(Blockchain, block)
        fmt.Println("🧱 Block added from peer:", block.Index)
    }
}

func receiveTx(w http.ResponseWriter, r *http.Request) {
    var tx Transaction
    json.NewDecoder(r.Body).Decode(&tx)
    AddTransaction(tx)
}

func chain(w http.ResponseWriter, r *http.Request) {
    BlockchainMutex.Lock()
    defer BlockchainMutex.Unlock()
    json.NewEncoder(w).Encode(Blockchain)
}

/* =======================
   MAIN
======================= */

func main() {
    GenerateKeyPair()

    // ✅ FIXED GENESIS — SAME ON ALL NODES
    genesis := Block{
        Index:        0,
        PreviousHash: "0",
        Timestamp:    0,
        Data:         []Transaction{},
        Hash:         "GENESIS",
    }
    Blockchain = append(Blockchain, genesis)

    http.HandleFunc("/createtx", createTx)
    http.HandleFunc("/createblock", createBlock)
    http.HandleFunc("/receiveblock", receiveBlock)
    http.HandleFunc("/receivetx", receiveTx)
    http.HandleFunc("/chain", chain)

    port := os.Getenv("NODE_PORT")
    fmt.Println("🚀 Node running on port", port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}
