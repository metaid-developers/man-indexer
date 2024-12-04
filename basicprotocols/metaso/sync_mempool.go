package metaso

import (
	"context"
	"encoding/json"
	"manindexer/common"
	"manindexer/database/mongodb"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MempoolDataFilter = bson.D{
	{Key: "$or", Value: bson.A{
		bson.D{{Key: "path", Value: "/protocols/paylike"}},
		bson.D{{Key: "path", Value: "/protocols/paycomment"}},
		bson.D{{Key: "path", Value: "/follow"}},
		bson.D{{Key: "path", Value: "/unfollow"}},
	}},
}

func (metaso *MetaSo) getLastMempoolData() (pinList []*Tweet, err error) {
	last, err := mongodb.GetSyncLastId("mempool")
	if err != nil {
		return
	}
	filter := MempoolDataFilter
	if last != primitive.NilObjectID {
		filter = append(filter, bson.E{Key: "_id", Value: bson.D{{Key: "$gt", Value: last}}})
	}
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "_id", Value: 1}})
	findOptions.SetLimit(500)
	result, err := mongoClient.Collection(mongodb.MempoolPinsCollection).Find(context.TODO(), filter, findOptions)
	if err != nil {
		return
	}
	err = result.All(context.TODO(), &pinList)
	return
}

func (metaso *MetaSo) synchMempoolData() (err error) {
	pinList, err := metaso.getLastMempoolData()
	if len(pinList) <= 0 {
		return
	}
	var lastId primitive.ObjectID
	for _, pinNode := range pinList {
		if mongodb.CompareObjectIDs(pinNode.MogoID, lastId) > 0 {
			lastId = pinNode.MogoID
		}
	}
	mongodb.UpdateSyncLastIdLog("mempool", lastId)
	var mempoolList []interface{}
	for _, pinNode := range pinList {
		mempoolData, err := metaso.getSyncMempoolData(pinNode)
		if err != nil {
			continue
		}
		mempoolList = append(mempoolList, mempoolData)
	}

	if len(mempoolList) <= 0 {
		return
	}
	metaso.updateSyncMempoolData(mempoolList)
	return
}

func (metaso *MetaSo) getSyncMempoolData(pinNode *Tweet) (mempoolData *MempoolData, err error) {
	mempoolData = &MempoolData{}
	mempoolData.CreateTime = pinNode.Timestamp
	mempoolData.Path = pinNode.Path
	mempoolData.PinId = pinNode.Id
	mempoolData.CreateAddress = pinNode.Address
	mempoolData.CreateMetaId = pinNode.MetaId
	switch pinNode.Path {
	case "/protocols/paylike":
		err = getPaylikeData(pinNode, mempoolData)
	case "/protocols/paycomment":
		err = getCommenteData(pinNode, mempoolData)
	case "/follow":
		err = getFollowData(pinNode, mempoolData)
	case "/unfollow":
		err = getUnFollowData(pinNode, mempoolData)
	}
	return
}
func (metaso *MetaSo) updateSyncMempoolData(list []interface{}) (err error) {
	insertOpts := options.InsertMany().SetOrdered(false)
	_, err = mongoClient.Collection(MetaSoMempoolCollection).InsertMany(context.TODO(), list, insertOpts)
	return
}
func getPaylikeData(pinNode *Tweet, mempoolData *MempoolData) (err error) {
	var pinLike PinLike
	err = json.Unmarshal(pinNode.ContentBody, &pinLike)
	if err != nil {
		return
	}
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
	mempoolData.Target = pinLike.LikeTo
	if pinLike.IsLike != "1" {
		mempoolData.IsCancel = 1
	}
	var contentByte []byte
	contentByte, err = json.Marshal(like)
	mempoolData.Content = string(contentByte)
	return
}

func getCommenteData(pinNode *Tweet, mempoolData *MempoolData) (err error) {
	var pinComment PinComment
	err = json.Unmarshal(pinNode.ContentBody, &pinComment)
	if err != nil {
		return
	}
	comment := TweetComment{
		PinId:         pinNode.Id,
		PinNumber:     pinNode.Number,
		ChainName:     pinNode.ChainName,
		CommentPinId:  pinComment.CommentTo,
		Content:       pinComment.Content,
		CreateAddress: pinNode.Address,
		CreateMetaid:  common.GetMetaIdByAddress(pinNode.Address),
		ContentType:   pinComment.ContentType,
		Timestamp:     pinNode.Timestamp,
	}
	mempoolData.Target = pinComment.CommentTo
	var contentByte []byte
	contentByte, err = json.Marshal(comment)
	mempoolData.Content = string(contentByte)
	return
}
func getFollowData(pinNode *Tweet, mempoolData *MempoolData) (err error) {
	mempoolData.Target = pinNode.MetaId
	mempoolData.Content = string(pinNode.ContentBody)
	return
}
func getUnFollowData(pinNode *Tweet, mempoolData *MempoolData) (err error) {
	mempoolData.Target = pinNode.MetaId
	mempoolData.Content = string(pinNode.ContentBody)
	mempoolData.IsCancel = 1
	return
}
