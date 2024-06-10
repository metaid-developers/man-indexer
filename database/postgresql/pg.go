package postgresql

import (
	"log"
	"manindexer/database"
	"manindexer/pin"
)

type Postgresql struct{}

func (pg *Postgresql) InitDatabase() {

}
func (pg *Postgresql) GetMaxHeight(chainName string) (height int64, err error) {
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
func (pg *Postgresql) UpdateTransferPin(trasferMap map[string]*pin.PinTransferInfo) (err error) {
	return
}

func (pg *Postgresql) GetMetaIdInfo(address string, mempool bool, metaid string) (info *pin.MetaIdInfo, unconfirmed string, err error) {
	return
}
func (pg *Postgresql) BatchUpsertMetaIdInfo(infoList map[string]*pin.MetaIdInfo) (err error) {
	return
}
func (pg *Postgresql) BatchAddPinTree(data []interface{}) (err error) {
	return
}
func (pg *Postgresql) GetPinPageList(page int64, size int64) (pins []*pin.PinInscription, err error) {
	return
}
func (pg *Postgresql) GetPinUtxoCountByAddress(address string) (utxoNum int64, utxoSum int64, err error) {
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
func (pg *Postgresql) GetPinByOutput(output string) (pinInscription *pin.PinInscription, err error) {
	return
}
func (pg *Postgresql) GetPinByMeatIdOrId(key string) (pinInscription *pin.PinInscription, err error) {
	return
}
func (pg *Postgresql) GetPinListByIdList(idList []string) (pinList []*pin.PinInscription, err error) {
	return
}
func (pg *Postgresql) GetPinListByOutPutList(outputList []string) (pinList []*pin.PinInscription, err error) {
	return
}
func (pg *Postgresql) GetMetaIdPageList(page int64, size int64, order string) (pins []*pin.MetaIdInfo, err error) {
	return
}
func (pg *Postgresql) GetBlockPin(height int64, size int64) (pins []*pin.PinInscription, total int64, err error) {
	return
}
func (pg *Postgresql) GetMetaIdPin(address string, page int64, size int64) (pins []*pin.PinInscription, total int64, err error) {
	return
}
func (pg *Postgresql) Count() (count pin.PinCount) {
	return
}
func (pg *Postgresql) GetPinListByAddress(address string, addressType string, cursor int64, size int64, cnt string, path string) (pins []*pin.PinInscription, total int64, err error) {
	return
}

func (pg *Postgresql) GetChildNodeById(pinId string) (pins []*pin.PinInscription, err error) {
	return
}
func (pg *Postgresql) GetParentNodeById(pinId string) (pinnode *pin.PinInscription, err error) {
	return
}
func (pg *Postgresql) GetAllPinByPath(page, limit int64, path string, metaidList []string) (pins []*pin.PinInscription, total int64, err error) {
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
func (pg *Postgresql) BatchUpsertFollowData(followData []*pin.FollowData) (err error) {
	return
}
func (pg *Postgresql) GetFollowDataByMetaId(metaId string, myFollow bool, followDetail bool, cursor int64, size int64) (metaIdList []interface{}, total int64, err error) {
	return
}
func (pg *Postgresql) GetFollowRecord(metaId string, followMetaid string) (followData pin.FollowData, err error) {
	return
}
func (pg *Postgresql) BatchUpsertMetaIdInfoAddition(infoList []*pin.MetaIdInfoAdditional) (err error) {
	return
}
func (pg *Postgresql) AddMempoolTransfer(transferData *pin.MemPoolTrasferPin) (err error) {
	return
}
func (pg *Postgresql) GetMempoolTransfer(address string, act string) (list []*pin.MemPoolTrasferPin, err error) {
	return
}
func (pg *Postgresql) GetMempoolTransferById(pinId string) (result *pin.MemPoolTrasferPin, err error) {
	return
}
func (pg *Postgresql) GetDataValueByMetaIdList(list []string) (result []*pin.MetaIdDataValue, err error) {
	return
}
