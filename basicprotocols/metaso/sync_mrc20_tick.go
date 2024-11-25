package metaso

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"manindexer/common"
	"manindexer/database/mongodb"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func getLastMrc20TickId() (lastId string) {
	findOp := options.FindOne()
	findOp.SetSort(bson.D{{Key: "_id", Value: -1}})
	var info Mrc20DeployInfo
	err := mongoClient.Collection(MetasoTickCollection).FindOne(context.TODO(), bson.D{}, findOp).Decode(&info)
	if err != nil && err == mongo.ErrNoDocuments {
		err = nil
		return
	}
	lastId = info.MogoID.String()
	return
}

func (metaso *MetaSo) syncMrc20TickData() (err error) {
	mongoId := getLastMrc20TickId()
	filter := bson.D{}
	if mongoId != "" {
		var objectId primitive.ObjectID
		objectId, err = primitive.ObjectIDFromHex(mongoId)
		if err != nil {
			return
		}
		filter = append(filter, bson.E{Key: "_id", Value: bson.D{{Key: "$lt", Value: objectId}}})
	}
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "_id", Value: 1}})
	findOptions.SetLimit(500)
	result, err := mongoClient.Collection(mongodb.Mrc20TickCollection).Find(context.TODO(), filter, findOptions)
	if err != nil {
		return
	}

	var list []*Mrc20DeployInfo
	err = result.All(context.TODO(), &list)
	var insertDocs []interface{}
	for _, deploy := range list {
		net := "main"
		if common.TestNet != "0" {
			net = "testnet"
		}
		idcoin := checkIdCoins(net, deploy.Tick, deploy.Metadata, deploy.DeployTime)
		if idcoin == "id-coins" {
			deploy.IdCoin = 1
		}
		insertDocs = append(insertDocs, deploy)
	}
	insertOpts := options.InsertMany().SetOrdered(false)
	_, err1 := mongoClient.Collection(MetasoTickCollection).InsertMany(context.TODO(), insertDocs, insertOpts)
	if err1 != nil {
		err = err1
		return
	}
	return
}

type MetaDataInfo struct {
	TickSign string `json:"tickSign"`
}

func checkIdCoins(net, tick, metaData string, deployTime int64) string {
	var (
		metaDataInfo  *MetaDataInfo
		err           error
		tickSign      string = ""
		signPublic    string = "022a5babbf8c7c8e304884979fd0d57837d215956b1d399927f2e44a4147ae6d05"
		signTimestamp int64  = 0
		verify        bool   = false
	)
	if net == "testnet" {
		signPublic = "0321d2dfca8d70476df45be51eab616d525307076de9f2cbc218350bf01419a153"
		signTimestamp = 1722579499107
	}
	err = json.Unmarshal([]byte(metaData), &metaDataInfo)
	if err != nil {
		return ""
	}
	tickSign = metaDataInfo.TickSign
	if tickSign == "" {
		return ""
	}

	if deployTime <= 0 {
		return ""
	}
	if signTimestamp > 0 && deployTime <= signTimestamp {
		return "id-coins"
	}
	verify, err = verifyIdCoinSign(strings.ToUpper(tick), tickSign, signPublic)
	if err != nil {
		return ""
	}
	if !verify {
		return ""
	}
	return "id-coins"
}

func verifyIdCoinSign(message, messageSign, publicKey string) (bool, error) {

	// Decode hex-encoded serialized public key.
	pubKeyBytes, err := hex.DecodeString(publicKey)
	if err != nil {
		return false, err
	}
	pubKey, err := btcec.ParsePubKey(pubKeyBytes)
	if err != nil {
		return false, err
	}

	// Decode hex-encoded serialized signature.
	sigBytes, err := hex.DecodeString(messageSign)
	if err != nil {
		return false, err
	}
	signature, err := ecdsa.ParseSignature(sigBytes)
	if err != nil {
		return false, err
	}

	// Verify the signature for the message using the public key.
	messageHash := chainhash.DoubleHashB([]byte(message))
	verified := signature.Verify(messageHash, pubKey)
	return verified, nil
}
