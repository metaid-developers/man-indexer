package pebbledb

import "manindexer/basicprotocols/metaaccess"

func (pb *Pebble) BatchSaveAccesscontrol(list []*metaaccess.AccessControl) (err error) {
	return
}
func (pb *Pebble) GetControlById(pinId string, isContentId bool) (data *metaaccess.AccessControl, err error) {
	return
}
func (pb *Pebble) BatchSaveAccessPass(passList []*metaaccess.AccessPassData) (err error) {
	return
}
func (pb *Pebble) CheckAccessPass(buyerAddress string, contentPinId string, controlPath string) (data *metaaccess.AccessPassData, err error) {
	return
}
