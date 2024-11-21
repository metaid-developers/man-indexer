package postgresql

import "manindexer/basicprotocols/metaaccess"

func (pg *Postgresql) BatchSaveAccesscontrol(list []*metaaccess.AccessControl) (err error) {
	return
}
func (pg *Postgresql) GetControlById(pinId string, isContentId bool) (data *metaaccess.AccessControl, err error) {
	return
}
func (pg *Postgresql) BatchSaveAccessPass(passList []*metaaccess.AccessPassData) (err error) {
	return
}
func (pg *Postgresql) CheckAccessPass(buyerAddress string, contentPinId string, controlPath string) (data *metaaccess.AccessPassData, err error) {
	return
}
