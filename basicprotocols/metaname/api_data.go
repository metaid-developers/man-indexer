package metaname

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func getNewest(lastId string, size int64, listType string) (list []*MetaNameData, total int64, err error) {
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
	result, err := mongoClient.Collection(MeatNameCollection).Find(context.TODO(), filter, findOptions)
	if err != nil {
		return
	}
	err = result.All(context.TODO(), &list)
	if err == mongo.ErrNoDocuments {
		err = nil
	}
	total, err = mongoClient.Collection(MeatNameCollection).CountDocuments(context.TODO(), bson.D{})
	return
}

func getInfo(name string) (info *MetaNameData, history []*MetaNameHistory, err error) {
	filter := bson.D{{Key: "name", Value: name}}
	err = mongoClient.Collection(MeatNameCollection).FindOne(context.TODO(), filter, nil).Decode(&info)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err = nil
		}
		return
	}
	result, err := mongoClient.Collection(MeatNameHistoryCollection).Find(context.TODO(), filter)
	if err == nil {
		result.All(context.TODO(), &history)
	}
	return
}
