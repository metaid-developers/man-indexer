package adapter

import "manindexer/pin"

type Indexer interface {
	CatchPins(blockHeight int64) (pinInscriptions []*pin.PinInscription, txInList []string)
	CatchTransfer(idMap map[string]struct{}) (trasferMap map[string]*pin.PinTransferInfo)
	GetAddress(pkScript []byte) (address string)
	ZmqRun(chanMsg chan []*pin.PinInscription)
	GetBlockTxHash(blockHeight int64) (txhashList []string)
	PopLevelCount(pop string) (lv int, lastStr string)
	ZmqHashblock()
}
