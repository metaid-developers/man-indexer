package mongodb

import (
	"context"
	"encoding/json"
	"manindexer/basicprotocols/metaaccess"
	"manindexer/pin"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (mg *Mongodb) BatchSaveAccesscontrol(list []*metaaccess.AccessControl) (err error) {
	// ordered := false
	// option := options.InsertManyOptions{Ordered: &ordered}
	// data := make([]interface{}, len(list))
	// for i, item := range list {
	// 	data[i] = item
	// }
	// _, err = mongoClient.Collection(AccessControlCollection).InsertMany(context.TODO(), data, &option)

	var models []mongo.WriteModel
	for _, acc := range list {
		filter := bson.D{{Key: "pinid", Value: acc.PinId}}
		update := bson.D{{Key: "$set", Value: acc}}
		m := mongo.NewUpdateOneModel()
		m.SetFilter(filter).SetUpdate(update).SetUpsert(true)
		models = append(models, m)
	}
	bulkWriteOptions := options.BulkWrite().SetOrdered(false)
	_, err = mongoClient.Collection(AccessControlCollection).BulkWrite(context.Background(), models, bulkWriteOptions)

	return
}
func (mg *Mongodb) GetControlById(pinId string, isContentId bool) (data *metaaccess.AccessControl, err error) {
	findOp := options.FindOne().SetSort(bson.D{{Key: "_id", Value: -1}})
	filter := bson.D{{Key: "pinid", Value: pinId}}
	if isContentId {
		filter = bson.D{
			{Key: "controlpins", Value: bson.D{{Key: "$in", Value: bson.A{pinId}}}},
		}
	}
	data = &metaaccess.AccessControl{}
	err = mongoClient.Collection(AccessControlCollection).FindOne(context.TODO(), filter, findOp).Decode(data)
	return
}
func (mg *Mongodb) BatchSaveAccessPass(passList []*metaaccess.AccessPassData) (err error) {

	ordered := false
	option := options.InsertManyOptions{Ordered: &ordered}
	data := make([]interface{}, len(passList))
	for i, item := range passList {
		data[i] = item
	}
	_, err = mongoClient.Collection(AccessPassCollection).InsertMany(context.TODO(), data, &option)
	return
}
func (mg *Mongodb) CheckAccessPass(buyerAddress string, contentPinId string, controlPath string) (data *metaaccess.AccessPassData, err error) {
	filer := bson.D{{Key: "buyeraddress", Value: buyerAddress}, {Key: "contentpinid", Value: contentPinId}}
	findOp := options.FindOne()
	data = &metaaccess.AccessPassData{}
	err = mongoClient.Collection(AccessPassCollection).FindOne(context.TODO(), filer, findOp).Decode(data)
	if err == nil {
		return
	}
	if controlPath == "" {
		return
	}
	filer2 := bson.D{{Key: "buyeraddress", Value: buyerAddress}, {Key: "controlpath", Value: controlPath}}
	err = mongoClient.Collection(AccessPassCollection).FindOne(context.TODO(), filer2, findOp).Decode(data)
	if err == nil {
		//TODO ValidPeriod check
		return
	}
	return
}

func CheckAccessPassInMempool(buyerAddress string, controlPinId string) (inMempool bool, err error) {
	filer2 := bson.D{{Key: "address", Value: buyerAddress}, {Key: "path", Value: "/metaaccess/accesspass"}}
	var pins []*pin.PinInscription
	var result *mongo.Cursor
	result, err = mongoClient.Collection(MempoolPinsCollection).Find(context.TODO(), filer2)
	if err != nil {
		return
	}
	err = result.All(context.TODO(), &pins)
	if err != nil {
		return
	}
	for _, pinNode := range pins {
		var pass metaaccess.AccessPass
		err = json.Unmarshal(pinNode.ContentBody, &pass)
		if err != nil {
			continue
		}
		if pass.AccessControlID == controlPinId {
			inMempool = true
			break
		}
	}
	return
}
