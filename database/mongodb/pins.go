package mongodb

import (
	"context"
	"encoding/json"
	"log"
	"manindexer/pin"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (mg *Mongodb) GetMaxHeight(chainName string) (height int64, err error) {
	filter := bson.M{"chainname": chainName}
	findOp := options.FindOne()
	findOp.SetSort(bson.D{{Key: "genesisheight", Value: -1}})
	var pinInscription pin.PinInscription
	err = mongoClient.Collection(PinsCollection).FindOne(context.TODO(), filter, findOp).Decode(&pinInscription)
	if err != nil && err == mongo.ErrNoDocuments {
		err = nil
		return
	}
	if pinInscription.GenesisHeight > 1 {
		height = pinInscription.GenesisHeight
	}
	return
}

func (mg *Mongodb) GetMaxNumber() (number int64) {
	findOp := options.FindOne()
	findOp.SetSort(bson.D{{Key: "number", Value: -1}})
	var pinInscription pin.PinInscription
	err := mongoClient.Collection(PinsCollection).FindOne(context.TODO(), bson.D{}, findOp).Decode(&pinInscription)
	if err != nil && err == mongo.ErrNoDocuments {
		err = nil
		return
	}
	number = pinInscription.Number + 1
	return
}

func (mg *Mongodb) BatchAddPins(pins []interface{}) (err error) {
	ordered := false
	option := options.InsertManyOptions{Ordered: &ordered}
	_, err = mongoClient.Collection(PinsCollection).InsertMany(context.TODO(), pins, &option)
	if err != nil {
		return
	}
	//add PDV & FDV
	addPDV(pins)
	addFDV(pins)
	return
}

func (mg *Mongodb) UpdateTransferPin(trasferMap map[string]*pin.PinTransferInfo) (err error) {
	var models []mongo.WriteModel
	for id, info := range trasferMap {
		filter := bson.D{{Key: "output", Value: id}}
		var updateInfo bson.D
		updateInfo = append(updateInfo, bson.E{Key: "istransfered", Value: true})
		updateInfo = append(updateInfo, bson.E{Key: "address", Value: info.Address})
		updateInfo = append(updateInfo, bson.E{Key: "location", Value: info.Location})
		updateInfo = append(updateInfo, bson.E{Key: "offset", Value: info.Offset})
		updateInfo = append(updateInfo, bson.E{Key: "output", Value: info.Output})
		updateInfo = append(updateInfo, bson.E{Key: "outputvalue", Value: info.OutputValue})
		update := bson.D{{Key: "$set", Value: updateInfo}}
		m := mongo.NewUpdateOneModel()
		m.SetFilter(filter).SetUpdate(update)
		models = append(models, m)
	}
	bulkWriteOptions := options.BulkWrite().SetOrdered(false)
	_, err = mongoClient.Collection(PinsCollection).BulkWrite(context.Background(), models, bulkWriteOptions)

	return
}
func (mg *Mongodb) BatchUpdatePins(pins []*pin.PinInscription) (err error) {
	var models []mongo.WriteModel
	for _, pin := range pins {
		if pin.OriginalId == "" {
			continue
		}
		filter := bson.D{{Key: "id", Value: pin.OriginalId}, {Key: "address", Value: pin.Address}}
		var updateInfo bson.D
		if pin.Status != 0 {
			updateInfo = append(updateInfo, bson.E{Key: "status", Value: pin.Status})
		}
		update := bson.D{{Key: "$set", Value: updateInfo}}
		m := mongo.NewUpdateOneModel()
		m.SetFilter(filter).SetUpdate(update)
		models = append(models, m)
	}
	bulkWriteOptions := options.BulkWrite().SetOrdered(false)
	_, err = mongoClient.Collection(PinsCollection).BulkWrite(context.Background(), models, bulkWriteOptions)

	return
}
func (mg *Mongodb) AddMempoolPin(pin *pin.PinInscription) (err error) {
	_, err = mongoClient.Collection(MempoolPinsCollection).InsertOne(context.TODO(), pin)
	return
}
func (mg *Mongodb) GetPinPageList(page int64, size int64) (pins []*pin.PinInscription, err error) {
	cursor := (page - 1) * size
	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}}).SetSkip(cursor).SetLimit(size)
	result, err := mongoClient.Collection(PinsCollection).Find(context.TODO(), bson.M{}, opts)
	if err != nil {
		return
	}
	err = result.All(context.TODO(), &pins)
	return
}

