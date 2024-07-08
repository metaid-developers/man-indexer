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
	followMap := make(map[string]int)
	for _, follow := range followData {
		if follow.Status {
			followList = append(followList, follow)
			followMap[follow.MetaId] += 1
			go addFollowFDV(follow.MetaId, follow.FollowMetaId, "follow")
		} else {
			unFollowList = append(unFollowList, follow)
			followMap[follow.MetaId] -= 1
			go addFollowFDV(follow.MetaId, follow.FollowMetaId, "unfollow")
		}
	}
	go batchUpdateFollowCount(followMap)
	var followModels []mongo.WriteModel
	for _, info := range followList {
		filter := bson.D{{Key: "metaid", Value: info.MetaId}, {Key: "followmetaid", Value: info.FollowMetaId}}
		var updateInfo bson.D
		//updateInfo = append(updateInfo, bson.E{Key: "followmetaid", Value: info.FollowMetaId})
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
func (mg *Mongodb) GetFollowDataByMetaId(metaId string, myFollow bool, followDetail bool, cursor int64, size int64) (metaIdList []interface{}, total int64, err error) {
	filterA := bson.M{"metaid": metaId, "status": true}
	if myFollow {
		filterA = bson.M{"followmetaid": metaId, "status": true}
	}
	opts := options.Find().SetSort(bson.D{{Key: "followtime", Value: -1}}).SetSkip(cursor).SetLimit(size)
	total, err = mongoClient.Collection(FollowCollection).CountDocuments(context.TODO(), filterA)
	if err != nil {
		return
	}
	result, err := mongoClient.Collection(FollowCollection).Find(context.TODO(), filterA, opts)
	if err != nil {
		return
	}
	var followData []*pin.FollowData //pin.FollowData
	err = result.All(context.TODO(), &followData)
	if err != nil || len(followData) <= 0 {
		return
	}
	var idList []string
	for _, f := range followData {
		if myFollow {
			if !followDetail {
				metaIdList = append(metaIdList, f.MetaId)
			} else {
				idList = append(idList, f.MetaId)
			}
		} else {
			if !followDetail {
				metaIdList = append(metaIdList, f.FollowMetaId)
			} else {
				idList = append(idList, f.FollowMetaId)
			}

		}
	}
	if !followDetail {
		return
	}

	filter := bson.M{"metaid": bson.M{"$in": idList}}
	find, err := mongoClient.Collection(MetaIdInfoCollection).Find(context.TODO(), filter)
	if err != nil {
		return
	}
	var list []*pin.MetaIdInfo
	err = find.All(context.TODO(), &list)
	if err == nil {
		for _, p := range list {
			metaIdList = append(metaIdList, p)
		}
	}
	return
}
func (mg *Mongodb) GetFollowRecord(metaId string, followMetaId string) (followData pin.FollowData, err error) {
	filter := bson.M{"metaid": metaId, "followmetaid": followMetaId, "status": true}
	err = mongoClient.Collection(FollowCollection).FindOne(context.TODO(), filter).Decode(&followData)
	return
}
