package mrc721

import (
	"context"
	"manindexer/common"
	"manindexer/database/mongodb"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (mrc721 *Mrc721) Synchronization() {
	connectMongoDb()
	for {
		mrc721.sync()
		mrc721.syncTransfer()
		time.Sleep(time.Second * 10)
	}
}
func (mrc721 *Mrc721) sync() (err error) {
	last, err := mongodb.GetSyncLastId("mrc721")
	if err != nil {
		return
	}
	var pinList []*Mrc721Pin
	filter := bson.D{
		{Key: "path", Value: bson.D{{Key: "$regex", Value: "^/nft/mrc721"}}},
	}
	if last != primitive.NilObjectID {
		filter = append(filter, bson.E{Key: "_id", Value: bson.D{{Key: "$gt", Value: last}}})
	}
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "_id", Value: 1}})
	findOptions.SetLimit(500)
	result, err := mongoClient.Collection(mongodb.PinsView).Find(context.TODO(), filter)
	if err != nil {
		return
	}
	result.All(context.TODO(), &pinList)
	if len(pinList) <= 0 {
		return
	}
	var lastId primitive.ObjectID
	mrc721.PinHandle(pinList)
	for _, doc := range pinList {
		if mongodb.CompareObjectIDs(doc.MogoID, lastId) > 0 {
			lastId = doc.MogoID
		}
	}
	mongodb.UpdateSyncLastIdLog("mrc721", lastId)
	return
}
func (mrc721 *Mrc721) syncTransfer() (err error) {
	last, err := mongodb.GetSyncLastId("mrc721transfer")
	if err != nil {
		return
	}
	var dataList []*PinTransferHistory
	filter := bson.D{}
	if last != primitive.NilObjectID {
		filter = append(filter, bson.E{Key: "_id", Value: bson.D{{Key: "$gt", Value: last}}})
	}
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "_id", Value: 1}})
	findOptions.SetLimit(500)
	result, err := mongoClient.Collection(mongodb.PinTransferHistory).Find(context.TODO(), filter)
	if err != nil {
		return
	}
	result.All(context.TODO(), &dataList)
	if len(dataList) <= 0 {
		return
	}
	var lastId primitive.ObjectID
	//mrc721.PinHandle(pinList)
	transferMap := make(map[string]string)
	var pinList []string
	for _, doc := range dataList {
		if mongodb.CompareObjectIDs(doc.MogoID, lastId) > 0 {
			lastId = doc.MogoID
		}
		transferMap[doc.PinId] = doc.ToAddress
		pinList = append(pinList, doc.PinId)
	}
	err = transferItemHandle(pinList, transferMap)
	if err == nil {
		mongodb.UpdateSyncLastIdLog("mrc721transfer", lastId)
	}

	return
}

func transferItemHandle(pinIdList []string, transferMap map[string]string) (err error) {
	filter := bson.M{"itempinid": bson.M{"$in": pinIdList}}
	result, err := mongoClient.Collection(Mrc721Item).Find(context.TODO(), filter)
	if err != nil {
		return
	}
	var itemList []*Mrc721ItemDescPin
	err = result.All(context.TODO(), &itemList)
	if err != nil || len(itemList) <= 0 {
		return
	}
	var models []mongo.WriteModel
	for _, item := range itemList {
		filter := bson.D{{Key: "itempinid", Value: item.ItemPinId}}
		var updateInfo bson.D
		newAddress := transferMap[item.ItemPinId]
		updateInfo = append(updateInfo, bson.E{Key: "address", Value: newAddress})
		updateInfo = append(updateInfo, bson.E{Key: "metaid", Value: common.GetMetaIdByAddress(newAddress)})
		update := bson.D{{Key: "$set", Value: updateInfo}}
		m := mongo.NewUpdateOneModel()
		m.SetFilter(filter).SetUpdate(update)
		models = append(models, m)
	}
	bulkWriteOptions := options.BulkWrite().SetOrdered(false)
	_, err = mongoClient.Collection(Mrc721Item).BulkWrite(context.Background(), models, bulkWriteOptions)
	return
}
