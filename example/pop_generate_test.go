package example

import (
	"manindexer/adapter/bitcoin"
	"manindexer/common"
	"testing"
)

func TestPopGenerate(t *testing.T) {
	pinid := "77aac2ae323748dee3b8b1ae6b7c33c1c4466f568c572ea488f584f041f0de4ei0"    // 64 char hash
	merkleRoot := "e56011a241cb196fc4efbeafef051ca901761ffb569a43146582f9133bfd41d2" // 64 char hash
	blockhash := "000000000000000004c2db0441a47fd3574992d508b8d9d866a789d371aa5060"  // real block hash

	// Generate a PoP and zero count
	pop, _ := common.GenPop(pinid, merkleRoot, blockhash)
	t.Logf("PoP: %s", pop)

	indexer := &bitcoin.Indexer{}
	popLv, popSummry := indexer.PopLevelCount(pop)
	t.Logf("PoP-Lv:%d, PoP-Summry:%s", popLv, popSummry)
}
