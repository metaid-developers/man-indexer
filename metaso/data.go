package metaso

import (
	"bytes"
	"context"
	"manindexer/database/mongodb"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	mongoClient *mongo.Database
)

const (
	TweetCollection        string = "metaso_tweet"
	TweetCountCollection   string = "metaso_tweet_count"
	TweetLikeCollection    string = "metaso_tweet_like"
	SyncLogCollection      string = "metaso_sync_log"
	TweetCommentCollection string = "metaso_sync_comment"
)

type Mongodb struct{}

func (metaso *MetaSo) Synchronization() {
	connectMongoDb()
	for {
		metaso.synchTweet()
		metaso.synchTweetLike()
		metaso.synchTweetComment()
		time.Sleep(time.Second * 10)
	}
}
func compareObjectIDs(id1, id2 primitive.ObjectID) int {
	return bytes.Compare(id1[:], id2[:])
}

func (metaso *MetaSo) synchTweet() (err error) {
	last, err := metaso.getLastId()
	if err != nil {
		return
	}
	var pinList []*Tweet
	filter := bson.D{
		{Key: "path", Value: "/protocols/simplebuzz"},
	}
	if last.Tweet != primitive.NilObjectID {
		filter = append(filter, bson.E{Key: "_id", Value: bson.D{{Key: "$gt", Value: last.Tweet}}})
	}
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "_id", Value: 1}})
	findOptions.SetLimit(500)
	result, err := mongoClient.Collection(mongodb.PinsCollection).Find(context.TODO(), filter)
	if err != nil {
		return
	}
	result.All(context.TODO(), &pinList)
	if len(pinList) <= 0 {
		return
	}

	var insertDocs []interface{}
	var lastId primitive.ObjectID
	for _, doc := range pinList {
		insertDocs = append(insertDocs, doc)
		if compareObjectIDs(doc.MogoID, lastId) > 0 {
			lastId = doc.MogoID
		}
	}
	insertOpts := options.InsertMany().SetOrdered(false)
	_, err1 := mongoClient.Collection(TweetCollection).InsertMany(context.TODO(), insertDocs, insertOpts)
	if err1 != nil {
		err = err1
		return
	}
	metaso.updateSyncLog("tweet", lastId)
	return
}
func (metaso *MetaSo) updateSyncLog(db string, id primitive.ObjectID) (err error) {
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: db, Value: id},
		}},
	}
	updateOpts := options.Update().SetUpsert(true)
	_, err = mongoClient.Collection(SyncLogCollection).UpdateOne(context.TODO(), bson.D{}, update, updateOpts)

	return
}

func (metaso *MetaSo) getLastId() (last SyncLastId, err error) {
	err = mongoClient.Collection(SyncLogCollection).FindOne(context.TODO(), bson.D{}, nil).Decode(&last)
	if err == mongo.ErrNoDocuments {
		err = nil
	}
	return
}
