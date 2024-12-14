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
type MetaNameProtocol struct {
	Space    string `json:"space"`
	Name     string `json:"name"`
	FullName string `json:"fullName"`
	Metadata string `json:"metadata"`
	Rev      string `json:"rev"`
	Relay    string `json:"relay"`
}
type MetaNameData struct {
	Space    string             `json:"space"`
	Name     string             `json:"name"`
	FullName string             `json:"fullName"`
	Metadata string             `json:"metadata"`
	Rev      string             `json:"rev"`
	Relay    string             `json:"relay"`
	PinId    string             `json:"pinId"`
	Address  string             `json:"opAddress"`
	MetaId   string             `json:"opMetaId"`
	MogoID   primitive.ObjectID `json:"mongoId" bson:"_id,omitempty"`
}
type MetaNameHistory struct {
	Name      string `json:"name"`
	Space     string `json:"space"`
	FullName  string `json:"fullName"`
	Op        string `json:"op"`
	OpAddress string `json:"opAddress"`
	OpMetaId  string `json:"opMetaId"`
	Timestamp int64  `json:"timestamp"`
	OpPinId   string `json:"opPinId"`
	OpContent string `json:"opContent"`
}
