package metaso

import (
	"context"
	"manindexer/database/mongodb"
	"manindexer/pin"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func getNewest(lastId string, size int64, listType string, metaid string, followed string) (list []*Tweet, total int64, err error) {
	filter := bson.D{}
	totalFilter := bson.D{}
	if lastId != "" {
		var objectId primitive.ObjectID
		objectId, err = primitive.ObjectIDFromHex(lastId)
		if err != nil {
			return
		}
		filter = append(filter, bson.E{Key: "_id", Value: bson.D{{Key: "$lt", Value: objectId}}})
	}
	if metaid != "" && followed == "1" {
		followList, err1 := getAddressFollowing(metaid)
		if err1 != nil || len(followList) == 0 {
			err = nil
			return
		}
		totalFilter = append(totalFilter, bson.E{Key: "createmetaid", Value: bson.D{{Key: "$in", Value: followList}}})
		filter = append(filter, bson.E{Key: "createmetaid", Value: bson.D{{Key: "$in", Value: followList}}})
	} else if metaid != "" && followed == "" {
		filter = append(filter, bson.E{Key: "createmetaid", Value: metaid})
		totalFilter = append(totalFilter, bson.E{Key: "createmetaid", Value: metaid})
	}
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: listType, Value: -1}})
	findOptions.SetLimit(size)
	result, err := mongoClient.Collection(BuzzView).Find(context.TODO(), filter, findOptions)
	if err != nil {
		return
	}
	err = result.All(context.TODO(), &list)
	if err == mongo.ErrNoDocuments {
		err = nil
	}
	for _, item := range list {
		item.Content = string(item.ContentBody)
		item.ContentBody = nil
	}
	total, err = mongoClient.Collection(BuzzView).CountDocuments(context.TODO(), totalFilter)
	return
}
func getAddressFollowing(metaid string) (list []string, err error) {
	filterA := bson.M{"followmetaid": metaid, "status": true}
	result, err := mongoClient.Collection(mongodb.FollowCollection).Find(context.TODO(), filterA)
	if err != nil {
		return
	}
	var followData []*pin.FollowData //pin.FollowData
	err = result.All(context.TODO(), &followData)
	for _, item := range followData {
		list = append(list, item.MetaId)
	}
	return
}
func getInfo(pinId string) (tweet *Tweet, comments []*TweetComment, like []*TweetLike, err error) {
	filter := bson.D{{Key: "id", Value: pinId}}
	err = mongoClient.Collection(BuzzView).FindOne(context.TODO(), filter, nil).Decode(&tweet)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err = nil
		}
		return
	}
	tweet.Content = string(tweet.ContentBody)
	tweet.ContentBody = nil
	filter2 := bson.D{{Key: "commentpinid", Value: pinId}}
	result, err := mongoClient.Collection(TweetCommentCollection).Find(context.TODO(), filter2)
	if err == nil {
		result.All(context.TODO(), &comments)
	}
	filter3 := bson.D{{Key: "liketopinid", Value: pinId}}
	result2, err := mongoClient.Collection(TweetLikeCollection).Find(context.TODO(), filter3)
	if err == nil {
		result2.All(context.TODO(), &like)
	}
	return
}
func getBlockInfo(height int64, host string, cursor int64, size int64, orderby string) (list []*HostData, err error) {
	var filter primitive.D
	if height > 0 {
		filter = bson.D{{Key: "blockHeight", Value: height}}
	} else {
		filter = bson.D{{Key: "host", Value: host}}
	}
	if orderby == "" {
		orderby = "txCount"
	}
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: orderby, Value: -1}})
	findOptions.SetSkip(cursor).SetLimit(size)
	result, err := mongoClient.Collection(HostDataCollection).Find(context.TODO(), filter, findOptions)
	if err != nil {
		return
	}
	err = result.All(context.TODO(), &list)
	if err == mongo.ErrNoDocuments {
		err = nil
	}
	return
}
