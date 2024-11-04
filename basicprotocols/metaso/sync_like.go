package metaso

import (
	"context"
	"encoding/json"
	"manindexer/common"
	"manindexer/database/mongodb"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (metaso *MetaSo) getLastPayLike() (pinList []*Tweet, err error) {
	last, err := mongodb.GetSyncLastId("tweetlike")
	if err != nil {
		return
	}
	filter := bson.D{
		{Key: "path", Value: "/protocols/paylike"},
	}
	if last != primitive.NilObjectID {
		filter = append(filter, bson.E{Key: "_id", Value: bson.D{{Key: "$gt", Value: last}}})
	}
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "_id", Value: 1}})
	findOptions.SetLimit(500)
	result, err := mongoClient.Collection(mongodb.PinsCollection).Find(context.TODO(), filter, findOptions)
	if err != nil {
		return
	}
	err = result.All(context.TODO(), &pinList)
	return
}
func (metaso *MetaSo) getSynchTweetLike(pinList []*Tweet) (likeList []*TweetLike, err error) {
	var likeToPinIdList []string
	likePinMap := make(map[string][]*TweetLike)
	for _, pinNode := range pinList {
		var pinLike PinLike
		err := json.Unmarshal(pinNode.ContentBody, &pinLike)
		if err != nil {
			continue
		}
		likeToPinIdList = append(likeToPinIdList, pinLike.LikeTo)
		like := TweetLike{
			PinId:         pinNode.Id,
			PinNumber:     pinNode.Number,
			ChainName:     pinNode.ChainName,
			LikeToPinId:   pinLike.LikeTo,
			CreateAddress: pinNode.Address,
			CreateMetaid:  common.GetMetaIdByAddress(pinNode.Address),
			IsLike:        pinLike.IsLike,
			Timestamp:     pinNode.Timestamp,
		}
		likePinMap[pinLike.LikeTo] = append(likePinMap[pinLike.LikeTo], &like)
	}
	filter2 := bson.M{"id": bson.M{"$in": likeToPinIdList}}
	result2, err := mongoClient.Collection(TweetCollection).Find(context.TODO(), filter2)
	if err != nil {
		return
	}
	var list []Tweet
	err = result2.All(context.TODO(), &list)
	if err != nil {
		return
	}
	for _, item := range list {
		likeList = append(likeList, likePinMap[item.Id]...)
	}
	return
}

type deleteLikeInfo struct {
	LikeTo  string
	Address string
}

func (metaso *MetaSo) synchTweetLike() (err error) {
	pinList, err := metaso.getLastPayLike()

	if len(pinList) <= 0 {
		return
	}
	var lastId primitive.ObjectID
	for _, pinNode := range pinList {
		if mongodb.CompareObjectIDs(pinNode.MogoID, lastId) > 0 {
			lastId = pinNode.MogoID
		}
	}
	mongodb.UpdateSyncLastIdLog("tweetlike", lastId)
	list, err := metaso.getSynchTweetLike(pinList)
	if len(list) <= 0 {
		return
	}
	var likeList []interface{}
	var deleteList []deleteLikeInfo
	cntMap := make(map[string]int)

	for _, item := range list {
		if item.IsLike != "1" {
			deleteList = append(deleteList, deleteLikeInfo{LikeTo: item.LikeToPinId, Address: item.CreateAddress})
			cntMap[item.LikeToPinId] -= 1
			continue
		}
		likeList = append(likeList, item)
		cntMap[item.LikeToPinId] += 1
	}
	metaso.updateSynchTweetLike(likeList, deleteList, cntMap)

	return
}
func (metaso *MetaSo) updateSynchTweetLike(likeList []interface{}, deleteList []deleteLikeInfo, cntMap map[string]int) (err error) {
	if len(deleteList) > 0 {
		var models2 []mongo.WriteModel
		for _, deleteItem := range deleteList {
			filter := bson.M{"id": bson.M{"$in": deleteItem.LikeTo}, "createAddress": deleteItem.Address}
			m := mongo.NewDeleteOneModel()
			m.SetFilter(filter)
			models2 = append(models2, m)
		}
		bulkWriteOptions2 := options.BulkWrite().SetOrdered(false)
		mongoClient.Collection(TweetLikeCollection).BulkWrite(context.Background(), models2, bulkWriteOptions2)
	}

	if len(likeList) > 0 {
		insertOpts := options.InsertMany().SetOrdered(false)
		_, err1 := mongoClient.Collection(TweetLikeCollection).InsertMany(context.TODO(), likeList, insertOpts)
		if err1 != nil {
			err = err1
			return
		}
	}

	if len(cntMap) > 0 {
		var models []mongo.WriteModel
		for pinid, cnt := range cntMap {
			filter := bson.D{{Key: "id", Value: pinid}}
			var updateInfo bson.D
			updateInfo = append(updateInfo, bson.E{Key: "likecount", Value: cnt})
			updateInfo = append(updateInfo, bson.E{Key: "hot", Value: cnt})
			update := bson.D{{Key: "$inc", Value: updateInfo}}
			m := mongo.NewUpdateOneModel()
			m.SetFilter(filter).SetUpdate(update)
			models = append(models, m)
		}
		bulkWriteOptions := options.BulkWrite().SetOrdered(false)
		mongoClient.Collection(TweetCollection).BulkWrite(context.Background(), models, bulkWriteOptions)
	}
	return
}
