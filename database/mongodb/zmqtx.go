package mongodb

import (
	"context"
	"log"
	"manindexer/pin"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (mg *Mongodb) SaveZmqReciveTx(data *pin.ZmqReciveTx) (err error) {
	_, err = mongoClient.Collection(ZmqReciveTx).InsertOne(context.TODO(), data)
	return
}
func (mg *Mongodb) GetOneZmqTx(tx string) (data *pin.ZmqReciveTx, err error) {
	filter := bson.M{"tx": tx}
	err = mongoClient.Collection(ZmqReciveTx).FindOne(context.TODO(), filter).Decode(&data)
	if err != nil && err != mongo.ErrNoDocuments {
		log.Println("GetOneZmqTx err", err)
	}
	return
}
func (mg *Mongodb) DeleteZmqTx(txList []string) (err error) {
	filter := bson.M{"tx": bson.M{"$in": txList}}
	_, err = mongoClient.Collection(ZmqReciveTx).DeleteMany(context.TODO(), filter)
	if err != nil {
		log.Println("DeleteMempoolInscription err", err)
	}
	return
}
