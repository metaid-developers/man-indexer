package microvisionchain

import (
	"fmt"
	"manindexer/common"
	"manindexer/pin"
	"time"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

var (
	client *rpcclient.Client
)

type MicroVisionChain struct {
	IsTest bool
}

func init() {
	mvc := common.Config.Mvc
	rpcConfig := &rpcclient.ConnConfig{
		Host:                 mvc.RpcHost,
		User:                 mvc.RpcUser,
		Pass:                 mvc.RpcPass,
		HTTPPostMode:         mvc.RpcHTTPPostMode, //only supports HTTP POST mode
		DisableTLS:           mvc.RpcDisableTLS,   //core does not provide TLS by default
		DisableAutoReconnect: true,
		DisableConnectOnNew:  true,
	}
	var err error
	client, err = rpcclient.New(rpcConfig, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("mvc rpc  connect")
}
func (chain *MicroVisionChain) GetBlock(blockHeight int64) (block interface{}, err error) {
	blockhash, err := client.GetBlockHash(blockHeight)
	if err != nil {
		return
	}
	block, err = client.GetBlock(blockhash)
	return
}
func (chain *MicroVisionChain) GetBlockTime(blockHeight int64) (timestamp int64, err error) {
	block, err := chain.GetBlock(blockHeight)
	if err != nil {
		return
	}
	b := block.(*wire.MsgBlock)
	timestamp = b.Header.Timestamp.Unix()
	return
}
func (chain *MicroVisionChain) GetBlockByHash(hash string) (block *btcjson.GetBlockVerboseResult, err error) {
	blockhash, err := chainhash.NewHashFromStr(hash)
	if err != nil {
		return
	}
	block, err = client.GetBlockVerbose(blockhash)

	return
}
func (chain *MicroVisionChain) GetTransaction(txId string) (tx interface{}, err error) {
	txHash, _ := chainhash.NewHashFromStr(txId)
	return client.GetRawTransaction(txHash)
}
func GetValueByTx(txId string, txIdx int) (value int64, err error) {
	txHash, _ := chainhash.NewHashFromStr(txId)
	tx, err := client.GetRawTransaction(txHash)
	if err != nil {
		return
	}
	value = tx.MsgTx().TxOut[txIdx].Value
	return
}
func (chain *MicroVisionChain) GetInitialHeight() (height int64) {
	return common.Config.Mvc.InitialHeight
}
func (chain *MicroVisionChain) GetBestHeight() (height int64) {
	info, err := client.GetBlockChainInfo()
	if err != nil {
		return
	}
	height = int64(info.Blocks)
	//fmt.Println(height)
	return
}
func (chain *MicroVisionChain) GetBlockMsg(height int64) (blockMsg *pin.BlockMsg) {
	blockhash, err := client.GetBlockHash(height)
	if err != nil {
		return
	}
	block, err := client.GetBlockVerbose(blockhash)
	if err != nil {
		return
	}
	blockMsg = &pin.BlockMsg{}
	blockMsg.BlockHash = block.Hash
	blockMsg.Target = block.MerkleRoot
	blockMsg.Weight = int64(block.Weight)
	blockMsg.Timestamp = time.Unix(block.Time, 0).Format("2006-01-02 15:04:05")
	blockMsg.Size = int64(block.Size)
	blockMsg.Transaction = block.Tx
	blockMsg.TransactionNum = len(block.Tx)
	return
}
func (chain *MicroVisionChain) GetCreatorAddress(txHashStr string, idx uint32, netParams *chaincfg.Params) (address string) {
	txHash, err := chainhash.NewHashFromStr(txHashStr)
	if err != nil {
		return "errorAddr"
	}
	tx, err := client.GetRawTransaction(txHash)
	if err != nil {
		return "errorAddr"
	}
	_, addresses, _, _ := txscript.ExtractPkScriptAddrs(tx.MsgTx().TxOut[idx].PkScript, netParams)
	if len(addresses) > 0 {
		address = addresses[0].String()
	} else {
		address = "errorAddr"
	}
	return
}
