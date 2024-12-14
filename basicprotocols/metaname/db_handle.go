package metaname

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
	MeatNameCollection        string = "metaname"
	MeatNameHistoryCollection string = "metaname_history"
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
	//MeatNameCollection
	mongo_util.CreateIndexIfNotExists(mongoClient, MeatNameCollection, "pinid_1", bson.D{{Key: "pinid", Value: 1}}, true)
	mongo_util.CreateIndexIfNotExists(mongoClient, MeatNameCollection, "name_space_1", bson.D{{Key: "name", Value: 1}, {Key: "space", Value: 1}}, true)
	mongo_util.CreateIndexIfNotExists(mongoClient, MeatNameCollection, "fullname_1", bson.D{{Key: "fullname", Value: 1}}, false)

	mongo_util.CreateIndexIfNotExists(mongoClient, MeatNameHistoryCollection, "oppinid_1", bson.D{{Key: "oppinid", Value: 1}}, true)
	mongo_util.CreateIndexIfNotExists(mongoClient, MeatNameHistoryCollection, "name_space_1", bson.D{{Key: "name", Value: 1}, {Key: "space", Value: 1}}, false)
	mongo_util.CreateIndexIfNotExists(mongoClient, MeatNameHistoryCollection, "fullname_1", bson.D{{Key: "fullname", Value: 1}}, false)
}