func (mg *Mongodb) GetPinListByIdList(idList []string) (pinList []*pin.PinInscription, err error) {
	filter := bson.M{"id": bson.M{"$in": idList}}
	result, err := mongoClient.Collection(PinsCollection).Find(context.TODO(), filter)
	if err != nil {
		return
	}
	err = result.All(context.TODO(), &pinList)
	return
}
func (mg *Mongodb) GetPinListByOutPutList(outputList []string) (pinList []*pin.PinInscription, err error) {
	filter := bson.M{"output": bson.M{"$in": outputList}}
	result, err := mongoClient.Collection(PinsCollection).Find(context.TODO(), filter)
	if err != nil {
		return
	}
	err = result.All(context.TODO(), &pinList)
	return
}
func (mg *Mongodb) GetMempoolPinPageList(page int64, size int64) (pins []*pin.PinInscription, err error) {
	cursor := (page - 1) * size
	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}, {Key: "number", Value: -1}}).SetSkip(cursor).SetLimit(size)
	result, err := mongoClient.Collection(MempoolPinsCollection).Find(context.TODO(), bson.M{}, opts)
	if err != nil {
		return
	}
	err = result.All(context.TODO(), &pins)
	return
}
func (mg *Mongodb) DeleteMempoolInscription(txIds []string) (err error) {
	filter := bson.M{"id": bson.M{"$in": txIds}}
	_, err = mongoClient.Collection(MempoolPinsCollection).DeleteMany(context.TODO(), filter)
	if err != nil {
		log.Println("DeleteMempoolInscription err", err)
	}
	var ts []string
	for _, id := range txIds {
		index := strings.LastIndex(id, "i")
		if index <= 0 {
			continue
		}
		ts = append(ts, id[:index])
	}
	filter2 := bson.M{"txhash": bson.M{"$in": ts}}
	_, err = mongoClient.Collection(MempoolTransferPinsCollection).DeleteMany(context.TODO(), filter2)
	if err != nil {
		log.Println("DeleteMempoolTransfer err", err)
	}
	return
}
func (mg *Mongodb) GetPinListByAddress(address string, addressType string, cursor int64, size int64, cnt string, path string) (pins []*pin.PinInscription, total int64, err error) {
	opts := options.Find().SetSort(bson.D{{Key: "number", Value: -1}}).SetSkip(cursor).SetLimit(size)
	addStr := "address"
	if addressType == "creator" {
		addStr = "createaddress"
	}
	// filter := bson.M{addStr: address, "status": 0}
	// if path != "" {
	// 	filter = bson.M{addStr: address, "status": 0, "originalpath": path}
	// }
	filter := bson.D{
		{Key: addStr, Value: address},
		{Key: "status", Value: 0},
		{Key: "operation", Value: bson.D{
			{Key: "$ne", Value: "hide"},
		}},
	}
	if path != "" {
		filter = append(filter, bson.E{Key: "originalpath", Value: path})
	}
	result, err := mongoClient.Collection(PinsCollection).Find(context.TODO(), filter, opts)
	if err != nil {
		return
	}
	err = result.All(context.TODO(), &pins)
	if cnt == "true" {
		total, err = mongoClient.Collection(PinsCollection).CountDocuments(context.TODO(), filter)
	}
	return
}
func (mg *Mongodb) GetPinUtxoCountByAddress(address string) (utxoNum int64, utxoSum int64, err error) {
	filter := bson.D{{Key: "address", Value: address}, {Key: "status", Value: 0}}
	match := bson.D{{Key: "$match", Value: filter}}
	groupStage := bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "utxo_sum", Value: bson.D{{Key: "$sum", Value: "$outputvalue"}}},
			{Key: "utxo_num", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}}
	cursor, err := mongoClient.Collection(PinsCollection).Aggregate(context.TODO(), mongo.Pipeline{match, groupStage})
	if err != nil {
		return
	}
	var results []bson.M
	if err = cursor.All(context.TODO(), &results); err != nil {
		return
	}
	for _, result := range results {
		utxoNum += int64(result["utxo_num"].(int32))
		utxoSum += result["utxo_sum"].(int64)
	}
	return
}

