package adapter

import (
	"manindexer/mrc20"
	"manindexer/pin"
)

type Indexer interface {
	CatchPins(blockHeight int64) (pinInscriptions []*pin.PinInscription, txInList []string)
	CatchTransfer(idMap map[string]struct{}) (trasferMap map[string]*pin.PinTransferInfo)
	GetAddress(pkScript []byte) (address string)
	ZmqRun(chanMsg chan []*pin.PinInscription)
	GetBlockTxHash(blockHeight int64) (txhashList []string)
	ZmqHashblock()
	CatchNativeMrc20Transfer(blockHeight int64, utxoList []*mrc20.Mrc20Utxo, mrc20TransferPinTx map[string]struct{}) (savelist []*mrc20.Mrc20Utxo)
}
