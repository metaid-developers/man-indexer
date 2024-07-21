package main

import (
	"fmt"
	"manindexer/cmd/cli"
	"manindexer/common"
	"manindexer/man"
	"testing"
)

func TestNewAddress(t *testing.T) {
	common.InitConfig()
	man.InitAdapter(common.Chain, common.Db, common.TestNet, common.Server)
	//cli.InitBtcRpc("")
	//r, err := cli.CreateWallet("wang01")
	//fmt.Println(err, r)
	cli.InitBtcRpc("/wallet/wang02")
	// s, err := cli.GetNewAddress("wang01")
	// fmt.Println(err, s)
	s, err := cli.DumpPrivKeyHex("tb1q5ycztew7frsskg38rhk6c4h0j8vsswukjr993v")
	fmt.Println(err, s)
	//cli.CreateHdWallet("wang02", "682372604e2935f25aa7377bfc6c4a76a0f09eef2be5cc000834d6a2d025e0a6")
}
func TestGetUtxo(t *testing.T) {
	common.InitConfig()
	cli.InitBtcRpc("/wallet/wang02")
	// list, err := cli.GetBtcUtxo()
	// fmt.Println(err)
	// fmt.Println(list)
	// amt, err := cli.GetMempool("wang02")
	// fmt.Println(amt.ToBTC())
	addressList := []string{"tb1q5ycztew7frsskg38rhk6c4h0j8vsswukjr993v"}
	list, err := cli.GetUtxo(addressList)
	fmt.Println(err)
	for _, x := range list {
		fmt.Printf("%+v\n", x)
	}
}
func TestDumpPrivKeyHex(t *testing.T) {
	common.InitConfig()
	man.InitAdapter(common.Chain, common.Db, common.TestNet, common.Server)
	cli.InitBtcRpc("/wallet/" + cli.WALLETNAME)
	s, err := cli.DumpPrivKeyHex("tb1qshvx2gcfrp5hfxx4jn9ua6lrxe7hj9xqm57uyn")
	fmt.Println(err, s)

}
