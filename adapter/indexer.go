package adapter

import "manindexer/pin"

type Indexer interface {
	CatchPins(blockHeight int64) (pinInscriptions []*pin.PinInscription, txInList []string)
	CatchTransfer(idMap map[string]struct{}) (addressMap map[string]string)
	GetAddress(pkScript []byte) (address string)
	ZmqRun(chanMsg chan []*pin.PinInscription)
	GetBlockTxHash(blockHeight int64) (txhashList []string)
	ZmqHashblock()
}
