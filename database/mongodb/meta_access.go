package mongodb

import (
	"context"
	"manindexer/basicprotocols/metaaccess"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (mg *Mongodb) BatchSaveAccesscontrol(list []*metaaccess.AccessControl) (err error) {
	ordered := false
	option := options.InsertManyOptions{Ordered: &ordered}
	data := make([]interface{}, len(list))
	for i, item := range list {
		data[i] = item
	}
	_, err = mongoClient.Collection(AccessControlCollection).InsertMany(context.TODO(), data, &option)
	return
}
func (mg *Mongodb) GetControlById(pinId string) (data *metaaccess.AccessControl, err error) {
	findOp := options.FindOne()
	data = &metaaccess.AccessControl{}
	err = mongoClient.Collection(AccessControlCollection).FindOne(context.TODO(), bson.D{{Key: "pinid", Value: pinId}}, findOp).Decode(data)
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
