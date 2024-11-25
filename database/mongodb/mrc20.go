package mongodb

import (
	"context"
	"fmt"
	"log"
	"manindexer/common"
	"manindexer/mrc20"
	"manindexer/pin"
	"sort"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (mg *Mongodb) GetMrc20TickInfo(mrc20Id string, tick string) (info mrc20.Mrc20DeployInfo, err error) {
	if mrc20Id == "" && tick == "" {
		return
	}
	filter := bson.D{}
	if mrc20Id != "" {
		filter = append(filter, bson.E{Key: "mrc20id", Value: mrc20Id})
	}
	if tick != "" {
		filter = append(filter, bson.E{Key: "tick", Value: tick})
	}
	err = mongoClient.Collection(Mrc20TickCollection).FindOne(context.TODO(), filter).Decode(&info)
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
func (mg *Mongodb) UpdateMrc20TickInfo(mrc20Id string, txPoint string, minted uint64) (err error) {
	//Check if already counted
	utxoFilter := bson.M{"txpoint": txPoint}
	var utxo mrc20.Mrc20Utxo
	find := mongoClient.Collection(Mrc20UtxoCollection).FindOne(context.TODO(), utxoFilter).Decode(&utxo)
	if find == mongo.ErrNoDocuments {
		filter := bson.M{"mrc20id": mrc20Id}
		update := bson.M{"totalminted": minted}
		_, err = mongoClient.Collection(Mrc20TickCollection).UpdateOne(context.Background(), filter, bson.M{"$set": update})
	}
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
	filter := bson.D{{Key: "mrc20id", Value: tickId}, {Key: "status", Value: 0}, {Key: "verify", Value: true}, {Key: "mrcoption", Value: bson.D{{Key: "$ne", Value: "deploy"}}}}
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
func (mg *Mongodb) GetMrc20UtxoByOutPutList(outputList []string, isMempool bool) (list []*mrc20.Mrc20Utxo, err error) {
	filter := bson.M{"txpoint": bson.M{"$in": outputList}, "status": 0, "verify": true}
	result, err := mongoClient.Collection(Mrc20UtxoCollection).Find(context.TODO(), filter, nil)
	if err != nil && err != mongo.ErrNoDocuments {
		return
	}
	err = result.All(context.TODO(), &list)
	if err != nil {
		return
	}
	if isMempool {
		var list2 []*mrc20.Mrc20Utxo
		result2, err2 := mongoClient.Collection(Mrc20UtxoMempoolCollection).Find(context.TODO(), filter, nil)
		if err2 == nil {
			result2.All(context.TODO(), &list2)
		}
		if len(list2) > 0 {
			list = append(list, list2...)
		}
	}
	return
}
func (mg *Mongodb) UpdateMrc20Utxo(list []*mrc20.Mrc20Utxo, isMempool bool) (err error) {
	var models []mongo.WriteModel
	collection := Mrc20UtxoCollection
	if isMempool {
		collection = Mrc20UtxoMempoolCollection
	}
	for _, info := range list {
		if info.AmtChange.Cmp(decimal.Zero) == -1 {
			continue
		}
		filter := bson.D{{Key: "txpoint", Value: info.TxPoint}, {Key: "index", Value: info.Index}, {Key: "mrc20id", Value: info.Mrc20Id}, {Key: "verify", Value: info.Verify}}
		var updateInfo bson.D
		//if info.Status == -1 {
		//	updateInfo = append(updateInfo, bson.E{Key: "status", Value: -1})
		//} else {
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
		updateInfo = append(updateInfo, bson.E{Key: "operationtx", Value: info.OperationTx})

		//}
		update := bson.D{{Key: "$set", Value: updateInfo}}
		m := mongo.NewUpdateOneModel()
		m.SetFilter(filter).SetUpdate(update).SetUpsert(true)
		models = append(models, m)
	}
	bulkWriteOptions := options.BulkWrite().SetOrdered(false)
	_, err = mongoClient.Collection(collection).BulkWrite(context.Background(), models, bulkWriteOptions)
	return
}
func GetTickBalance(tickId string, address string) (totalAmt decimal.Decimal, err error) {
	totalAmt = decimal.Zero
	filter := bson.D{
		{Key: "tick", Value: strings.ToUpper(tickId)},
		{Key: "toaddress", Value: address},
		{Key: "status", Value: 0},
		{Key: "verify", Value: true},
		{Key: "amtchange", Value: bson.D{
			{Key: "$gt", Value: 0},
		}},
	}
	result, err := mongoClient.Collection(Mrc20UtxoCollection).Find(context.TODO(), filter)
	if err != nil {
		return
	}
	var list []mrc20.Mrc20Utxo
	err = result.All(context.TODO(), &list)
	if err != nil {
		return
	}
	for _, item := range list {
		totalAmt = totalAmt.Add(item.AmtChange)
	}
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
	//if query status is 0 && verify is true,search mempool
	//if status == "0" && verify == "true" {
	mempoolList, mempoolTotal, err1 := mg.GetMempoolHistoryByAddress(tickId, address)
	if err1 != nil {
		return
	}

	if mempoolTotal > 0 {
		total -= int64(len(list))
	}

	if len(mempoolList) > 0 {
		memMap := make(map[string]mrc20.Mrc20Utxo, len(mempoolList))
		for _, item := range mempoolList {
			k := fmt.Sprintf("%s-%d", item.TxPoint, item.Index)
			memMap[k] = item
		}
		var newList []mrc20.Mrc20Utxo
		for _, item := range list {
			// if item.MrcOption == "mint" {
			// 	newList = append(newList, item)
			// 	continue
			// }
			//transfer data
			k := fmt.Sprintf("%s-%d", item.TxPoint, item.Index)
			_, existed := memMap[k]
			if !existed {
				newList = append(newList, item)
			}
		}
		for _, item := range mempoolList {
			if item.Status == -1 {
				continue
			}
			newList = append(newList, item)
		}
		list = newList
		total += int64(len(list))
	}
	//}
	return
}

func (mg *Mongodb) GetMempoolHistoryByAddress(tickId string, address string) (list []mrc20.Mrc20Utxo, total int64, err error) {
	//cursor := (page - 1) * size
	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}})
	filter := bson.D{
		{Key: "mrc20id", Value: tickId},
		{Key: "toaddress", Value: address},
		//{Key: "status", Value: -1}, //add mint,mint status = 0
		{Key: "amtchange", Value: bson.D{
			{Key: "$gt", Value: 0},
		}},
	}
	result, err := mongoClient.Collection(Mrc20UtxoMempoolCollection).Find(context.TODO(), filter, opts)
	if err != nil {
		return
	}
	err = result.All(context.TODO(), &list)
	if err != nil {
		return
	}
	total, err = mongoClient.Collection(Mrc20UtxoMempoolCollection).CountDocuments(context.TODO(), filter)
	return
}