func (mg *Mongodb) GetPinByNumberOrId(numberOrId string) (pinInscription *pin.PinInscription, err error) {
	number, err1 := strconv.ParseInt(numberOrId, 10, 64)
	if err1 == nil {
		err = mongoClient.Collection(PinsCollection).FindOne(context.TODO(), bson.D{{Key: "number", Value: number}}).Decode(&pinInscription)
	} else {
		err = mongoClient.Collection(PinsCollection).FindOne(context.TODO(), bson.D{{Key: "id", Value: numberOrId}}).Decode(&pinInscription)
	}
	if err == mongo.ErrNoDocuments {
		pinInscription, err = mg.GetMemPoolPinByNumberOrId(numberOrId)
	}
	if err == mongo.ErrNoDocuments {
		err = nil
	}
	return
}
func (mg *Mongodb) GetPinByMeatIdOrId(key string) (pinInscription *pin.PinInscription, err error) {
	err = mongoClient.Collection(PinsCollection).FindOne(context.TODO(), bson.M{"$or": bson.A{bson.M{"id": key}, bson.M{"metaid": key}, bson.M{"genesistransaction": key}}}).Decode(&pinInscription)
	return
}
func (mg *Mongodb) GetPinByOutput(output string) (pinInscription *pin.PinInscription, err error) {
	err = mongoClient.Collection(PinsCollection).FindOne(context.TODO(), bson.D{{Key: "output", Value: output}}).Decode(&pinInscription)
	return
}
func (mg *Mongodb) GetMemPoolPinByNumberOrId(numberOrId string) (pinInscription *pin.PinInscription, err error) {
	number, err1 := strconv.ParseInt(numberOrId, 10, 64)
	if err1 == nil {
		err = mongoClient.Collection(MempoolPinsCollection).FindOne(context.TODO(), bson.D{{Key: "number", Value: number}}).Decode(&pinInscription)
	} else {
		err = mongoClient.Collection(MempoolPinsCollection).FindOne(context.TODO(), bson.D{{Key: "id", Value: numberOrId}}).Decode(&pinInscription)
	}
	if err == mongo.ErrNoDocuments {
		err = nil
	}
	return
}

func (mg *Mongodb) GetBlockPin(height int64, size int64) (pins []*pin.PinInscription, total int64, err error) {
	filter := bson.D{{Key: "genesisheight", Value: height}}
	opts := options.Find().SetSort(bson.D{{Key: "number", Value: -1}}).SetLimit(size)
	result, err := mongoClient.Collection(PinsCollection).Find(context.TODO(), filter, opts)
	if err != nil {
		return
	}
	err = result.All(context.TODO(), &pins)
	if err != nil {
		return
	}
	total, err = mongoClient.Collection(PinsCollection).CountDocuments(context.TODO(), filter)
	return
}

