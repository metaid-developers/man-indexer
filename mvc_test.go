package main

import (
	"fmt"
	"manindexer/adapter/microvisionchain"
	"manindexer/common"
	"manindexer/man"
	"testing"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
)

func TestMvcCatchPinsByTx(t *testing.T) {
	man.InitAdapter("mvc", "mongo", "1", "1")
	txResult, err := man.ChainAdapter["mvc"].GetTransaction("4b774a6b1bdba67d9c28f44ecad53c591643ff7d4dac6161a74e622a190c8e58")
	fmt.Println(err)
	tx := txResult.(*btcutil.Tx)
	fmt.Println(tx.Hash().String())
	index := microvisionchain.Indexer{
		ChainParams: &chaincfg.TestNet3Params,
		PopCutNum:   common.Config.Mvc.PopCutNum,
		DbAdapter:   &man.DbAdapter,
	}
	hash := "4b774a6b1bdba67d9c28f44ecad53c591643ff7d4dac6161a74e622a190c8e58"
	index.CatchPinsByTx(tx.MsgTx(), 91722, 0, hash, "", 0)
}
func TestMvcGetSaveData(t *testing.T) {
	man.InitAdapter("mvc", "mongo", "1", "1")
	pinList, _, _, _, _, _, _, _, err := man.GetSaveData("mvc", 91722)
	fmt.Println(err, len(pinList))
}