func (mg *Mongodb) GetMrc20BalanceByAddress(address string, cursor int64, size int64) (balanceList []mrc20.Mrc20Balance, total int64, err error) {
	filter := bson.D{
		{Key: "toaddress", Value: address},
		{Key: "status", Value: 0},
		{Key: "verify", Value: true},
		{Key: "mrcoption", Value: bson.D{{Key: "$ne", Value: "deploy"}}},
	}
	opts := options.Find().SetSort(bson.D{{Key: "tick", Value: 1}})
	result, err := mongoClient.Collection(Mrc20UtxoCollection).Find(context.TODO(), filter, opts)
	if err != nil && err != mongo.ErrNoDocuments {
		return
	}
	var list []*mrc20.Mrc20Utxo
	if err != mongo.ErrNoDocuments {
		err = result.All(context.TODO(), &list)
		if err != nil {
			return
		}
	}
	//mempool data
	mempoolData, err := getMempoolMrc20UtxoByAddress(address)
	if err == nil && len(mempoolData) > 0 {
		list = append(list, mempoolData...)
	}
	var nameList []string
	balanceMap := make(map[string]*mrc20.Mrc20Balance)
	for _, utxo := range list {
		if balance, ok := balanceMap[utxo.Tick]; ok {
			if utxo.BlockHeight == -1 {
				balance.UnsafeBalance = balance.UnsafeBalance.Add(utxo.AmtChange)
			} else {
				balance.Balance = balance.Balance.Add(utxo.AmtChange)
			}
		} else {
			balanceMap[utxo.Tick] = &mrc20.Mrc20Balance{
				Id:   utxo.Mrc20Id,
				Name: utxo.Tick,
			}
			if utxo.BlockHeight == -1 {
				balanceMap[utxo.Tick].UnsafeBalance = utxo.AmtChange
			} else {
				balanceMap[utxo.Tick].Balance = utxo.AmtChange
			}
			nameList = append(nameList, utxo.Tick)
			total += 1
		}
	}
	if len(nameList) <= 0 {
		return
	}
	//sort
	sort.Strings(nameList)
	if len(nameList) > int(cursor+size) {
		nameList = nameList[cursor:size]
	}
	for _, name := range nameList {
		if balance, ok := balanceMap[name]; ok {
			balanceList = append(balanceList, *balance)
		}
	}
	return
}
func getMempoolMrc20BalanceByAddress(address string) (balanceMap map[string]*mrc20.Mrc20MempoolBalance, err error) {
	balanceMap = make(map[string]*mrc20.Mrc20MempoolBalance)
	filter := bson.D{
		{Key: "toaddress", Value: address},
		{Key: "verify", Value: true},
	}
	result, err := mongoClient.Collection(Mrc20UtxoMempoolCollection).Find(context.TODO(), filter)
	if err != nil {
		return
	}
	var list []mrc20.Mrc20Utxo
	err = result.All(context.TODO(), &list)
	if len(list) <= 0 {
		return
	}
	for _, utxo := range list {
		if _, ok := balanceMap[utxo.Mrc20Id]; !ok {
			b := mrc20.Mrc20MempoolBalance{Id: utxo.Mrc20Id, Name: utxo.Tick}
			balanceMap[utxo.Mrc20Id] = &b
		}
		if utxo.Status == 0 {
			balanceMap[utxo.Mrc20Id].Recive = balanceMap[utxo.Mrc20Id].Recive.Add(utxo.AmtChange)
		} else if utxo.BlockHeight > 0 && utxo.Status == -1 {
			id := fmt.Sprintf("%s:%d", utxo.TxPoint, utxo.Index)
			balanceMap[utxo.Mrc20Id].SpendUtxo = append(balanceMap[utxo.Mrc20Id].SpendUtxo, id)
		}
	}
	return
}
func getMempoolMrc20UtxoByAddress(address string) (list []*mrc20.Mrc20Utxo, err error) {
	filter := bson.D{
		{Key: "toaddress", Value: address},
		{Key: "verify", Value: true},
	}
	result, err := mongoClient.Collection(Mrc20UtxoMempoolCollection).Find(context.TODO(), filter)
	if err != nil {
		return
	}
	var tmpList []*mrc20.Mrc20Utxo
	err = result.All(context.TODO(), &tmpList)
	if err != nil {
		return
	}

	for _, utxo := range tmpList {
		if utxo.Status == 0 && utxo.BlockHeight == -1 {
			list = append(list, utxo)
		} else if utxo.BlockHeight > 0 && utxo.Status == -1 {
			utxo.AmtChange = utxo.AmtChange.Neg()
			list = append(list, utxo)
		}
	}
	return
}
func (mg *Mongodb) GetHistoryByTx(txId string, index int64, cursor int64, size int64) (list []mrc20.Mrc20Utxo, total int64, err error) {
	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}}).SetSkip(cursor).SetLimit(size)
	txpoint := fmt.Sprintf("%s:%d", txId, index)
	filter := bson.M{"txpoint": txpoint}
	result, err := mongoClient.Collection(Mrc20UtxoView).Find(context.TODO(), filter, opts)
	if err != nil {
		return
	}
	err = result.All(context.TODO(), &list)
	if err != nil {
		return
	}
	total, err = mongoClient.Collection(Mrc20UtxoView).CountDocuments(context.TODO(), filter)
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
	} else if path != "" && path != "/" {
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
func (mg *Mongodb) DeleteMempoolMc20(txIds []string) (err error) {
	filter := bson.M{"operationtx": bson.M{"$in": txIds}}
	_, err = mongoClient.Collection(Mrc20UtxoMempoolCollection).DeleteMany(context.TODO(), filter)
	if err != nil {
		log.Println("DeleteMempoolMc20 err", err)
	}
	return
}
func (mg *Mongodb) CheckOperationtx(operationtx string, isMempool bool) (data *mrc20.Mrc20Utxo, err error) {
	filter := bson.M{"operationtx": operationtx}
	collection := Mrc20UtxoCollection
	if isMempool {
		collection = Mrc20UtxoMempoolCollection
	}
	err = mongoClient.Collection(collection).FindOne(context.TODO(), filter).Decode(&data)
	if err == mongo.ErrNoDocuments {
		err = nil
	}
	if err != nil {
		log.Println("CheckOperationtx err", err)
	}
	return
}
