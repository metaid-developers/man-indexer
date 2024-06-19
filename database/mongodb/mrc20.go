package mongodb

import (
	"context"
	"fmt"
	"manindexer/mrc20"
	"manindexer/pin"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (mg *Mongodb) GetMrc20TickInfo(mrc20Id string) (info mrc20.Mrc20DeployInfo, err error) {
	err = mongoClient.Collection(Mrc20TickCollection).FindOne(context.TODO(), bson.M{"mrc20id": mrc20Id}).Decode(&info)
	if err == mongo.ErrNoDocuments {
		err = nil
	}
	return
}

func (mg *Mongodb) SaveMrc20Pin(data []mrc20.Mrc20Utxo) (err error) {
	var list []interface{}
	for _, item := range data {
		list = append(list, item)
	}
	ordered := false
	option := options.InsertManyOptions{Ordered: &ordered}
	_, err = mongoClient.Collection(Mrc20UtxoCollection).InsertMany(context.TODO(), list, &option)
	return
}
func (mg *Mongodb) SaveMrc20Tick(data []mrc20.Mrc20DeployInfo) (err error) {
	var list []interface{}
	for _, item := range data {
		list = append(list, item)
	}
	ordered := false
	option := options.InsertManyOptions{Ordered: &ordered}
	_, err = mongoClient.Collection(Mrc20TickCollection).InsertMany(context.TODO(), list, &option)
	return
}
func (mg *Mongodb) GetMrc20TickPageList(cursor int64, size int64, order string) (total int64, list []mrc20.Mrc20DeployInfo, err error) {
	//cursor := (page - 1) * size
	if order == "" {
		order = "pinnumber"
	}
	opts := options.Find().SetSort(bson.D{{Key: order, Value: -1}}).SetSkip(cursor).SetLimit(size)
	result, err := mongoClient.Collection(Mrc20TickCollection).Find(context.TODO(), bson.M{}, opts)
	if err != nil {
		return
	}
	err = result.All(context.TODO(), &list)
	if err != nil {
		return
	}
	total, err = mongoClient.Collection(Mrc20TickCollection).CountDocuments(context.TODO(), bson.M{})
	return
}
func (mg *Mongodb) AddMrc20Shovel(shovelList []string, pinId string) (err error) {
	var models []mongo.WriteModel
	for _, id := range shovelList {
		filter := bson.D{{Key: "id", Value: id}}
		var updateInfo bson.D
		updateInfo = append(updateInfo, bson.E{Key: "mrc20minted", Value: true})
		updateInfo = append(updateInfo, bson.E{Key: "mrc20mintpin", Value: pinId})
		update := bson.D{{Key: "$set", Value: updateInfo}}
		m := mongo.NewUpdateOneModel()
		m.SetFilter(filter).SetUpdate(update)
		models = append(models, m)
	}
	bulkWriteOptions := options.BulkWrite().SetOrdered(false)
	_, err = mongoClient.Collection(PinsCollection).BulkWrite(context.Background(), models, bulkWriteOptions)

	return

	// var list []interface{}
	// for _, s := range shovelList {
	// 	list = append(list, mrc20.Mrc20Shovel{Shovel: s, UsePinId: pinId})
	// }
	// ordered := false
	// option := options.InsertManyOptions{Ordered: &ordered}
	// _, err = mongoClient.Collection(Mrc20MintShovel).InsertMany(context.TODO(), list, &option)
	// return
}
func (mg *Mongodb) GetMrc20Shovel(shovels []string) (data map[string]mrc20.Mrc20Shovel, err error) {
	filter := bson.M{"id": bson.M{"$in": shovels}, "mrc20minted": true}
	result, err := mongoClient.Collection(PinsCollection).Find(context.TODO(), filter)
	data = make(map[string]mrc20.Mrc20Shovel)
	if err != nil {
		return
	}
	var list []mrc20.Mrc20Shovel
	err = result.All(context.TODO(), &list)
	if err != nil {
		return
	}
	for _, item := range list {
		data[item.Id] = item
	}
	return
}
func (mg *Mongodb) UpdateMrc20TickInfo(mrc20Id string, minted int64) (err error) {
	filter := bson.M{"mrc20id": mrc20Id}
	update := bson.M{"totalminted": minted}
	_, err = mongoClient.Collection(Mrc20TickCollection).UpdateOne(context.Background(), filter, bson.M{"$set": update})
	return
}
func (mg *Mongodb) GetMrc20ByAddressAndTick(address string, mrc20Id string) (list []mrc20.Mrc20Utxo, err error) {
	filter := bson.M{"mrc20id": mrc20Id, "toaddress": address, "verify": true}
	result, err := mongoClient.Collection(Mrc20UtxoCollection).Find(context.TODO(), filter)
	if err != nil {
		return
	}
	err = result.All(context.TODO(), &list)
	return
}
func (mg *Mongodb) GetMrc20HistoryPageList(tickId string, isPage bool, page int64, size int64) (list []mrc20.Mrc20Utxo, total int64, err error) {
	cursor := page
	if isPage {
		cursor = (page - 1) * size
	}
	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}}).SetSkip(cursor).SetLimit(size)
	filter := bson.M{"mrc20id": tickId}
	result, err := mongoClient.Collection(Mrc20UtxoCollection).Find(context.TODO(), filter, opts)
	if err != nil {
		return
	}
	err = result.All(context.TODO(), &list)
	if err != nil {
		return
	}
	total, err = mongoClient.Collection(Mrc20UtxoCollection).CountDocuments(context.TODO(), filter)
	return
}
func (mg *Mongodb) GetMrc20UtxoByOutPutList(outputList []string) (list []*mrc20.Mrc20Utxo, err error) {
	filter := bson.M{"txpoint": bson.M{"$in": outputList}, "status": 0, "verify": true}
	result, err := mongoClient.Collection(Mrc20UtxoCollection).Find(context.TODO(), filter, nil)
	if err != nil {
		return
	}
	err = result.All(context.TODO(), &list)
	return
}
func (mg *Mongodb) UpdateMrc20Utxo(list []*mrc20.Mrc20Utxo) (err error) {
	var models []mongo.WriteModel
	for _, info := range list {
		filter := bson.D{{Key: "txpoint", Value: info.TxPoint}, {Key: "index", Value: info.Index}, {Key: "mrc20id", Value: info.Mrc20Id}, {Key: "verify", Value: info.Verify}}
		var updateInfo bson.D
		if info.Status == -1 {
			updateInfo = append(updateInfo, bson.E{Key: "status", Value: -1})
		} else {
			updateInfo = append(updateInfo, bson.E{Key: "amtchange", Value: info.AmtChange})
			updateInfo = append(updateInfo, bson.E{Key: "blockheight", Value: info.BlockHeight})
			updateInfo = append(updateInfo, bson.E{Key: "errormsg", Value: info.ErrorMsg})
			updateInfo = append(updateInfo, bson.E{Key: "fromaddress", Value: info.FromAddress})
			updateInfo = append(updateInfo, bson.E{Key: "mrc20id", Value: info.Mrc20Id})
			updateInfo = append(updateInfo, bson.E{Key: "mrcoption", Value: info.MrcOption})
			updateInfo = append(updateInfo, bson.E{Key: "status", Value: info.Status})
			updateInfo = append(updateInfo, bson.E{Key: "tick", Value: info.Tick})
			updateInfo = append(updateInfo, bson.E{Key: "toaddress", Value: info.ToAddress})
			updateInfo = append(updateInfo, bson.E{Key: "txpoint", Value: info.TxPoint})
			updateInfo = append(updateInfo, bson.E{Key: "verify", Value: info.Verify})
			updateInfo = append(updateInfo, bson.E{Key: "chain", Value: info.Chain})
			updateInfo = append(updateInfo, bson.E{Key: "index", Value: info.Index})
			updateInfo = append(updateInfo, bson.E{Key: "timestamp", Value: info.Timestamp})
		}
		update := bson.D{{Key: "$set", Value: updateInfo}}
		m := mongo.NewUpdateOneModel()
		m.SetFilter(filter).SetUpdate(update).SetUpsert(true)
		models = append(models, m)
	}
	bulkWriteOptions := options.BulkWrite().SetOrdered(false)
	_, err = mongoClient.Collection(Mrc20UtxoCollection).BulkWrite(context.Background(), models, bulkWriteOptions)
	return
}
func (mg *Mongodb) GetHistoryByAddress(tickId string, address string, cursor int64, size int64) (list []mrc20.Mrc20Utxo, total int64, err error) {
	//cursor := (page - 1) * size
	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}}).SetSkip(cursor).SetLimit(size)
	filter := bson.M{"mrc20id": tickId, "toaddress": address}
	result, err := mongoClient.Collection(Mrc20UtxoCollection).Find(context.TODO(), filter, opts)
	if err != nil {
		return
	}
	err = result.All(context.TODO(), &list)
	if err != nil {
		return
	}
	total, err = mongoClient.Collection(Mrc20UtxoCollection).CountDocuments(context.TODO(), filter)
	return
}
func (mg *Mongodb) GetMrc20BalanceByAddress(address string) (list []mrc20.Mrc20Balance, total int64, err error) {
	filter := bson.D{{Key: "toaddress", Value: address}}
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$mrc20id"},
			{Key: "total", Value: bson.D{{Key: "$sum", Value: "$amtchange"}}},
		}}},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cursor, err := mongoClient.Collection(Mrc20UtxoCollection).Aggregate(ctx, pipeline)
	if err != nil {
		return
	}
	defer cursor.Close(ctx)
	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return
	}
	var idList []string
	for _, result := range results {
		idList = append(idList, result["_id"].(string))
		b := mrc20.Mrc20Balance{Id: result["_id"].(string), Balance: result["total"].(int64)}
		//fmt.Printf("Category: %v, Total: %v\n", result["_id"], result["total"])
		list = append(list, b)
	}
	tickFilter := bson.M{"mrc20id": bson.M{"$in": idList}}
	ret, err := mongoClient.Collection(Mrc20TickCollection).Find(context.TODO(), tickFilter)
	var tickList []mrc20.Mrc20DeployInfo
	if err = ret.All(ctx, &tickList); err != nil {
		return
	}
	m := make(map[string]string)
	for _, tick := range tickList {
		m[tick.Mrc20Id] = tick.Tick
	}
	for i := range list {
		if v, ok := m[list[i].Id]; ok {
			list[i].Name = v
		}
	}
	total = int64(len(list))
	return
}
func (mg *Mongodb) GetHistoryByTx(txId string, index int64, cursor int64, size int64) (list []mrc20.Mrc20Utxo, total int64, err error) {
	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}}).SetSkip(cursor).SetLimit(size)
	txpoint := fmt.Sprintf("%s:%d", txId, index)
	filter := bson.M{"txpoint": txpoint}
	result, err := mongoClient.Collection(Mrc20UtxoCollection).Find(context.TODO(), filter, opts)
	if err != nil {
		return
	}
	err = result.All(context.TODO(), &list)
	if err != nil {
		return
	}
	total, err = mongoClient.Collection(Mrc20UtxoCollection).CountDocuments(context.TODO(), filter)
	return
}
func (mg *Mongodb) GetShovelListByAddress(address string, lv int, path, query, key, operator, value string, cursor int64, size int64) (list []*pin.PinInscription, total int64, err error) {
	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}}).SetSkip(cursor).SetLimit(size)
	//filter := bson.M{"txpoint": txpoint}
	filter := bson.D{{Key: "address", Value: address}, {Key: "mrc20minted", Value: false}}
	if lv > 0 {
		filter = append(filter, bson.E{Key: "poplv", Value: bson.D{{Key: "$gte", Value: lv}}})
	}
	if key != "" && operator != "" && value != "" {
		protocols := strings.ReplaceAll(path, "/protocols", "")
		idList, err1 := getPinIdInProtocols(strings.ToLower(protocols), key, value)
		if err1 != nil || len(idList) <= 0 {
			err = err1
			return
		}
		filter = append(filter, bson.E{Key: "id", Value: bson.E{Key: "$in", Value: idList}})
	} else if path == "/follow" && query != "" {
		pinId, err1 := getFollowPinId(query, address)
		if err1 != nil || pinId == "" {
			err = err1
			return
		}
		filter = append(filter, bson.E{Key: "id", Value: pinId})
	} else if path != "" {
		filter = append(filter, bson.E{Key: "path", Value: path})
	}
	result, err := mongoClient.Collection(PinsCollection).Find(context.TODO(), filter, opts)
	if err != nil {
		return
	}
	err = result.All(context.TODO(), &list)
	if err != nil {
		return
	}
	total, err = mongoClient.Collection(PinsCollection).CountDocuments(context.TODO(), filter)
	return
}
func getPinIdInProtocols(protocols string, key string, value string) (idList []string, err error) {
	filter := bson.M{key: value}
	result, err := mongoClient.Collection(protocols).Find(context.TODO(), filter)
	if err != nil {
		return
	}
	var list []map[string]interface{}
	err = result.All(context.TODO(), &list)
	if err != nil {
		return
	}
	for _, item := range list {
		idList = append(idList, item["pinId"].(string))
	}
	return
}
func getFollowPinId(metaid string, address string) (pinId string, err error) {
	filter := bson.M{"metaid": metaid, "followmetaid": address}
	var f pin.FollowData
	err = mongoClient.Collection(FollowCollection).FindOne(context.TODO(), filter).Decode(&f)
	if err != nil {
		return
	}
	pinId = f.FollowPinId
	return
}
