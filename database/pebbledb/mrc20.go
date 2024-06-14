package pebbledb

import (
	"manindexer/mrc20"
)

func (pb *Pebble) GetMrc20TickInfo(tick string) (info mrc20.Mrc20DeployInfo, err error) {
	return
}

func (pb *Pebble) SaveMrc20Pin(data []mrc20.Mrc20Utxo) (err error) {
	return
}
func (pb *Pebble) SaveMrc20Tick(data []mrc20.Mrc20DeployInfo) (err error) {
	return
}
func (pb *Pebble) GetMrc20TickPageList(page int64, size int64, order string) (total int64, list []mrc20.Mrc20DeployInfo, err error) {
	return
}
func (pb *Pebble) AddMrc20Shovel(shovel string, pinId string) (err error) {
	return
}
func (pb *Pebble) GetMrc20Shovel(shovels []string) (data map[string]mrc20.Mrc20Shovel, err error) {
	return
}
func (pb *Pebble) UpdateMrc20TickInfo(tickId string, minted int64) (err error) {
	return
}
func (pb *Pebble) GetMrc20ByAddressAndTick(address string, tickId string) (list []mrc20.Mrc20Utxo, err error) {
	return
}
func (pb *Pebble) GetMrc20HistoryPageList(tickId string, page int64, size int64) (list []mrc20.Mrc20Utxo, err error) {
	return
}
func (pb *Pebble) GetMrc20UtxoByOutPutList(outputList []string) (list []*mrc20.Mrc20Utxo, err error) {
	return
}
func (pb *Pebble) UpdateMrc20Utxo(list []*mrc20.Mrc20Utxo) (err error) {
	return
}
