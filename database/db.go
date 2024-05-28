package database

import (
	"manindexer/mrc20"
	"manindexer/pin"
)

type Generator struct {
	Collection     string   `json:"collection"`
	Action         string   `json:"action"`         //get,count,sum
	Field          []string `json:"field"`          //["a","b"]
	FilterRelation string   `json:"filterRelation"` //and,or
	// [{"operator":"=","key":"a","value":"1"},{"operator":"=","key":"b","value":"2"}]
	Filters []GeneratorFilter `json:"filter"`
	Limit   int64             `json:"limit"`
	Cursor  int64             `json:"cursor"`
	Sort    []string          `json:"sort"` //["a","desc"]
}
type GeneratorFilter struct {
	Operator string      `json:"operator"` // =,>,<,>=,<=
	Key      string      `json:"key"`
	Value    interface{} `json:"value"`
}
type Db interface {
	InitDatabase()
	GetMaxHeight() (height int64, err error)
	GetMaxNumber() (number int64)
	GetMaxMetaIdNumber() (number int64)
	BatchAddPins(pins []interface{}) (err error)
	BatchUpdatePins(pins []*pin.PinInscription) (err error)
	UpdateTransferPin(trasferMap map[string]*pin.PinTransferInfo) (err error)
	AddMempoolPin(pin *pin.PinInscription) (err error)
	GetMetaIdInfo(address string, mempool bool) (info *pin.MetaIdInfo, unconfirmed string, err error)
	BatchUpsertMetaIdInfo(infoList map[string]*pin.MetaIdInfo) (err error)
	BatchUpsertMetaIdInfoAddition(infoList []*pin.MetaIdInfoAdditional) (err error)
	BatchAddPinTree(data []interface{}) (err error)
	GetPinPageList(page int64, size int64) (pins []*pin.PinInscription, err error)
	GetPinUtxoCountByAddress(address string) (utxoNum int64, utxoSum int64, err error)
	GetMempoolPinPageList(page int64, size int64) (pins []*pin.PinInscription, err error)
	DeleteMempoolInscription(txIds []string) (err error)
	GetPinListByAddress(address string, addressType string, cursor int64, size int64, cnt string, path string) (pins []*pin.PinInscription, total int64, err error)
	GetPinByNumberOrId(number string) (pinInscription *pin.PinInscription, err error)
	GetPinByOutput(output string) (pinInscription *pin.PinInscription, err error)
	GetPinByMeatIdOrId(key string) (pinInscription *pin.PinInscription, err error)
	GetPinListByIdList(idList []string) (pinList []*pin.PinInscription, err error)
	GetPinListByOutPutList(outputList []string) (pinList []*pin.PinInscription, err error)
	GetMetaIdPageList(page int64, size int64) (pins []*pin.MetaIdInfo, err error)
	GetBlockPin(height int64, size int64) (pins []*pin.PinInscription, total int64, err error)
	GetMetaIdPin(address string, page int64, size int64) (pins []*pin.PinInscription, total int64, err error)
	Count() (count pin.PinCount)
	GetChildNodeById(pinId string) (pins []*pin.PinInscription, err error)
	GetParentNodeById(pinId string) (pinnode *pin.PinInscription, err error)
	GetAllPinByPath(page, limit int64, path string, metaidList []string) (pins []*pin.PinInscription, total int64, err error)
	BatchAddProtocolData(pins []*pin.PinInscription) (err error)
	GeneratorFind(generator Generator) (data []map[string]interface{}, err error)
	//mempoolTransfer
	AddMempoolTransfer(transferData *pin.MemPoolTrasferPin) (err error)
	GetMempoolTransfer(address string, act string) (list []*pin.MemPoolTrasferPin, err error)
	GetMempoolTransferById(pinId string) (result *pin.MemPoolTrasferPin, err error)
	//follow
	BatchUpsertFollowData(followData []*pin.FollowData) (err error)
	GetFollowDataByMetaId(metaId string, myFollow bool, followDetail bool, cursor int64, size int64) (metaIdList []interface{}, total int64, err error)
	GetFollowRecord(metaId string, followMetaid string) (followData pin.FollowData, err error)
	//mrc20
	GetMrc20TickInfo(tick string) (info mrc20.Mrc20DeployInfo, err error)
	SaveMrc20Pin(data []mrc20.Mrc20Pin) (err error)
	SaveMrc20Tick(data []mrc20.Mrc20DeployInfo) (err error)
	GetMrc20TickPageList(page int64, size int64) (list []mrc20.Mrc20DeployInfo, err error)
	AddMrc20Shovel(shovel string, pinId string) (err error)
	GetMrc20Shovel(shovels []string) (data map[string]mrc20.Mrc20Shovel, err error)
	UpdateMrc20TickInfo(tickId string, minted int64) (err error)
	GetMrc20ByAddressAndTick(address string, tickId string) (list []mrc20.Mrc20Pin, err error)
	GetMrc20HistoryPageList(tickId string, page int64, size int64) (list []mrc20.Mrc20Pin, err error)
}
