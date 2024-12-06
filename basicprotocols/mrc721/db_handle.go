package mrc721

import (
	"context"
	"log"
	"manindexer/common"
	"manindexer/common/mongo_util"
	"reflect"
	"time"

	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	mongoClient *mongo.Database
)

const (
	Mrc721Collection string = "mrc721collection"
	Mrc721Item       string = "mrc721item"
)

func connectMongoDb() {
	mg := common.Config.MongoDb
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(mg.TimeOut))
	defer cancel()
	o := options.Client().ApplyURI(mg.MongoURI)
	o.SetMaxPoolSize(uint64(mg.PoolSize))
	o.SetRegistry(bson.NewRegistryBuilder().
		RegisterDecoder(reflect.TypeOf(decimal.Decimal{}), mongo_util.Decimal{}).
		RegisterEncoder(reflect.TypeOf(decimal.Decimal{}), mongo_util.Decimal{}).
		Build())
	client, err := mongo.Connect(ctx, o)
	if err != nil {
		log.Panic("ConnectToDB", err)
		return
	}
	if err = client.Ping(context.Background(), readpref.Primary()); err != nil {
		log.Panic("ConnectToDB", err)
		return
	}
	mongoClient = client.Database(mg.DbName)
	createIndex(mongoClient)
}
func createIndex(mongoClient *mongo.Database) {
	//mrc721
	mongo_util.CreateIndexIfNotExists(mongoClient, Mrc721Collection, "collectionname_1", bson.D{{Key: "collectionname", Value: 1}}, true)
	mongo_util.CreateIndexIfNotExists(mongoClient, Mrc721Collection, "pinid_1", bson.D{{Key: "pinid", Value: 1}}, true)
	mongo_util.CreateIndexIfNotExists(mongoClient, Mrc721Item, "itempinid_1", bson.D{{Key: "itempinid", Value: 1}}, true)
	mongo_util.CreateIndexIfNotExists(mongoClient, Mrc721Item, "collectionname_1", bson.D{{Key: "collectionname", Value: 1}}, false)
	mongo_util.CreateIndexIfNotExists(mongoClient, Mrc721Item, "collectionname_itempinid_1", bson.D{{Key: "collectionname", Value: 1}, {Key: "itempinid", Value: 1}}, false)
	mongo_util.CreateIndexIfNotExists(mongoClient, Mrc721Item, "itempinid_descadded_1", bson.D{{Key: "itempinid", Value: 1}, {Key: "descadded", Value: 1}}, false)

}
