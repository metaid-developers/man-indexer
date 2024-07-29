package mongodb

import (
	"context"
	"manindexer/mrc721"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (mg *Mongodb) SaveMrc721Collection(collection *mrc721.Mrc721CollectionDescPin) (err error) {
	_, err = mongoClient.Collection(Mrc721Collection).InsertOne(context.TODO(), collection)
	return
}
func (mg *Mongodb) GetMrc721Collection(collectionName, pinId string) (data *mrc721.Mrc721CollectionDescPin, err error) {
	if collectionName == "" && pinId == "" {
		return
	}
	filter := bson.D{}
	if collectionName != "" {
		filter = append(filter, bson.E{Key: "collectionname", Value: collectionName})
	}
	if pinId != "" {
		filter = append(filter, bson.E{Key: "pinid", Value: pinId})
	}
	err = mongoClient.Collection(Mrc721Collection).FindOne(context.TODO(), filter).Decode(&data)
	if err == mongo.ErrNoDocuments {
		err = nil
	}
	return
}
func (mg *Mongodb) GetMrc721CollectionList(nameList []string, cnt bool) (data []*mrc721.Mrc721CollectionDescPin, total int64, err error) {
	filter := bson.D{}
	if len(nameList) > 0 {
		filter = append(filter, bson.E{Key: "collectionname", Value: bson.M{"$in": nameList}})
	}
	result, err := mongoClient.Collection(Mrc721Collection).Find(context.TODO(), filter)
	if err != nil {
		return
	}
	err = result.All(context.TODO(), &data)
	if cnt {
		total, err = mongoClient.Collection(Mrc721Collection).CountDocuments(context.TODO(), filter)
	}
	return
}

func (mg *Mongodb) BatchUpdateMrc721CollectionCount(nameList []string) (err error) {
	groupFilter := bson.M{"collectionname": bson.M{"$in": nameList}}
	pipelineCount := bson.A{
		bson.D{{Key: "$match", Value: groupFilter}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$collectionname"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}
	cursorb, err := mongoClient.Collection(Mrc721Item).Aggregate(context.Background(), pipelineCount)
	if err != nil {
		return
	}
	defer cursorb.Close(context.Background())
	var results2 []bson.M
	if err = cursorb.All(context.Background(), &results2); err != nil {
		return
	}
	var models []mongo.WriteModel
	for _, item := range results2 {
		filter := bson.D{{Key: "collectionname", Value: item["_id"]}}
		var updateInfo bson.D
		cnt := item["count"].(int32)
		if cnt > 0 {
			updateInfo = append(updateInfo, bson.E{Key: "totalnum", Value: cnt})
		}
		update := bson.D{{Key: "$set", Value: updateInfo}}
		m := mongo.NewUpdateOneModel()
		m.SetFilter(filter).SetUpdate(update)
		models = append(models, m)
	}
	bulkWriteOptions := options.BulkWrite().SetOrdered(false)
	_, err = mongoClient.Collection(Mrc721Collection).BulkWrite(context.Background(), models, bulkWriteOptions)

	return
}
func (mg *Mongodb) SaveMrc721Item(itemList []*mrc721.Mrc721ItemDescPin) (err error) {
	ordered := false
	option := options.InsertManyOptions{Ordered: &ordered}
	var data []interface{}
	for _, item := range itemList {
		data = append(data, item)
	}
	_, err = mongoClient.Collection(Mrc721Item).InsertMany(context.TODO(), data, &option)
	return
}
func (mg *Mongodb) GetMrc721ItemList(collectionName string, pinIdList []string, cnt bool) (itemList []*mrc721.Mrc721ItemDescPin, total int64, err error) {
	if collectionName == "" {
		return
	}
	filter := bson.D{
		bson.E{Key: "collectionname", Value: collectionName},
	}
	if len(pinIdList) > 0 {
		filter = append(filter, bson.E{Key: "itempinid", Value: bson.M{"$in": pinIdList}})
	}

	result, err := mongoClient.Collection(Mrc721Item).Find(context.TODO(), filter)
	if err != nil {
		return
	}
	err = result.All(context.TODO(), &itemList)
	if cnt {
		total, err = mongoClient.Collection(Mrc721Item).CountDocuments(context.TODO(), filter)
	}
	return
}
func (mg *Mongodb) UpdateMrc721ItemDesc(itemList []*mrc721.Mrc721ItemDescPin) (err error) {
	var models []mongo.WriteModel
	for _, item := range itemList {
		filter := bson.D{{Key: "itempinid", Value: item.ItemPinId}, {Key: "descadded", Value: false}}
		var updateInfo bson.D
		if item.Name != "" {
			updateInfo = append(updateInfo, bson.E{Key: "name", Value: item.Name})
		}
		if item.Desc != "" {
			updateInfo = append(updateInfo, bson.E{Key: "desc", Value: item.Desc})
		}
		if item.Cover != "" {
			updateInfo = append(updateInfo, bson.E{Key: "cover", Value: item.Cover})
		}
		if item.Metadata != "" {
			updateInfo = append(updateInfo, bson.E{Key: "metadata", Value: item.Metadata})
		}
		updateInfo = append(updateInfo, bson.E{Key: "descadded", Value: true})
		update := bson.D{{Key: "$set", Value: updateInfo}}
		m := mongo.NewUpdateOneModel()
		m.SetFilter(filter).SetUpdate(update)
		models = append(models, m)
	}
	bulkWriteOptions := options.BulkWrite().SetOrdered(false)
	_, err = mongoClient.Collection(Mrc721Item).BulkWrite(context.Background(), models, bulkWriteOptions)

	return
}
