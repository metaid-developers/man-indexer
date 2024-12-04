package metaso

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MetaSo struct {
}

type Tweet struct {
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
	LikeCount          int                `json:"likeCount" bson:"likecount"`
	CommentCount       int                `json:"commentCount" bson:"commentcount"`
	ShareCount         int                `json:"shareCount" bson:"sharecount"`
	Hot                int                `json:"hot" bson:"hot"`
}
type SyncLastId struct {
	Tweet        primitive.ObjectID `bson:"tweet"`
	TweetLike    primitive.ObjectID `bson:"tweetlike"`
	TweetComment primitive.ObjectID `bson:"tweetcomment"`
}
type TweetLike struct {
	PinId         string `json:"pinId" bson:"pinid"`
	PinNumber     int64  `json:"pinNumber" bson:"pinnumber"`
	ChainName     string `json:"chainName" bson:"chainname"`
	LikeToPinId   string `json:"likeToPinId" bson:"liketopinid"`
	CreateAddress string `json:"createAddress" bson:"createaddress"`
	CreateMetaid  string `json:"CreateMetaid" bson:"createmetaid"`
	IsLike        string `json:"isLike" bson:"islike"`
	Timestamp     int64  `json:"timestamp" bson:"timestamp"`
}
type TweetComment struct {
	PinId         string `json:"pinId" bson:"pinid"`
	PinNumber     int64  `json:"pinNumber" bson:"pinnumber"`
	ChainName     string `json:"chainName" bson:"chainname"`
	CommentPinId  string `json:"commentToPinId" bson:"commentpinid"`
	CreateAddress string `json:"createAddress" bson:"createaddress"`
	CreateMetaid  string `json:"CreateMetaid" bson:"createmetaid"`
	Content       string `json:"content" bson:"content"`
	ContentType   string `json:"contentType" bson:"contenttype"`
	Timestamp     int64  `json:"timestamp" bson:"timestamp"`
}
type PinLike struct {
	IsLike string `json:"isLike" bson:"islike"`
	LikeTo string `json:"likeTo" bson:"liketo"`
}
type PinComment struct {
	CommentTo   string `json:"commentTo" bson:"commentto"`
	Content     string `json:"content" bson:"content"`
	ContentType string `json:"contentType" bson:"contenttype"`
}
type HostData struct {
	Host        string `json:"host" bson:"host"`
	BlockHeight int64  `json:"blockHeight" bson:"blockHeight"`
	BlockHash   string `json:"blockHash" bson:"blockHash"`
	TxCount     int64  `json:"txCount" bson:"txCount"`
	TxSize      int64  `json:"txSize" bson:"txSize"`
	TxFee       int64  `json:"txFee" bson:"txFee"`
}

type PayBuzz struct {
	PublicContent  string   `json:"publicContent"`
	EncryptContent string   `json:"encryptContent"`
	ContentType    string   `json:"contentType"`
	PublicFiles    []string `json:"publicFiles"`
	EncryptFiles   []string `json:"encryptFiles"`
}
type Mrc20DeployInfo struct {
	MogoID       primitive.ObjectID  `bson:"_id,omitempty"`
	Tick         string              `json:"tick"`
	TokenName    string              `json:"tokenName"`
	Decimals     string              `json:"decimals"`
	AmtPerMint   string              `json:"amtPerMint"`
	MintCount    uint64              `json:"mintCount"`
	BeginHeight  string              `json:"beginHeight"`
	EndHeight    string              `json:"endHeight"`
	Metadata     string              `json:"metadata"`
	DeployType   string              `json:"type"`
	PremineCount uint64              `json:"premineCount"`
	PinCheck     Mrc20DeployQual     `json:"pinCheck"`
	PayCheck     Mrc20DeployPayCheck `json:"payCheck"`
	TotalMinted  uint64              `json:"totalMinted"`
	Mrc20Id      string              `json:"mrc20Id"`
	PinNumber    int64               `json:"pinNumber"`
	Chain        string              `json:"chain"`
	Holders      uint64              `json:"holders"`
	TxCount      uint64              `json:"txCount"`
	MetaId       string              `json:"metaId"`
	Address      string              `json:"address"`
	DeployTime   int64               `json:"deployTime"`
	IdCoin       int                 `json:"idCoin"`
}
type Mrc20DeployQual struct {
	Creator string `json:"creator"`
	Lv      string `json:"lvl"`
	Path    string `json:"path"`
	Count   string `json:"count"`
}
type Mrc20DeployPayCheck struct {
	PayTo     string `json:"payTo"`
	PayAmount string `json:"payAmount"`
}
type MempoolData struct {
	Path          string `json:"path"`
	PinId         string `json:"pinId"`
	CreateTime    int64  `json:"createTime"`
	Target        string `json:"target"`
	Content       string `json:"content"`
	IsCancel      int    `json:"isCancel"`
	CreateMetaId  string `json:"createMetaId"`
	CreateAddress string `json:"createAddress"`
}
