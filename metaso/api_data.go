package metaso

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func getNewest(lastId string, size int64, listType string) (list []*Tweet, total int64, err error) {
	filter := bson.D{}
	if lastId != "" {
		var objectId primitive.ObjectID
		objectId, err = primitive.ObjectIDFromHex(lastId)
		if err != nil {
			return
		}
		filter = append(filter, bson.E{Key: "_id", Value: bson.D{{Key: "$lt", Value: objectId}}})
	}
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: listType, Value: -1}})
	findOptions.SetLimit(size)
	result, err := mongoClient.Collection(TweetCollection).Find(context.TODO(), filter, findOptions)
	if err != nil {
		return
	}
	err = result.All(context.TODO(), &list)
	if err == mongo.ErrNoDocuments {
		err = nil
	}
	for _, item := range list {
		item.ContentBody = nil
		item.Content = string(item.ContentBody)
	}
	total, err = mongoClient.Collection(TweetCollection).CountDocuments(context.TODO(), bson.D{})
	return
}

func getInfo(pinId string) (tweet *Tweet, comments []*TweetComment, like []*TweetLike, err error) {
	filter := bson.D{{Key: "id", Value: pinId}}
	err = mongoClient.Collection(TweetCollection).FindOne(context.TODO(), filter, nil).Decode(&tweet)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err = nil
		}
		return
	}
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
