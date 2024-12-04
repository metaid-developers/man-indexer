package metaso

import (
	"context"
	"log"
	"manindexer/common"
	"manindexer/common/mongo_util"
	"manindexer/database/mongodb"
	"reflect"
	"time"

	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Mongodb struct{}

var (
	mongoClient *mongo.Database
)

const (
	TweetCollection         string = "metaso_tweet"
	TweetCountCollection    string = "metaso_tweet_count"
	TweetLikeCollection     string = "metaso_tweet_like"
	TweetCommentCollection  string = "metaso_sync_comment"
	BuzzView                string = "buzzview"
	HostDataCollection      string = "host_data"
	MetasoTickCollection    string = "metaso_tick"
	MetaSoMempoolCollection string = "metaso_mempool"
)

var DataFilter = bson.D{
	{Key: "$or", Value: bson.A{
		bson.D{{Key: "path", Value: "/protocols/simplebuzz"}},
		bson.D{{Key: "path", Value: "/protocols/banana"}},
		bson.D{{Key: "path", Value: "/protocols/paybuzz"}},
		bson.D{{Key: "path", Value: "/protocols/subscribebuzz"}},
	}},
}

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
	createBuzzView()
}
func createIndex(mongoClient *mongo.Database) {
	//Tweet
	mongo_util.CreateIndexIfNotExists(mongoClient, TweetCollection, "pinid_1", bson.D{{Key: "id", Value: 1}}, true)
	mongo_util.CreateIndexIfNotExists(mongoClient, TweetCollection, "output_1", bson.D{{Key: "output", Value: 1}}, false)
	mongo_util.CreateIndexIfNotExists(mongoClient, TweetCollection, "path_1", bson.D{{Key: "path", Value: 1}}, false)
	mongo_util.CreateIndexIfNotExists(mongoClient, TweetCollection, "chainname_1", bson.D{{Key: "chainname", Value: 1}}, false)
	mongo_util.CreateIndexIfNotExists(mongoClient, TweetCollection, "timestamp_1", bson.D{{Key: "timestamp", Value: 1}}, false)
	mongo_util.CreateIndexIfNotExists(mongoClient, TweetCollection, "metaid_1", bson.D{{Key: "metaid", Value: 1}}, false)
	mongo_util.CreateIndexIfNotExists(mongoClient, TweetCollection, "creatormetaid_1", bson.D{{Key: "creatormetaid", Value: 1}}, false)
	mongo_util.CreateIndexIfNotExists(mongoClient, TweetCollection, "number_1", bson.D{{Key: "number", Value: 1}}, false)
	mongo_util.CreateIndexIfNotExists(mongoClient, TweetCollection, "operation_1", bson.D{{Key: "operation", Value: 1}}, false)
	//payLike
	mongo_util.CreateIndexIfNotExists(mongoClient, TweetLikeCollection, "pinid_1", bson.D{{Key: "pinid", Value: 1}}, true)
	mongo_util.CreateIndexIfNotExists(mongoClient, TweetLikeCollection, "liketopinid_1", bson.D{{Key: "liketopinid", Value: 1}}, false)
	mongo_util.CreateIndexIfNotExists(mongoClient, TweetLikeCollection, "createaddress_1", bson.D{{Key: "createaddress", Value: 1}}, false)
	mongo_util.CreateIndexIfNotExists(mongoClient, TweetLikeCollection, "createmetaid_1", bson.D{{Key: "createmetaid", Value: 1}}, false)
	mongo_util.CreateIndexIfNotExists(mongoClient, TweetLikeCollection, "islike_1", bson.D{{Key: "islike", Value: 1}}, false)
	mongo_util.CreateIndexIfNotExists(mongoClient, TweetLikeCollection, "timestamp_1", bson.D{{Key: "timestamp", Value: 1}}, false)
	//comment
	mongo_util.CreateIndexIfNotExists(mongoClient, TweetCommentCollection, "pinid_1", bson.D{{Key: "pinid", Value: 1}}, true)
	mongo_util.CreateIndexIfNotExists(mongoClient, TweetCommentCollection, "commentpinid_1", bson.D{{Key: "commentpinid", Value: 1}}, false)
	mongo_util.CreateIndexIfNotExists(mongoClient, TweetCommentCollection, "createaddress_1", bson.D{{Key: "createaddress", Value: 1}}, false)
	mongo_util.CreateIndexIfNotExists(mongoClient, TweetCommentCollection, "createmetaid_1", bson.D{{Key: "createmetaid", Value: 1}}, false)
	mongo_util.CreateIndexIfNotExists(mongoClient, TweetCommentCollection, "islike_1", bson.D{{Key: "islike", Value: 1}}, false)
	mongo_util.CreateIndexIfNotExists(mongoClient, TweetCommentCollection, "timestamp_1", bson.D{{Key: "timestamp", Value: 1}}, false)
	//hostData
	mongo_util.CreateIndexIfNotExists(mongoClient, HostDataCollection, "host_height_1", bson.D{{Key: "host", Value: 1}, {Key: "blockHeight", Value: 1}}, true)
	mongo_util.CreateIndexIfNotExists(mongoClient, HostDataCollection, "host_1", bson.D{{Key: "host", Value: 1}}, false)
	mongo_util.CreateIndexIfNotExists(mongoClient, HostDataCollection, "height_1", bson.D{{Key: "blockHeight", Value: 1}}, false)
	//MetasoTickCollection
	mongo_util.CreateIndexWithFilterIfNotExists(mongoClient, MetasoTickCollection, "idcoin_address_1", bson.D{{Key: "address", Value: 1}, {Key: "idcoin", Value: 1}}, true, bson.D{{Key: "idcoin", Value: 1}})
	mongo_util.CreateIndexIfNotExists(mongoClient, MetasoTickCollection, "address_1", bson.D{{Key: "address", Value: 1}}, false)
	mongo_util.CreateIndexIfNotExists(mongoClient, MetasoTickCollection, "idconin_1", bson.D{{Key: "idconin", Value: 1}}, false)
	//MetaSoMempoolCollection
	mongo_util.CreateIndexIfNotExists(mongoClient, MetaSoMempoolCollection, "pinid_1", bson.D{{Key: "pinid", Value: 1}}, true)
	mongo_util.CreateIndexIfNotExists(mongoClient, MetaSoMempoolCollection, "target_1", bson.D{{Key: "target", Value: 1}}, false)
}
func createBuzzView() {
	views, err := mongoClient.ListCollectionNames(context.Background(), bson.M{"name": BuzzView})
	if err != nil {
		return
	}
	if len(views) == 0 {
		mongoClient.CreateView(
			context.Background(),
			BuzzView,
			TweetCollection,
			bson.A{
				bson.D{{Key: "$unionWith", Value: bson.D{
					{Key: "coll", Value: mongodb.MempoolPinsCollection},
					{Key: "pipeline", Value: mongo.Pipeline{
						{{Key: "$match", Value: DataFilter}},
					}},
				}}},
			},
		)
	}
}
