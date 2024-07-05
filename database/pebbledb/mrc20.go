package pebbledb

import (
	"manindexer/mrc20"
	"manindexer/pin"
)

func (pb *Pebble) GetMrc20TickInfo(mrc20Id string, tick string) (info mrc20.Mrc20DeployInfo, err error) {
	return
}

func (pb *Pebble) SaveMrc20Pin(data []mrc20.Mrc20Utxo) (err error) {
	return
}
func (pb *Pebble) SaveMrc20Tick(data []mrc20.Mrc20DeployInfo) (err error) {
	return
}
func (pb *Pebble) GetMrc20TickPageList(cursor int64, size int64, order string, completed string, orderType string) (total int64, list []mrc20.Mrc20DeployInfo, err error) {
	return
}
func (pb *Pebble) AddMrc20Shovel(shovelList []string, pinId string, mrc20Id string) (err error) {
	return
}
func (pb *Pebble) GetMrc20Shovel(shovels []string, mrc20Id string) (data map[string]mrc20.Mrc20Shovel, err error) {
	return
}
func (pb *Pebble) UpdateMrc20TickInfo(tickId string, txPoint string, minted int64) (err error) {
	return
}
func (pb *Pebble) UpdateMrc20TickHolder(tickId string, txNum int64) (err error) {
	return
}
func (pb *Pebble) GetMrc20ByAddressAndTick(address string, tickId string) (list []mrc20.Mrc20Utxo, err error) {
	return
}
func (pb *Pebble) GetMrc20HistoryPageList(tickId string, isPage bool, page int64, size int64) (list []mrc20.Mrc20Utxo, total int64, err error) {
	return
}
func (pb *Pebble) GetMrc20UtxoByOutPutList(outputList []string) (list []*mrc20.Mrc20Utxo, err error) {
	return
}
func (pb *Pebble) UpdateMrc20Utxo(list []*mrc20.Mrc20Utxo) (err error) {
	return
}
func (pb *Pebble) GetHistoryByAddress(tickId string, address string, page int64, size int64, status string, verify string) (list []mrc20.Mrc20Utxo, total int64, err error) {
	return
}
func (pb *Pebble) GetMrc20BalanceByAddress(address string, cursor int64, size int64) (list []mrc20.Mrc20Balance, total int64, err error) {
	return
}
func (pb *Pebble) GetHistoryByTx(txId string, index int64, cursor int64, size int64) (list []mrc20.Mrc20Utxo, total int64, err error) {
	return
}
func (pb *Pebble) GetShovelListByAddress(address string, mrc20Id string, creator string, lv int, path, query, key, operator, value string, cursor int64, size int64) (list []*pin.PinInscription, total int64, err error) {
	return
}
func (pb *Pebble) GetUsedShovelIdListByAddress(address string, tickId string, cursor int64, size int64) (list []*string, total int64, err error) {
	return
}
