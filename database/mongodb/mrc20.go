package mongodb

import (
	"context"
	"fmt"
	"manindexer/common"
	"manindexer/mrc20"
	"manindexer/pin"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
func (mg *Mongodb) GetMrc20TickPageList(cursor int64, size int64, order string, completed string, orderType string) (total int64, list []mrc20.Mrc20DeployInfo, err error) {
	//cursor := (page - 1) * size
	if order == "" {
		order = "pinnumber"
	}
	filter := bson.M{}
	if completed == "true" {
		filter = bson.M{"chain": "btc", "$expr": bson.M{"$gte": []string{"$totalminted", "$mintcount"}}}
	} else if completed == "false" {
		filter = bson.M{"chain": "btc", "$expr": bson.M{"$gt": []string{"$mintcount", "$totalminted"}}}
	}
	sortNum := -1
	if orderType == "asc" {
		sortNum = 1
	}
	opts := options.Find().SetSort(bson.D{{Key: order, Value: sortNum}}).SetSkip(cursor).SetLimit(size)
	result, err := mongoClient.Collection(Mrc20TickCollection).Find(context.TODO(), filter, opts)
	if err != nil {
		return
	}
	err = result.All(context.TODO(), &list)
	if err != nil {
		return
	}
	total, err = mongoClient.Collection(Mrc20TickCollection).CountDocuments(context.TODO(), filter)
	return
}
func (mg *Mongodb) AddMrc20Shovel(shovelList []string, pinId string, mrc20Id string) (err error) {
	var models []mongo.WriteModel
	for _, id := range shovelList {
		filter := bson.D{{Key: "id", Value: id}}
		var updateInfo bson.D
		//updateInfo = append(updateInfo, bson.E{Key: "mrc20minted", Value: true})
		//updateInfo = append(updateInfo, bson.E{Key: "mrc20mintpin", Value: pinId})
		updateInfo = append(updateInfo, bson.E{Key: "mrc20mintid", Value: mrc20Id})
		update := bson.D{{Key: "$push", Value: updateInfo}}
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
func (mg *Mongodb) GetMrc20Shovel(shovels []string, mrc20Id string) (data map[string]mrc20.Mrc20Shovel, err error) {
	filter := bson.M{"id": bson.M{"$in": shovels}, "mrc20mintid": bson.M{"$in": bson.A{mrc20Id}}}
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
func (mg *Mongodb) UpdateMrc20TickHolder(tickId string, txNum int64) (err error) {
	//get holder count
	filter := bson.M{"mrc20id": tickId}
	holderNum := getHolderCount(tickId)
	update := bson.M{"$set": bson.M{"holders": holderNum}, "$inc": bson.M{"txcount": txNum}}
	_, err = mongoClient.Collection(Mrc20TickCollection).UpdateOne(context.Background(), filter, update)
	return
}
func getHolderCount(tickId string) (count int64) {
	filter := bson.D{{Key: "mrc20id", Value: tickId}}
	match := bson.D{{Key: "$match", Value: filter}}
	project := bson.D{{Key: "$project", Value: bson.D{{Key: "toaddress", Value: true}}}}
	groupStage := bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$toaddress"},
		}}}
	groupStage2 := bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}}
	cursor, err := mongoClient.Collection(Mrc20UtxoCollection).Aggregate(context.TODO(), mongo.Pipeline{match, project, groupStage, groupStage2})
	if err != nil {
		return
	}
	var results []bson.M
	if err = cursor.All(context.TODO(), &results); err != nil {
		return
	}
	if len(results) > 0 {
		count = int64(results[0]["count"].(int32))
		return
	}
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
			//updateInfo = append(updateInfo, bson.E{Key: "amtchange", Value: info.AmtChange})
		} else {
			updateInfo = append(updateInfo, bson.E{Key: "amtchange", Value: info.AmtChange})
			updateInfo = append(updateInfo, bson.E{Key: "blockheight", Value: info.BlockHeight})
			updateInfo = append(updateInfo, bson.E{Key: "msg", Value: info.Msg})
			updateInfo = append(updateInfo, bson.E{Key: "fromaddress", Value: info.FromAddress})
			updateInfo = append(updateInfo, bson.E{Key: "mrc20id", Value: info.Mrc20Id})
			updateInfo = append(updateInfo, bson.E{Key: "mrcoption", Value: info.MrcOption})
			updateInfo = append(updateInfo, bson.E{Key: "status", Value: info.Status})
			updateInfo = append(updateInfo, bson.E{Key: "tick", Value: info.Tick})
			updateInfo = append(updateInfo, bson.E{Key: "toaddress", Value: info.ToAddress})
			updateInfo = append(updateInfo, bson.E{Key: "txpoint", Value: info.TxPoint})
			updateInfo = append(updateInfo, bson.E{Key: "pointvalue", Value: info.PointValue})
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
func (mg *Mongodb) GetHistoryByAddress(tickId string, address string, cursor int64, size int64, status string, verify string) (list []mrc20.Mrc20Utxo, total int64, err error) {
	//cursor := (page - 1) * size
	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}}).SetSkip(cursor).SetLimit(size)
	filter := bson.D{
		{Key: "mrc20id", Value: tickId},
		{Key: "toaddress", Value: address},
		{Key: "amtchange", Value: bson.D{
			{Key: "$gt", Value: 0},
		}},
	}
	if status != "" {
		s, err := strconv.Atoi(status)
		if err == nil {
			filter = append(filter, bson.E{Key: "status", Value: s})
		}
	}
	if verify != "" && (verify == "true" || verify == "false") {
		v := false
		if verify == "true" {
			v = true
		}
		filter = append(filter, bson.E{Key: "verify", Value: v})
	}
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
func (mg *Mongodb) GetMrc20BalanceByAddress(address string, cursor int64, size int64) (list []mrc20.Mrc20Balance, total int64, err error) {
	filter := bson.D{
		{Key: "toaddress", Value: address},
		{Key: "status", Value: 0},
		{Key: "verify", Value: true},
	}
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$mrc20id"},
			{Key: "total", Value: bson.D{{Key: "$sum", Value: "$amtchange"}}},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "timestamp", Value: -1}}}},
		{{Key: "$skip", Value: cursor}},
		{{Key: "$limit", Value: size}},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cursora, err := mongoClient.Collection(Mrc20UtxoCollection).Aggregate(ctx, pipeline)
	if err != nil {
		return
	}
	defer cursora.Close(ctx)
	var results []bson.M
	if err = cursora.All(ctx, &results); err != nil {
		return
	}
	var idList []string
	for _, result := range results {
		idList = append(idList, result["_id"].(string))
		balance := result["total"].(primitive.Decimal128)
		balanceDecimal, _ := decimal.NewFromString(balance.String())
		b := mrc20.Mrc20Balance{Id: result["_id"].(string), Balance: balanceDecimal}
		//fmt.Printf("Category: %v, Total: %v\n", result["_id"], result["total"])
		list = append(list, b)
	}
	if len(idList) <= 0 {
		return
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
	//count
	pipelineCount := bson.A{
		bson.D{{Key: "$match", Value: filter}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$mrc20id"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
		bson.D{{Key: "$count", Value: "total"}},
	}
	cursorb, err := mongoClient.Collection(Mrc20UtxoCollection).Aggregate(ctx, pipelineCount)
	if err != nil {
		return
	}
	defer cursorb.Close(ctx)
	var results2 []bson.M
	if err = cursorb.All(ctx, &results2); err != nil {
		return
	}
	if len(results2) > 0 {
		total = int64(results2[0]["total"].(int32))
	}
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
func (mg *Mongodb) GetShovelListByAddress(address string, mrc20Id string, creator string, lv int, path, query, key, operator, value string, cursor int64, size int64) (list []*pin.PinInscription, total int64, err error) {
	//fmt.Println(lv, path, query, key, operator, value)
	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}}).SetSkip(cursor).SetLimit(size)
	//filter := bson.M{"txpoint": txpoint}
	filter := bson.D{
		{Key: "address", Value: address},
		{Key: "mrc20mintid", Value: bson.D{
			{Key: "$nin", Value: bson.A{mrc20Id}},
		}},
		{Key: "operation", Value: bson.D{
			{Key: "$ne", Value: "hide"},
		}},
	}
	if lv > 0 {
		filter = append(filter, bson.E{Key: "poplv", Value: bson.D{{Key: "$gte", Value: lv}}})
	}
	if creator != "" {
		filter = append(filter, bson.E{Key: "createmetaid", Value: creator})
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
		pathArr := strings.Split(path, "/")
		//Wildcard
		if pathArr[len(pathArr)-1] == "*" {
			path = path[0 : len(path)-2]
			filter = append(filter, bson.E{Key: "path", Value: bson.D{{Key: "$regex", Value: "^" + path}}})
		} else {
			filter = append(filter, bson.E{Key: "path", Value: path})
		}

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
	filter := bson.M{"metaid": metaid, "followmetaid": common.GetMetaIdByAddress(address)}
	var f pin.FollowData
	err = mongoClient.Collection(FollowCollection).FindOne(context.TODO(), filter).Decode(&f)
	if err != nil {
		return
	}
	pinId = f.FollowPinId
	return
}

func (mg *Mongodb) GetUsedShovelIdListByAddress(address string, tickId string, cursor int64, size int64) (list []*string, total int64, err error) {
	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}}).SetSkip(cursor).SetLimit(size)
	projection := bson.D{
		{Key: "_id", Value: 0},
		{Key: "id", Value: 1},
	}
	opts.SetProjection(projection)
	//filter := bson.M{"address": address, "mrc20mintid": tickId}
	filter := bson.M{"mrc20mintid": tickId}
	result, err := mongoClient.Collection(PinsCollection).Find(context.TODO(), filter, opts)
	if err != nil {
		return
	}
	var rr []bson.M
	err = result.All(context.TODO(), &rr)
	if err != nil {
		return
	}
	for _, item := range rr {
		s := item["id"].(string)
		list = append(list, &s)
	}
	total, err = mongoClient.Collection(PinsCollection).CountDocuments(context.TODO(), filter)
	return
}
