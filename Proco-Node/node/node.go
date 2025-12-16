package node

type Node struct {
    Config     *Config
    Blockchain *Blockchain
}

// Create a new node
func NewNode(configPath string) (*Node, error) {
    cfg, err := LoadConfig(configPath)
    if err != nil {
        return nil, err
    }

    node := &Node{
        Config:     cfg,
        Blockchain: NewBlockchain(),
    }
    return node, nil
}

// Start the node (placeholder for now)
func (n *Node) Start() error {
    // In real network, start listening and processing blocks
    return nil
}
