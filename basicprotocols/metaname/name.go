package metaname

import "go.mongodb.org/mongo-driver/bson/primitive"

type MetaName struct {
}
type MetaNamePin struct {
	Id                 string             `json:"id"`
	Number             int64              `json:"number"`
	MetaId             string             `json:"metaid"`
	Address            string             `json:"address"`
	CreateAddress      string             `json:"creator"`
	CreateMetaId       string             `json:"createMetaId"`
	InitialOwner       string             `json:"initialOwner"`
	Output             string             `json:"output"`
	OutputValue        int64              `json:"outputValue"`
	Timestamp          int64              `json:"timestamp"`
	GenesisFee         int64              `json:"genesisFee"`
	GenesisHeight      int64              `json:"genesisHeight"`
	GenesisTransaction string             `json:"genesisTransaction"`
	TxIndex            int                `json:"txIndex"`
	TxInIndex          uint32             `json:"txInIndex"`
	Offset             uint64             `json:"offset"`
	Location           string             `json:"location"`
	Operation          string             `json:"operation"`
	Path               string             `json:"path"`
	ParentPath         string             `json:"parentPath"`
	OriginalPath       string             `json:"originalPath"`
	Encryption         string             `json:"encryption"`
	Version            string             `json:"version"`
	ContentType        string             `json:"contentType"`
	ContentTypeDetect  string             `json:"contentTypeDetect"`
	ContentBody        []byte             `json:"contentBody"`
	ContentLength      uint64             `json:"contentLength"`
	ContentSummary     string             `json:"contentSummary"`
	Status             int                `json:"status"`
	OriginalId         string             `json:"originalId"`
	IsTransfered       bool               `json:"isTransfered"`
	Preview            string             `json:"preview"`
	Content            string             `json:"content"`
	Pop                string             `json:"pop"`
	PopLv              int                `json:"popLv"`
	ChainName          string             `json:"chainName"`
	DataValue          int                `json:"dataValue"`
	Mrc20MintId        []string           `json:"mrc20MintId"`
	MogoID             primitive.ObjectID `bson:"_id,omitempty"`
}
type MetaNameProtocol struct {
	Op     string `json:"op"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
	Rev    string `json:"rev"`
	Relay  string `json:"relay"`
}
type MetaNameData struct {
	Name   string             `json:"name"`
	Avatar string             `json:"avatar"`
	Rev    string             `json:"rev"`
	Relay  string             `json:"relay"`
	PinId  string             `json:"pinId"`
	MogoID primitive.ObjectID `json:"mongoId" bson:"_id"`
}
type MetaNameHistory struct {
	Name      string `json:"name"`
	Op        string `json:"op"`
	OpAddress string `json:"opAddress"`
	OpMetaId  string `json:"opMetaId"`
	Timestamp int64  `json:"timestamp"`
	OpPinId   string `json:"opPinId"`
	OpContent string `json:"opContent"`
}
