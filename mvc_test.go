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
	txid := "c90955c83ac07fbf4351ede0381becd4e47922e25c4f9f3dc6c11339a1ed360f"
	txResult, err := man.ChainAdapter["mvc"].GetTransaction(txid)
	fmt.Println(err)
	tx := txResult.(*btcutil.Tx)
	fmt.Println(tx.Hash().String())
	index := microvisionchain.Indexer{
		ChainParams: &chaincfg.TestNet3Params,
		PopCutNum:   common.Config.Mvc.PopCutNum,
		DbAdapter:   &man.DbAdapter,
	}
	hash := txid
	list := index.CatchPinsByTx(tx.MsgTx(), 91722, 0, hash, "", 0)
	fmt.Println(list)
}
func TestCatchMvcData(t *testing.T) {
	common.InitConfig()
	man.InitAdapter("mvc", "mongo", "1", "1")
	//from := 2870989
	//to := 2870990
	// for i := from; i <= to; i++ {
	// 	man.DoIndexerRun("btc", int64(i))
	// }
	man.DoIndexerRun("mvc", int64(101608))

}
func TestMvcGetSaveData(t *testing.T) {
	man.InitAdapter("mvc", "mongo", "1", "1")
	pinList, _, _, _, _, _, _, _, _, _, err := man.GetSaveData("mvc", 91722)
	fmt.Println(err, len(pinList))
}
func TestGetBestHeight(t *testing.T) {
	common.InitConfig()
	man.InitAdapter("mvc", "mongo", "1", "1")
	bestHeight := man.ChainAdapter["mvc"].GetBestHeight()
	fmt.Println(bestHeight)
}
