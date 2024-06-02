package mongodb

import (
	"context"
	"manindexer/mrc20"

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
func (mg *Mongodb) GetMrc20TickPageList(page int64, size int64) (list []mrc20.Mrc20DeployInfo, err error) {
	cursor := (page - 1) * size
	opts := options.Find().SetSort(bson.D{{Key: "pinnumber", Value: -1}}).SetSkip(cursor).SetLimit(size)
	result, err := mongoClient.Collection(Mrc20TickCollection).Find(context.TODO(), bson.M{}, opts)
	if err != nil {
		return
	}
	err = result.All(context.TODO(), &list)
	return
}
func (mg *Mongodb) AddMrc20Shovel(shovel string, pinId string) (err error) {
	d := mrc20.Mrc20Shovel{Shovel: shovel, UsePinId: pinId}
	_, err = mongoClient.Collection(Mrc20MintShovel).InsertOne(context.TODO(), d)
	return
}
func (mg *Mongodb) GetMrc20Shovel(shovels []string) (data map[string]mrc20.Mrc20Shovel, err error) {
	filter := bson.M{"shovel": bson.M{"$in": shovels}}
	result, err := mongoClient.Collection(Mrc20MintShovel).Find(context.TODO(), filter)
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
		data[item.Shovel] = item
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
func (mg *Mongodb) GetMrc20HistoryPageList(mrc20Id string, page int64, size int64) (list []mrc20.Mrc20Utxo, err error) {
	cursor := (page - 1) * size
	opts := options.Find().SetSort(bson.D{{Key: "blockheight", Value: -1}}).SetSkip(cursor).SetLimit(size)
	filter := bson.M{"mrc20id": mrc20Id}
	result, err := mongoClient.Collection(Mrc20UtxoCollection).Find(context.TODO(), filter, opts)
	if err != nil {
		return
	}
	err = result.All(context.TODO(), &list)
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
		filter := bson.D{{Key: "txpoint", Value: info.TxPoint}, {Key: "mrc20id", Value: info.Mrc20Id}}
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
