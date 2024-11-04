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

func (metaso *MetaSo) getLastPayComment() (pinList []*Tweet, err error) {
	last, err := mongodb.GetSyncLastId("tweetcomment")
	if err != nil {
		return
	}
	filter := bson.D{
		{Key: "path", Value: "/protocols/paycomment"},
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
func (metaso *MetaSo) getSynchTweetComment(pinList []*Tweet) (commentList []*TweetComment, err error) {
	var commentToPinIdList []string
	commentPinMap := make(map[string][]*TweetComment)
	for _, pinNode := range pinList {
		var pinComment PinComment
		err := json.Unmarshal(pinNode.ContentBody, &pinComment)
		if err != nil {
			continue
		}
		commentToPinIdList = append(commentToPinIdList, pinComment.CommentTo)
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
		commentPinMap[pinComment.CommentTo] = append(commentPinMap[pinComment.CommentTo], &comment)
	}
	filter2 := bson.M{"id": bson.M{"$in": commentToPinIdList}}
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
		commentList = append(commentList, commentPinMap[item.Id]...)
	}
	return
}

func (metaso *MetaSo) synchTweetComment() (err error) {
	pinList, err := metaso.getLastPayComment()
	if len(pinList) <= 0 {
		return
	}
	var lastId primitive.ObjectID
	for _, pinNode := range pinList {
		if mongodb.CompareObjectIDs(pinNode.MogoID, lastId) > 0 {
			lastId = pinNode.MogoID
		}
	}
	mongodb.UpdateSyncLastIdLog("tweetcomment", lastId)
	list, err := metaso.getSynchTweetComment(pinList)
	if len(list) <= 0 {
		return
	}
	var commentList []interface{}
	cntMap := make(map[string]int)
	for _, item := range list {
		commentList = append(commentList, item)
		cntMap[item.CommentPinId] += 1
	}
	metaso.updateSynchTweetComment(commentList, cntMap)

	return
}
func (metaso *MetaSo) updateSynchTweetComment(commentList []interface{}, cntMap map[string]int) (err error) {

	if len(commentList) > 0 {
		insertOpts := options.InsertMany().SetOrdered(false)
		_, err1 := mongoClient.Collection(TweetCommentCollection).InsertMany(context.TODO(), commentList, insertOpts)
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
			updateInfo = append(updateInfo, bson.E{Key: "commentcount", Value: cnt})
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
