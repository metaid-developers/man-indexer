package metaname

import (
	"context"
	"encoding/json"
	"errors"
	"manindexer/database/mongodb"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (metaName *MetaName) Synchronization() {
	connectMongoDb()
	for {
		metaName.sync()
		time.Sleep(time.Second * 10)
	}
}
func (metaName *MetaName) sync() (err error) {
	last, err := mongodb.GetSyncLastId("metaname")
	if err != nil {
		return
	}
	var pinList []*MetaNamePin
	filter := bson.D{
		{Key: "path", Value: "/info/metaname"},
	}
	if last != primitive.NilObjectID {
		filter = append(filter, bson.E{Key: "_id", Value: bson.D{{Key: "$gt", Value: last}}})
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
	var historyDocs []interface{}
	var lastId primitive.ObjectID
	for _, doc := range pinList {
		data, history, err := validator(doc)
		if data == nil && err != nil {
			continue
		}
		if history.Op == "reg" {
			insertDocs = append(insertDocs, data)
		}
		historyDocs = append(historyDocs, history)
		if mongodb.CompareObjectIDs(doc.MogoID, lastId) > 0 {
			lastId = doc.MogoID
		}
	}
	insertOpts := options.InsertMany().SetOrdered(false)
	_, err1 := mongoClient.Collection(MeatNameCollection).InsertMany(context.TODO(), insertDocs, insertOpts)
	if err1 != nil {
		err = err1
		return
	}
	mongoClient.Collection(MeatNameHistoryCollection).InsertMany(context.TODO(), historyDocs, insertOpts)
	mongodb.UpdateSyncLastIdLog("metaname", lastId)
	return
}

func validator(pinNode *MetaNamePin) (data *MetaNameData, history *MetaNameHistory, err error) {
	var nameJson MetaNameProtocol
	content := strings.ToLower(string(pinNode.ContentBody))
	err = json.Unmarshal([]byte(content), &nameJson)
	if err != nil {
		return
	}
	pattern := `^[^\s\n.]+\.(metaid)$`
	re, err := regexp.Compile(pattern)
	if err != nil {
		err = errors.New("regexp error")
		return
	}
	if !re.MatchString(nameJson.Name) {
		err = errors.New("regexp match error")
		return
	}
	if nameJson.Op == "reg" {
		finded, err1 := checkNameExits(nameJson.Name)
		if finded || err1 != nil {
			err = errors.New("name check error")
			return
		}
	}
	data = &MetaNameData{
		Name:   nameJson.Name,
		Avatar: nameJson.Avatar,
		Rev:    nameJson.Rev,
		Relay:  nameJson.Relay,
		PinId:  pinNode.Id,
	}
	history = &MetaNameHistory{
		Name:      nameJson.Name,
		Op:        nameJson.Op,
		OpAddress: pinNode.Address,
		OpMetaId:  pinNode.MetaId,
		Timestamp: pinNode.Timestamp,
		OpPinId:   pinNode.Id,
		OpContent: string(pinNode.ContentBody),
	}
	return
}
func checkNameExits(name string) (finded bool, err error) {
	filer := bson.D{{Key: "name", Value: name}}
	findOp := options.FindOne()
	var data MetaNameData
	err = mongoClient.Collection(MeatNameCollection).FindOne(context.TODO(), filer, findOp).Decode(&data)
	if err == nil {
		finded = true
	}
	if err == mongo.ErrNoDocuments {
		err = nil
	}
	return
}
