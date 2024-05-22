package bitcoin

import (
	"manindexer/common"
	"manindexer/pin"
	"time"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
)

var (
	client *rpcclient.Client
)

type BitcoinChain struct {
	IsTest bool
}

func init() {
	btc := common.Config.Btc
	rpcConfig := &rpcclient.ConnConfig{
		Host:                 btc.RpcHost,
		User:                 btc.RpcUser,
		Pass:                 btc.RpcPass,
		HTTPPostMode:         btc.RpcHTTPPostMode, // Bitcoin core only supports HTTP POST mode
		DisableTLS:           btc.RpcDisableTLS,   // Bitcoin core does not provide TLS by default
		DisableAutoReconnect: true,
		DisableConnectOnNew:  true,
	}
	var err error
	client, err = rpcclient.New(rpcConfig, nil)
	if err != nil {
		panic(err)
	}
}
func (chain *BitcoinChain) GetBlock(blockHeight int64) (block interface{}, err error) {
	blockhash, err := client.GetBlockHash(blockHeight)
	if err != nil {
		return
	}
	block, err = client.GetBlock(blockhash)
	return
}
func (chain *BitcoinChain) GetBlockByHash(hash string) (block *btcjson.GetBlockVerboseResult, err error) {
	blockhash, err := chainhash.NewHashFromStr(hash)
	if err != nil {
		return
	}
	block, err = client.GetBlockVerbose(blockhash)

	return
}
func (chain *BitcoinChain) GetTransaction(txId string) (tx interface{}, err error) {
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
func (chain *BitcoinChain) GetInitialHeight() (height int64) {
	return common.Config.Btc.InitialHeight
}
func (chain *BitcoinChain) GetBestHeight() (height int64) {
	info, err := client.GetBlockChainInfo()
	if err != nil {
		return
	}
	height = int64(info.Blocks)
	return
}
func (chain *BitcoinChain) GetBlockMsg(height int64) (blockMsg *pin.BlockMsg) {
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
