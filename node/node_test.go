package node

import "testing"

func TestAddBlock(t *testing.T) {
	bc := NewBlockchain()

	if len(bc.Blocks) != 1 {
		t.Fatalf("expected genesis block, got %d blocks", len(bc.Blocks))
	}

	bc.AddBlock("test transaction")

	if len(bc.Blocks) != 2 {
		t.Fatalf("expected 2 blocks after adding, got %d", len(bc.Blocks))
	}

	if bc.Blocks[1].PrevHash != bc.Blocks[0].Hash {
		t.Fatal("block hash linkage is broken")
	}
}
