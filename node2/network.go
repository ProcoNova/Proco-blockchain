// network.go
// Peer discovery, transaction gossip, mempool, networking improvements for ProCo
// This file contains helper types & functions only — it must NOT contain main().

package main

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

// --- CONFIG ---
const (
	NetworkListenPort     = "8001" // change per node if needed or make configurable
	PeerExchangeInterval  = 15 * time.Second
	ReconnectInterval     = 5 * time.Second
	MessageReadTimeout    = 30 * time.Second
)

// --- MESSAGE TYPES ---
const (
	MsgTypePeerList = "PEER_LIST"
	MsgTypePing     = "PING"
	MsgTypePong     = "PONG"
	MsgTypeTx       = "TX"
	MsgTypeBlock    = "BLOCK"
)

// Envelope for network messages
type NetMessage struct {
	Type string          `json:"type"`
	From string          `json:"from"` // address:port
	Body json.RawMessage `json:"body"`
}

// Peer information
type Peer struct {
	Addr      string    `json:"addr"`
	LastSeen  time.Time `json:"last_seen"`
	Connected bool      `json:"-"`
	conn      net.Conn  `json:"-"`
}

// Mempool: simple in-memory tx pool
type Mempool struct {
	mu sync.Mutex
	m  map[string]json.RawMessage // txid -> txjson
}

func NewMempool() *Mempool {
	return &Mempool{m: make(map[string]json.RawMessage)}
}

func (mp *Mempool) Add(txid string, raw json.RawMessage) bool {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	if _, ok := mp.m[txid]; ok {
		return false
	}
	mp.m[txid] = raw
	return true
}

func (mp *Mempool) Remove(txid string) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	delete(mp.m, txid)
}

func (mp *Mempool) List() []json.RawMessage {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	out := make([]json.RawMessage, 0, len(mp.m))
	for _, v := range mp.m {
		out = append(out, v)
	}
	return out
}

// --- Peer Manager ---
type PeerManager struct {
	mu    sync.Mutex
	peers map[string]*Peer // addr -> peer
}

func NewPeerManager() *PeerManager {
	return &PeerManager{peers: make(map[string]*Peer)}
}

func (pm *PeerManager) Add(addr string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if p, ok := pm.peers[addr]; ok {
		p.LastSeen = time.Now()
		return
	}
	pm.peers[addr] = &Peer{Addr: addr, LastSeen: time.Now()}
}

func (pm *PeerManager) Remove(addr string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if p, ok := pm.peers[addr]; ok {
		if p.conn != nil {
			p.conn.Close()
		}
		delete(pm.peers, addr)
	}
}

func (pm *PeerManager) List() []string {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	out := make([]string, 0, len(pm.peers))
	for addr := range pm.peers {
		out = append(out, addr)
	}
	return out
}

func (pm *PeerManager) UpdateConn(addr string, conn net.Conn) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if p, ok := pm.peers[addr]; ok {
		p.conn = conn
		p.Connected = true
		p.LastSeen = time.Now()
	} else {
		pm.peers[addr] = &Peer{Addr: addr, conn: conn, Connected: true, LastSeen: time.Now()}
	}
}

func (pm *PeerManager) MarkDisconnected(addr string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if p, ok := pm.peers[addr]; ok {
		p.Connected = false
		if p.conn != nil {
			p.conn.Close()
			p.conn = nil
		}
	}
}

// --- Network Node ---

type NetworkNode struct {
	listenAddr string
	pm         *PeerManager
	mempool    *Mempool
	selfAddr   string
	quit       chan struct{}
}

func NewNetworkNode(listenAddr string, bootstrapPeers []string) *NetworkNode {
	pm := NewPeerManager()
	for _, p := range bootstrapPeers {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		pm.Add(p)
	}
	return &NetworkNode{listenAddr: listenAddr, pm: pm, mempool: NewMempool(), selfAddr: listenAddr, quit: make(chan struct{})}
}

func (n *NetworkNode) Start() error {
	ln, err := net.Listen("tcp", n.listenAddr)
	if err != nil {
		return err
	}
	log.Printf("[net] Listening on %s\n", n.listenAddr)

	// accept loop
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				select {
				case <-n.quit:
					return
				default:
					log.Printf("[net] Accept error: %v\n", err)
					continue
				}
			}
			go n.handleConn(conn)
		}
	}()

	// start periodic peer exchange
	go n.peerExchangeLoop()

	// start auto-reconnect loop
	go n.autoConnectLoop()

	return nil
}

func (n *NetworkNode) Stop() {
	close(n.quit)
}

func (n *NetworkNode) handleConn(conn net.Conn) {
	remote := conn.RemoteAddr().String()
	log.Printf("[net] New connection from %s\n", remote)
	n.pm.UpdateConn(remote, conn)
	defer func() {
		n.pm.MarkDisconnected(remote)
		conn.Close()
		log.Printf("[net] Connection from %s closed\n", remote)
	}()

	r := bufio.NewReader(conn)
	for {
		conn.SetReadDeadline(time.Now().Add(MessageReadTimeout))
		line, err := r.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return
			}
			netErr, ok := err.(net.Error)
			if ok && netErr.Timeout() {
				// continue to wait, peer may be alive
				continue
			}
			log.Printf("[net] read error from %s: %v\n", remote, err)
			return
		}
		var msg NetMessage
		if err := json.Unmarshal(line, &msg); err != nil {
			log.Printf("[net] invalid msg from %s: %v\n", remote, err)
			continue
		}
		n.handleMessage(msg, conn)
	}
}

