package mrc721

import "go.mongodb.org/mongo-driver/bson/primitive"

const (
	ErrOperation           = "operation is error"
	ErrPinContent          = "pin content is error"
	ErrTotalSupply         = "totalSupply  must be between -1 and 1e12"
	ErrRoyaltyRate         = "royaltyRate  must be between 0 and 20"
	ErrCollectionExist     = "collectionName already exists"
	ErrCollectionNotExist  = "collectionName not exists"
	ErrTotalSupplyEexceeds = "exceeds total supply"
)

type Mrc721 struct {
}
type Mrc721Pin struct {
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
type PinTransferHistory struct {
	PinId          string             `json:"pinId"`
	TransferTime   int64              `json:"transferTime"`
	TransferHeight int64              `json:"transferHeight"`
	TransferBlock  string             `json:"transferBlock"`
	TransferTx     string             `json:"transferTx"`
	ChainName      string             `json:"chainName"`
	FromAddress    string             `json:"fromAddress"`
	ToAddress      string             `json:"toAddress"`
	MogoID         primitive.ObjectID `bson:"_id,omitempty"`
}
type Mrc721CollectionDesc struct {
	Name        string `json:"name"`
	TotalSupply int64  `json:"totalsupply"`
	RoyaltyRate int    `json:"royaltyrate"`
	Desc        string `json:"desc"`
	Website     string `json:"website"`
	Cover       string `json:"cover"`
	Metadata    string `json:"metadata"`
}

type Mrc721CollectionDescPin struct {
	CollectionName string      `json:"collectionname"`
	Name           string      `json:"name"`
	TotalSupply    int64       `json:"totalsupply"`
	RoyaltyRate    int         `json:"royaltyrate"`
	Desc           string      `json:"desc"`
	Website        string      `json:"website"`
	Cover          string      `json:"cover"`
	Metadata       interface{} `json:"metadata"`
	PinId          string      `json:"pinid"`
	Address        string      `json:"address"`
	MetaId         string      `json:"metaid"`
	CreateTime     int64       `json:"createtime"`
	TotalNum       int64       `json:"totalnum"`
}

type Mrc721ItemDescList struct {
	Items []Mrc721ItemDesc `json:"items"`
}

type Mrc721ItemDesc struct {
	PinId    string `json:"pinid"`
	Name     string `json:"name"`
	Desc     string `json:"desc"`
	Cover    string `json:"cover"`
	Metadata string `json:"metadata"`
}

type Mrc721ItemDescPin struct {
	CollectionPinId   string `json:"collectionPinId"`
	CollectionName    string `json:"collectionName"`
	ItemPinId         string `json:"itemPinId"`
	ItemPinNumber     int64  `json:"itemPinNumber"`
	DescPinId         string `json:"descPinId"`
	Name              string `json:"name"`
	Desc              string `json:"desc"`
	Cover             string `json:"cover"`
	Metadata          string `json:"metaData"`
	CreateTime        int64  `json:"createTime"`
	Address           string `json:"address"`
	Content           []byte `json:"content"`
	MetaId            string `json:"metaId"`
	DescAdded         bool   `json:"descadded"`
	ContentType       string `json:"contentType"`
	ContentTypeDetect string `json:"contentTypeDetect"`
	ContentString     string `json:"contentString"`
}
