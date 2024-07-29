package pebbledb

import "manindexer/mrc721"

func (pb *Pebble) SaveMrc721Collection(collection *mrc721.Mrc721CollectionDescPin) (err error) {
	return
}
func (pb *Pebble) GetMrc721Collection(collectionName, pinId string) (data *mrc721.Mrc721CollectionDescPin, err error) {
	return
}
func (pb *Pebble) GetMrc721CollectionList(nameList []string, cnt bool) (data []*mrc721.Mrc721CollectionDescPin, total int64, err error) {
	return
}

func (pb *Pebble) BatchUpdateMrc721CollectionCount(nameList []string) (err error) {
	return
}
func (pb *Pebble) SaveMrc721Item(itemList []*mrc721.Mrc721ItemDescPin) (err error) {
	return
}
func (pb *Pebble) GetMrc721ItemList(collectionName string, pinIdList []string, cnt bool) (itemList []*mrc721.Mrc721ItemDescPin, total int64, err error) {
	return
}
func (pb *Pebble) UpdateMrc721ItemDesc(itemList []*mrc721.Mrc721ItemDescPin) (err error) {
	return
}
