package mongodb

import (
	"context"
	"manindexer/pin"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (mg *Mongodb) BatchUpsertFollowData(followData []*pin.FollowData) (err error) {
	var followList []*pin.FollowData
	var unFollowList []*pin.FollowData
	for _, follow := range followData {
		if follow.Status {
			followList = append(followList, follow)
		} else {
			unFollowList = append(unFollowList, follow)
		}
	}
	var followModels []mongo.WriteModel
	for _, info := range followList {
		filter := bson.D{{Key: "metaid", Value: info.MetaId}}
		var updateInfo bson.D
		updateInfo = append(updateInfo, bson.E{Key: "followmetaid", Value: info.FollowMetaId})
		updateInfo = append(updateInfo, bson.E{Key: "followpinid", Value: info.FollowPinId})
		updateInfo = append(updateInfo, bson.E{Key: "followtime", Value: info.FollowTime})
		updateInfo = append(updateInfo, bson.E{Key: "status", Value: info.Status})
		update := bson.D{{Key: "$set", Value: updateInfo}}
		m := mongo.NewUpdateOneModel()
		m.SetFilter(filter).SetUpdate(update).SetUpsert(true)
		followModels = append(followModels, m)
	}

	var unFollowModels []mongo.WriteModel
	for _, info := range unFollowList {
		filter := bson.D{{Key: "followpinid", Value: info.FollowPinId}}
		var updateInfo bson.D
		updateInfo = append(updateInfo, bson.E{Key: "unfollowpinid", Value: info.UnFollowPinId})
		updateInfo = append(updateInfo, bson.E{Key: "status", Value: info.Status})
		update := bson.D{{Key: "$set", Value: updateInfo}}
		m := mongo.NewUpdateOneModel()
		m.SetFilter(filter).SetUpdate(update).SetUpsert(true)
		unFollowModels = append(unFollowModels, m)
	}

	bulkWriteOptions := options.BulkWrite().SetOrdered(false)
	_, err1 := mongoClient.Collection(FollowCollection).BulkWrite(context.Background(), followModels, bulkWriteOptions)

	_, err2 := mongoClient.Collection(FollowCollection).BulkWrite(context.Background(), unFollowModels, bulkWriteOptions)
	if err1 != nil {
		err = err1
		return
	}
	if err2 != nil {
		err = err2
		return
	}
	return
}
func (mg *Mongodb) GetFollowDataByMetaId(metaId string) (followData []*pin.FollowData, err error) {
	return
}
