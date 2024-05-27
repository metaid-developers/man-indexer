package pebbledb

import "manindexer/pin"

func (pb *Pebble) BatchUpsertFollowData(followData []*pin.FollowData) (err error) {
	return
}
func (pb *Pebble) GetFollowDataByMetaId(metaId string, myFollow bool, followDetail bool, cursor int64, size int64) (metaIdList []interface{}, total int64, err error) {
	return
}
func (pb *Pebble) GetFollowRecord(metaId string, followMetaid string) (followData pin.FollowData, err error) {
	return
}
