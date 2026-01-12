# Proco Blockchain (Experimental)

Proco is an **experimental blockchain project written in Go**, built to explore how a simple Layer-1 blockchain works at a low level — including **P2P networking, block propagation, node synchronization, and basic consensus logic**.

This project is **learning-focused** and intended for developers who want to understand blockchain internals by reading and running real code.

⚠️ **This is not production-ready software.**

This repository includes a basic Go unit test (`node/node_test.go`) verifying that the blockchain adds blocks correctly.  

Run:

```bash
go test ./node
```


## 🎬 Live Local Demo (Running Node)
Below is a real screen recording of the **ProCo node running locally**.

![ProCo Node Demo](docs/demo.gif)




**What problem does this project experiment with?**

Most blockchain tutorials stop at theory or isolated components.
This project experiments with putting the pieces together:

How nodes discover and communicate with each other (P2P)

How blocks are created, broadcast, and synced

How a simple consensus flow can be structured

How a minimal blockchain node behaves when running continuously

The goal is clarity over complexity — understanding the flow rather than building a feature-heavy chain.
---
How to run a node (3 steps)
1️⃣ Prerequisites

Go installed (recommended Go 1.20+)

Git

Verify:

go version

2️⃣ Clone the repository
git clone https://github.com/ProcoNova/Proco-blockchain.git
cd Proco-blockchain

3️⃣ Run the node
go run main.go


You should see logs indicating:

Node startup

Block creation / syncing

Network activity (if peers are connected)

Multiple nodes can be run on different ports or machines to observe syncing behavior.

**Current limitations (important & honest)**

This project is intentionally minimal. Current limitations include:

❌ No production-grade security

❌ No economic or incentive model

❌ No advanced consensus (e.g., PoS, BFT)

❌ No transaction validation rules for real-world use

❌ No formal testing or audits

**Project status**

Actively developed as a learning & experimentation project

Focused on core blockchain mechanics

Open to feedback, suggestions, and code reviews.

**Who is this for?**

Developers learning Go

Engineers curious about blockchain internals

Anyone who prefers reading real code over whitepapers

📜 License

MIT License

🧠 Disclaimer

This project is for educational and research purposes only.
It does not represent financial advice or an investment product.
