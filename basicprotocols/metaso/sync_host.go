package metaso

import (
	"context"
	"manindexer/database/mongodb"
	"manindexer/man"
	"manindexer/pin"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func getSyncHeight() (syncHeight int64, hostDataHeight int64) {
	findOp := options.FindOne()
	findOp.SetSort(bson.D{{Key: "blockHeight", Value: -1}})
	var info HostData
	err := mongoClient.Collection(HostDataCollection).FindOne(context.TODO(), bson.D{}, findOp).Decode(&info)
	if err != nil && err == mongo.ErrNoDocuments {
		err = nil
		hostDataHeight = 0
		syncHeight = getPinMinHeight(0)
		return
	}
	hostDataHeight = info.BlockHeight
	syncHeight = getPinMinHeight(info.BlockHeight)
	return
}
func getPinMinHeight(height int64) (minHeight int64) {
	findOp := options.FindOne()
	filter := bson.D{{Key: "genesisheight", Value: bson.D{{Key: "$gt", Value: height}}}, {Key: "host", Value: bson.D{{Key: "$ne", Value: nil}}}}
	findOp.SetSort(bson.D{{Key: "genesisheight", Value: 1}})
	var info pin.PinInscription
	err := mongoClient.Collection(mongodb.PinsCollection).FindOne(context.TODO(), filter, findOp).Decode(&info)
	if err != nil && err == mongo.ErrNoDocuments {
		err = nil
		return
	}
	minHeight = info.GenesisHeight
	return
}
func (metaso *MetaSo) syncHostData() (err error) {
	_, syncHeight := getSyncHeight()
	if syncHeight <= 0 {
		return
	}
	filter := bson.D{{Key: "genesisheight", Value: syncHeight}, {Key: "host", Value: bson.D{{Key: "$ne", Value: nil}}}}
	result, err := mongoClient.Collection(mongodb.PinsCollection).Find(context.TODO(), filter)
	if err != nil {
		return
	}
	var pinList []*pin.PinInscription
	err = result.All(context.TODO(), &pinList)
	saveData := make(map[string]*HostData)
	var insertDocs []interface{}
	for _, pinNode := range pinList {
		if pinNode.Host == "" {
			continue
		}
		fee, size, blockHash, _ := man.ChainAdapter[pinNode.ChainName].GetTxSizeAndFees(pinNode.GenesisTransaction)
		if _, ok := saveData[pinNode.Host]; !ok {
			saveData[pinNode.Host] = &HostData{Host: pinNode.Host, BlockHeight: syncHeight}
		}
		saveData[pinNode.Host].BlockHash = blockHash
		saveData[pinNode.Host].TxSize += size
		saveData[pinNode.Host].TxCount += 1
		saveData[pinNode.Host].TxFee += fee
	}
	for _, d := range saveData {
		insertDocs = append(insertDocs, d)
	}
	insertOpts := options.InsertMany().SetOrdered(false)
	_, err1 := mongoClient.Collection(HostDataCollection).InsertMany(context.TODO(), insertDocs, insertOpts)
	if err1 != nil {
		err = err1
		return
	}
	return
}
