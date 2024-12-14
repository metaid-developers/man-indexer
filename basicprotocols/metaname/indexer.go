package metaname

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"manindexer/common"
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
		metaName.syncTransfer()
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
		{Key: "originalpath", Value: bson.D{{Key: "$regex", Value: "^/metaname"}}},
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
	var updateDocs []*MetaNameData
	var lastId primitive.ObjectID
	for _, doc := range pinList {
		if mongodb.CompareObjectIDs(doc.MogoID, lastId) > 0 {
			lastId = doc.MogoID
		}
		data, history, err := validator(doc)
		//fmt.Println(doc.OriginalPath, string(doc.ContentBody), err)
		if data == nil && err != nil {
			continue
		}
		if history.Op == "reg" {
			insertDocs = append(insertDocs, data)
		} else if history.Op == "modify" {
			updateDocs = append(updateDocs, data)
		}
		historyDocs = append(historyDocs, history)
	}
	mongodb.UpdateSyncLastIdLog("metaname", lastId)
	insertOpts := options.InsertMany().SetOrdered(false)
	if len(insertDocs) > 0 {
		mongoClient.Collection(MeatNameCollection).InsertMany(context.TODO(), insertDocs, insertOpts)
	}
	if len(updateDocs) > 0 {
		updateMetaName(updateDocs)
	}
	if len(historyDocs) > 0 {
		mongoClient.Collection(MeatNameHistoryCollection).InsertMany(context.TODO(), historyDocs, insertOpts)
	}
	return
}

func validator(pinNode *MetaNamePin) (data *MetaNameData, history *MetaNameHistory, err error) {
	pathArr := strings.Split(pinNode.OriginalPath, "/")
	if len(pathArr) != 3 {
		err = errors.New("PIN path error")
		return
	}
	space := strings.ToLower(pathArr[2])
	pattern := `^[^\s.\n]*$`
	re := regexp.MustCompile(pattern)
	if !re.MatchString(space) {
		err = errors.New("PIN path space error")
		return
	}
	var nameJson MetaNameProtocol
	content := string(pinNode.ContentBody)
	err = json.Unmarshal([]byte(content), &nameJson)
	if err != nil {
		return
	}
	nameJson.Name = strings.ToLower(nameJson.Name)
	if !re.MatchString(nameJson.Name) {
		err = errors.New("name error")
		return
	}
	op := "reg"
	if pinNode.Operation == "create" {
		finded, err1 := checkNameExits(nameJson.Name)
		if finded || err1 != nil {
			err = errors.New("name check error")
			return
		}
	} else {
		op = "modify"
		pinId := strings.Replace(pinNode.Path, "@", "", -1)
		data, err := findNameData(pinId)
		if err == nil && data != nil {
			nameJson.Name = data.Name
			space = data.Space
		}
	}
	fullName := fmt.Sprintf("%s.%s", nameJson.Name, space)
	data = &MetaNameData{
		Name:     nameJson.Name,
		Space:    space,
		FullName: fullName,
		Rev:      nameJson.Rev,
		Relay:    nameJson.Relay,
		PinId:    pinNode.Id,
		Address:  pinNode.Address,
		MetaId:   pinNode.MetaId,
		Metadata: nameJson.Metadata,
	}
	history = &MetaNameHistory{
		Name:      nameJson.Name,
		Space:     space,
		FullName:  fullName,
		Op:        op,
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
func findNameData(pinId string) (data *MetaNameData, err error) {
	filer := bson.D{{Key: "pinid", Value: pinId}}
	findOp := options.FindOne()
	err = mongoClient.Collection(MeatNameCollection).FindOne(context.TODO(), filer, findOp).Decode(&data)
	if err == mongo.ErrNoDocuments {
		err = nil
	}
	return
}
func updateMetaName(list []*MetaNameData) (err error) {
	var models []mongo.WriteModel
	for _, data := range list {
		filter := bson.D{{Key: "name", Value: data.Name}, {Key: "space", Value: data.Space}}
		var updateInfo bson.D
		if data.Rev != "" {
			updateInfo = append(updateInfo, bson.E{Key: "rev", Value: data.Rev})
		}
		if data.Relay != "" {
			updateInfo = append(updateInfo, bson.E{Key: "relay", Value: data.Relay})
		}
		if data.Metadata != "" {
			updateInfo = append(updateInfo, bson.E{Key: "metadata", Value: data.Metadata})
		}
		update := bson.D{{Key: "$set", Value: updateInfo}}
		m := mongo.NewUpdateOneModel()
		m.SetFilter(filter).SetUpdate(update).SetUpsert(false)
		models = append(models, m)
	}
	bulkWriteOptions := options.BulkWrite().SetOrdered(false)
	_, err = mongoClient.Collection(MeatNameCollection).BulkWrite(context.Background(), models, bulkWriteOptions)
	return
}

func (metaName *MetaName) syncTransfer() (err error) {
	last, err := mongodb.GetSyncLastId("metanametransfer")
	if err != nil {
		return
	}
	var dataList []*PinTransferHistory
	filter := bson.D{}
	if last != primitive.NilObjectID {
		filter = append(filter, bson.E{Key: "_id", Value: bson.D{{Key: "$gt", Value: last}}})
	}
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "_id", Value: 1}})
	findOptions.SetLimit(500)
	result, err := mongoClient.Collection(mongodb.PinTransferHistory).Find(context.TODO(), filter)
	if err != nil {
		return
	}
	result.All(context.TODO(), &dataList)
	if len(dataList) <= 0 {
		return
	}
	var lastId primitive.ObjectID
	//mrc721.PinHandle(pinList)
	transferMap := make(map[string]string)
	var pinList []string
	for _, doc := range dataList {
		if mongodb.CompareObjectIDs(doc.MogoID, lastId) > 0 {
			lastId = doc.MogoID
		}
		transferMap[doc.PinId] = doc.ToAddress
		pinList = append(pinList, doc.PinId)
	}
	err = transferItemHandle(pinList, transferMap)
	if err == nil {
		mongodb.UpdateSyncLastIdLog("metanametransfer", lastId)
	}

	return
}

func transferItemHandle(pinIdList []string, transferMap map[string]string) (err error) {
	filter := bson.M{"pinid": bson.M{"$in": pinIdList}}
	result, err := mongoClient.Collection(MeatNameCollection).Find(context.TODO(), filter)
	if err != nil {
		return
	}
	var itemList []*MetaNameData
	err = result.All(context.TODO(), &itemList)
	if err != nil || len(itemList) <= 0 {
		return
	}
	var models []mongo.WriteModel
	for _, item := range itemList {
		filter := bson.D{{Key: "pinid", Value: item.PinId}}
		var updateInfo bson.D
		newAddress := transferMap[item.PinId]
		updateInfo = append(updateInfo, bson.E{Key: "address", Value: newAddress})
		updateInfo = append(updateInfo, bson.E{Key: "metaid", Value: common.GetMetaIdByAddress(newAddress)})
		update := bson.D{{Key: "$set", Value: updateInfo}}
		m := mongo.NewUpdateOneModel()
		m.SetFilter(filter).SetUpdate(update)
		models = append(models, m)
	}
	bulkWriteOptions := options.BulkWrite().SetOrdered(false)
	_, err = mongoClient.Collection(MeatNameCollection).BulkWrite(context.Background(), models, bulkWriteOptions)
	return
}