func (mg *Mongodb) GetMetaIdPin(address string, page int64, size int64) (pins []*pin.PinInscription, total int64, err error) {
	cursor := (page - 1) * size
	filter := bson.D{{Key: "address", Value: address}}
	opts := options.Find().SetSort(bson.D{{Key: "number", Value: -1}}).SetSkip(cursor).SetLimit(size)
	result, err := mongoClient.Collection(PinsCollection).Find(context.TODO(), filter, opts)
	if err != nil {
		return
	}
	err = result.All(context.TODO(), &pins)
	if err != nil {
		return
	}
	total, err = mongoClient.Collection(PinsCollection).CountDocuments(context.TODO(), filter)
	return
}
func (mg *Mongodb) GetChildNodeById(pinId string) (pins []*pin.PinInscription, err error) {
	var p *pin.PinInscription
	err = mongoClient.Collection(PinsCollection).FindOne(context.TODO(), bson.M{"id": pinId}).Decode(&p)
	if err != nil {
		return
	}
	filter := bson.D{{Key: "parentpath", Value: p.Path}}
	opts := options.Find().SetSort(bson.D{{Key: "number", Value: -1}})
	result, err := mongoClient.Collection(PinsCollection).Find(context.TODO(), filter, opts)
	if err != nil {
		return
	}
	err = result.All(context.TODO(), &pins)
	return
}
func (mg *Mongodb) GetParentNodeById(pinId string) (pinnode *pin.PinInscription, err error) {
	var p *pin.PinInscription
	err = mongoClient.Collection(PinsCollection).FindOne(context.TODO(), bson.M{"id": pinId}).Decode(&p)
	if err != nil {
		return
	}
	err = mongoClient.Collection(PinsCollection).FindOne(context.TODO(), bson.M{"metaid": p.MetaId, "path": p.ParentPath}).Decode(&p)
	if err != nil {
		return
	}
	return
}
func (mg *Mongodb) GetAllPinByPath(page, limit int64, path string, metaidList []string) (pins []*pin.PinInscription, total int64, err error) {
	pathList := strings.Split(path, ",")
	filter := bson.M{"path": bson.M{"$in": pathList}}
	if len(metaidList) > 0 {
		filter = bson.M{"path": bson.M{"$in": pathList}, "metaid": bson.M{"$in": metaidList}}
	}
	cursor := (page - 1) * limit
	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}, {Key: "number", Value: -1}}).SetSkip(cursor).SetLimit(limit)
	mempoolResult, err := mongoClient.Collection(MempoolPinsCollection).Find(context.TODO(), filter, opts)
	if err != nil && err != mongo.ErrNoDocuments {
		return
	}
	var memPins []*pin.PinInscription
	var blockPins []*pin.PinInscription
	if mempoolResult != nil {
		err = mempoolResult.All(context.TODO(), &memPins)
		if err != nil {
			return
		}
	}
	newLimit := limit - int64(len(memPins))
	if newLimit > 0 {
		opts = options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}, {Key: "number", Value: -1}}).SetSkip(cursor).SetLimit(newLimit)
		result, err1 := mongoClient.Collection(PinsCollection).Find(context.TODO(), filter, opts)
		if err1 != nil {
			return
		}
		err = result.All(context.TODO(), &blockPins)
		if err != nil {
			return
		}
	}
	var blockTotal int64
	var memTotal int64
	blockTotal, err = mongoClient.Collection(PinsCollection).CountDocuments(context.TODO(), filter)
	memTotal, err = mongoClient.Collection(MempoolPinsCollection).CountDocuments(context.TODO(), filter)
	total = blockTotal + memTotal
	pins = append(pins, memPins...)
	pins = append(pins, blockPins...)
	return
}
func (mg *Mongodb) BatchAddProtocolData(pins []*pin.PinInscription) (err error) {
	dataMap := make(map[string][]*pin.PinInscription)
	for _, pinItem := range pins {
		keyArr := strings.Split(pinItem.Path, "/")
		key := keyArr[len(keyArr)-1]
		if list, ok := dataMap[key]; ok {
			dataMap[key] = append(list, pinItem)
		} else {
			dataMap[key] = []*pin.PinInscription{pinItem}
		}
	}
	//ordered := false
	//option := options.InsertManyOptions{Ordered: &ordered}
	for collectionName, pinList := range dataMap {
		data := getDataByContent(pinList)
		if len(data) > 0 {
			upsertProtocolData(data, collectionName)
			//mongoClient.Collection(collectionName).InsertMany(context.TODO(), data, &option)
		}
	}
	return
}
func upsertProtocolData(data []map[string]interface{}, collectionName string) (err error) {
	var models []mongo.WriteModel
	for _, info := range data {
		filter := bson.D{{Key: "pinId", Value: info["pinId"]}}
		var updateInfo bson.D
		for k, v := range info {
			updateInfo = append(updateInfo, bson.E{Key: k, Value: v})
		}
		update := bson.D{{Key: "$set", Value: updateInfo}}
		m := mongo.NewUpdateOneModel()
		m.SetFilter(filter).SetUpdate(update).SetUpsert(true)
		models = append(models, m)
	}
	_, err = mongoClient.Collection(collectionName).BulkWrite(context.Background(), models)
	return
}
func getDataByContent(pinList []*pin.PinInscription) (data []map[string]interface{}) {
	for _, pinItem := range pinList {
		var d map[string]interface{}
		err := json.Unmarshal(pinItem.ContentBody, &d)
		if err == nil {
			d["pinId"] = pinItem.Id
			d["pinNumber"] = pinItem.Number
			d["pinAddress"] = pinItem.Address
			data = append(data, d)
		} else {
			//fmt.Println(err)
		}
	}
	return
}
func (mg *Mongodb) AddMempoolTransfer(transferData *pin.MemPoolTrasferPin) (err error) {
	_, err = mongoClient.Collection(MempoolTransferPinsCollection).InsertOne(context.TODO(), transferData)
	return
}
func (mg *Mongodb) GetMempoolTransfer(address string, act string) (list []*pin.MemPoolTrasferPin, err error) {
	filter := bson.M{"$or": bson.A{bson.M{"toaddress": address}, bson.M{"fromaddress": address}}}
	result, err := mongoClient.Collection(MempoolTransferPinsCollection).Find(context.TODO(), filter)
	if err != nil {
		return
	}
	err = result.All(context.TODO(), &list)
	return
}
func (mg *Mongodb) GetMempoolTransferById(pinId string) (result *pin.MemPoolTrasferPin, err error) {
	err = mongoClient.Collection(MempoolTransferPinsCollection).FindOne(context.TODO(), bson.M{"pinid": pinId}).Decode(&result)
	return
}
