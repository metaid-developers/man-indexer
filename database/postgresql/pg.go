package postgresql

import (
	"log"
	"manindexer/database"
	"manindexer/pin"
)

type Postgresql struct{}

func (pg *Postgresql) InitDatabase() {

}
func (pg *Postgresql) GetMaxHeight() (height int64, err error) {
	log.Println("Postgresql TODO")
	return
}
func (pg *Postgresql) GetMaxNumber() (number int64) {
	log.Println("Postgresql TODO")
	return
}
func (pg *Postgresql) GetMaxMetaIdNumber() (number int64) {
	log.Println("Postgresql TODO")
	return
}

func (pg *Postgresql) BatchAddPins(pins []interface{}) (err error) {
	log.Println("Postgresql TODO")
	return
}
func (pg *Postgresql) BatchUpdatePins(pins []*pin.PinInscription) (err error) {
	return
}
func (pg *Postgresql) UpdateTransferPin(addressMap map[string]string) (err error) {
	return
}
func (pg *Postgresql) GetRootTxId(address string) (metaId string, err error) {
	return
}
func (pg *Postgresql) GetMetaIdInfo(rootTxid string, key string) (info *pin.MetaIdInfo, unconfirmed string, err error) {
	return
}
func (pg *Postgresql) BatchUpsertMetaIdInfo(infoList []*pin.MetaIdInfo) (err error) {
	return
}
func (pg *Postgresql) BatchAddPinTree(data []interface{}) (err error) {
	return
}
func (pg *Postgresql) GetPinPageList(page int64, size int64) (pins []*pin.PinInscription, err error) {
	return
}
func (pg *Postgresql) GetMempoolPinPageList(page int64, size int64) (pins []*pin.PinInscription, err error) {
	return
}
func (pg *Postgresql) DeleteMempoolInscription(txIds []string) (err error) {
	return
}
func (pg *Postgresql) GetPinByNumberOrId(number string) (pinInscription *pin.PinInscription, err error) {
	return
}
func (pg *Postgresql) GetPinListByIdList(idList []string) (pinList []*pin.PinInscription, err error) {
	return
}
func (pg *Postgresql) GetMetaIdPageList(page int64, size int64) (pins []*pin.MetaIdInfo, err error) {
	return
}
func (pg *Postgresql) GetBlockPin(height int64, size int64) (pins []*pin.PinInscription, total int64, err error) {
	return
}
func (pg *Postgresql) GetMetaIdPin(roottxid string, page int64, size int64) (pins []*pin.PinInscription, total int64, err error) {
	return
}
func (pg *Postgresql) Count() (count pin.PinCount) {
	return
}
func (pg *Postgresql) GetPinListByAddress(address string, addressType string, cursor int64, size int64) (pins []*pin.PinInscription, err error) {
	return
}
func (pg *Postgresql) GetPinRootByAddress(address string) (pin *pin.PinInscription, err error) {
	return
}
func (pg *Postgresql) GetChildNodeById(roottxid string) (pins []*pin.PinInscription, err error) {
	return
}
func (pg *Postgresql) GetParentNodeById(pinId string) (pinnode *pin.PinInscription, err error) {
	return
}
func (pg *Postgresql) GetAllPinByPath(page, limit int64, path string) (pins []*pin.PinInscription, total int64, err error) {
	return
}
func (pg *Postgresql) AddMempoolPin(pin *pin.PinInscription) (err error) {
	return
}
func (pg *Postgresql) BatchAddProtocolData(pins []*pin.PinInscription) (err error) {
	return
}
func (pg *Postgresql) GeneratorFind(generator database.Generator) (data []map[string]interface{}, err error) {
	return
}
