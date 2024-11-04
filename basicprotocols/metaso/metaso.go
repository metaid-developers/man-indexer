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
