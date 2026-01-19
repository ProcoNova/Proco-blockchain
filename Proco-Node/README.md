ProCo Blockchain Project – Progress Update

Date: 19-Jan-2026
Author: Prosenjit (Captain) & ChatGPT (Tony)

✅ Recent Refactor & Achievements

Node Refactor

Merged blockchain.go and config.go into a single node.go.

Node logic now includes:

Blockchain structure & blocks

Transaction pool (mempool)

Mining logic

Networking (send/receive blocks between peers)

Multi-Node Setup

Created independent command folders for nodes:

cmd/node1/main.go

cmd/node2/main.go

Each node runs independently and communicates via TCP.

Node1 listens on port 3001; Node2 listens on port 3002.

Example successful test:

Node1 mined a block → Node2 received and accepted it.

Old Files Cleaned

Removed old common main.go in cmd/proco-node.

Removed blockchain.go and config.go from node/.

Reduces confusion and centralizes node logic in node.go.

🗂 Current Folder Structure
proco-node
│
├─ go.mod
├─ node
│   └─ node.go          # All blockchain & node logic
├─ configs
│   └─ genesis.json
└─ cmd
    ├─ node1
    │   └─ main.go      # Node1 runnable program
    └─ node2
        └─ main.go      # Node2 runnable program

🧪 Testing & Results

Node1 ran and mined blocks successfully:

Transaction added: {Sender:NODE1 Recipient:NODE2 Amount:10 ...}
Mined new block #1 with 1 transactions
Block sent to 127.0.0.1:3002


Node2 ran simultaneously:

Incoming connection from 127.0.0.1
✅ Received and accepted block #1


Observations:

Both nodes run independently.

Blocks propagate correctly between nodes.

Network messaging uses TCP sockets.

📝 Next Steps

Add real transaction input from user or CLI.

Improve block validation and consensus logic.

Add logging & metrics to monitor nodes.

Explore peer discovery for dynamic networks.!\[ProCo Node Demo](docs/demo.gif)