func (n *NetworkNode) handleMessage(msg NetMessage, conn net.Conn) {
	switch msg.Type {
	case MsgTypePing:
		// reply pong
		pong := NetMessage{Type: MsgTypePong, From: n.selfAddr, Body: nil}
		n.sendRaw(conn, pong)
	case MsgTypePong:
		// update peer last seen
		n.pm.Add(msg.From)
	case MsgTypePeerList:
		// merge peer list
		var peers []string
		if err := json.Unmarshal(msg.Body, &peers); err != nil {
			log.Printf("[net] invalid peerlist from %s: %v\n", msg.From, err)
			return
		}
		for _, p := range peers {
			if p == n.selfAddr {
				continue
			}
			n.pm.Add(p)
		}
	case MsgTypeTx:
		// receive transaction gossip
		var tx json.RawMessage = msg.Body
		// assume tx contains a field txid - but to keep generic, compute a lightweight id
		txid := simpleTxID(tx)
		if n.mempool.Add(txid, tx) {
			log.Printf("[mempool] Added tx %s from %s\n", txid, msg.From)
			// forward to other peers
			n.BroadcastMessageExcept(msg, msg.From)
		} else {
			// already had it
		}
	case MsgTypeBlock:
		// integrate block handling with main node (stub)
		log.Printf("[net] Received BLOCK from %s - integrate with chain logic\n", msg.From)
		// TODO: validate and append block, remove included txs from mempool
	default:
		log.Printf("[net] Unknown message type %s from %s\n", msg.Type, msg.From)
	}
}

func (n *NetworkNode) sendRaw(conn net.Conn, msg NetMessage) error {
	b, _ := json.Marshal(msg)
	b = append(b, '\n')
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err := conn.Write(b)
	return err
}

func (n *NetworkNode) sendToAddr(addr string, msg NetMessage) error {
	n.pm.mu.Lock()
	p, ok := n.pm.peers[addr]
	n.pm.mu.Unlock()
	if ok && p.conn != nil {
		return n.sendRaw(p.conn, msg)
	}
	// try to dial
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return err
	}
	n.pm.UpdateConn(addr, conn)
	go n.handleConn(conn)
	return n.sendRaw(conn, msg)
}

// Broadcast to all known peers
func (n *NetworkNode) BroadcastMessage(msg NetMessage) {
	peers := n.pm.List()
	for _, addr := range peers {
		if addr == n.selfAddr {
			continue
		}
		_ = n.sendToAddr(addr, msg)
	}
}

// Broadcast to all except excludedAddr
func (n *NetworkNode) BroadcastMessageExcept(msg NetMessage, excludedAddr string) {
	peers := n.pm.List()
	for _, addr := range peers {
		if addr == n.selfAddr || addr == excludedAddr {
			continue
		}
		_ = n.sendToAddr(addr, msg)
	}
}

// peerExchangeLoop periodically sends our peer list to connected peers
func (n *NetworkNode) peerExchangeLoop() {
	ticker := time.NewTicker(PeerExchangeInterval)
	defer ticker.Stop()
	for {
		select {
		case <-n.quit:
			return
		case <-ticker.C:
			peers := n.pm.List()
			body, _ := json.Marshal(peers)
			msg := NetMessage{Type: MsgTypePeerList, From: n.selfAddr, Body: body}
			n.BroadcastMessage(msg)
		}
	}
}

// autoConnectLoop periodically tries to connect to peers without connections
func (n *NetworkNode) autoConnectLoop() {
	ticker := time.NewTicker(ReconnectInterval)
	defer ticker.Stop()
	for {
		select {
		case <-n.quit:
			return
		case <-ticker.C:
			peers := n.pm.List()
			for _, addr := range peers {
				p := n.pm.peers[addr]
				if p == nil || p.conn == nil {
					if addr == n.selfAddr {
						continue
					}
					conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
					if err != nil {
						continue
					}
					n.pm.UpdateConn(addr, conn)
					go n.handleConn(conn)
				}
			}
		}
	}
}

// Utility for tx id (very simple - in production use proper hashing)
func simpleTxID(raw json.RawMessage) string {
	// take first 20 chars of raw json as id (not secure but deterministic)
	s := string(raw)
	if len(s) > 20 {
		return s[:20]
	}
	return s
}

// Public API: BroadcastTransaction - called by node code when creating local tx
func (n *NetworkNode) BroadcastTransaction(tx json.RawMessage) {
	msg := NetMessage{Type: MsgTypeTx, From: n.selfAddr, Body: tx}
	// add to mempool locally too
	txid := simpleTxID(tx)
	n.mempool.Add(txid, tx)
	// broadcast
	n.BroadcastMessage(msg)
}

// small helpers used by node.go CLI or integration code
func splitComma(s string) []string {
	out := []string{}
	if s == "" {
		return out
	}
	parts := strings.Split(s, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func trim(s string) string {
	return strings.TrimSpace(s)
}
