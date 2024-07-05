package mongodb

import (
	"context"
	"fmt"
	"log"
	"manindexer/common"
	"manindexer/database"
	"manindexer/pin"
	"reflect"
	"time"

	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	PinsCollection                string = "pins"
	PinsView                      string = "pinsview"
	MempoolPinsCollection         string = "mempoolpins"
	MempoolTransferPinsCollection string = "mempooltransferpins"
	MetaIdInfoCollection          string = "metaid"
	PinTreeCatalogCollection      string = "pintree"
	FollowCollection              string = "follow"
	InfoCollection                string = "info"
	Mrc20UtxoCollection           string = "mrc20utxos"
	Mrc20TickCollection           string = "mrc20ticks"
	//Mrc20MintShovel               string = "mrc20shovel"
)

var (
	mongoClient *mongo.Database
)

type Mongodb struct{}

func (mg *Mongodb) InitDatabase() {
	connectMongoDb()
	protocolsInit(mongoClient)
}
func connectMongoDb() {
	mg := common.Config.MongoDb
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(mg.TimeOut))
	defer cancel()
	o := options.Client().ApplyURI(mg.MongoURI)
	o.SetMaxPoolSize(uint64(mg.PoolSize))
	o.SetRegistry(bson.NewRegistryBuilder().
		RegisterDecoder(reflect.TypeOf(decimal.Decimal{}), Decimal{}).
		RegisterEncoder(reflect.TypeOf(decimal.Decimal{}), Decimal{}).
		Build())
	client, err := mongo.Connect(ctx, o)
	if err != nil {
		log.Panic("ConnectToDB", err)
		return
	}
	if err = client.Ping(context.Background(), readpref.Primary()); err != nil {
		log.Panic("ConnectToDB", err)
		return
	} else {
		log.Println("mongodb connected")
	}
	mongoClient = client.Database(mg.DbName)
	createPinsView()
	createIndexIfNotExists(mongoClient, PinsCollection, "id_1", bson.D{{Key: "id", Value: 1}}, true)
	createIndexIfNotExists(mongoClient, PinsCollection, "output_1", bson.D{{Key: "output", Value: 1}}, false)
	createIndexIfNotExists(mongoClient, PinsCollection, "path_1", bson.D{{Key: "path", Value: 1}}, false)
	createIndexIfNotExists(mongoClient, PinsCollection, "chainname_1", bson.D{{Key: "chainname", Value: 1}}, false)
	createIndexIfNotExists(mongoClient, PinsCollection, "timestamp_1", bson.D{{Key: "timestamp", Value: 1}}, false)
	createIndexIfNotExists(mongoClient, PinsCollection, "metaid_1", bson.D{{Key: "metaid", Value: 1}}, false)
	createIndexIfNotExists(mongoClient, PinsCollection, "creatormetaid_1", bson.D{{Key: "creatormetaid", Value: 1}}, false)
	createIndexIfNotExists(mongoClient, PinsCollection, "number_1", bson.D{{Key: "number", Value: 1}}, false)
	createIndexIfNotExists(mongoClient, PinsCollection, "operation_1", bson.D{{Key: "operation", Value: 1}}, false)
	createIndexIfNotExists(mongoClient, PinsCollection, "address_status_1", bson.D{{Key: "address", Value: 1}, {Key: "status", Value: 1}}, false)

	createIndexIfNotExists(mongoClient, MempoolPinsCollection, "id_1", bson.D{{Key: "id", Value: 1}}, true)

	createIndexIfNotExists(mongoClient, MetaIdInfoCollection, "address_1", bson.D{{Key: "address", Value: 1}}, true)

	createIndexIfNotExists(mongoClient, PinTreeCatalogCollection, "treepath_1", bson.D{{Key: "treepath", Value: 1}}, true)

	createIndexIfNotExists(mongoClient, FollowCollection, "metaid_1", bson.D{{Key: "metaid", Value: 1}}, false)
	createIndexIfNotExists(mongoClient, FollowCollection, "followmetaid_1", bson.D{{Key: "followmetaid", Value: 1}}, false)
	createIndexIfNotExists(mongoClient, FollowCollection, "followpinid_1", bson.D{{Key: "followpinid", Value: 1}}, false)

	createIndexIfNotExists(mongoClient, InfoCollection, "metaid_1", bson.D{{Key: "metaid", Value: 1}}, false)
	createIndexIfNotExists(mongoClient, InfoCollection, "metaid_infokey_1", bson.D{{Key: "metaid", Value: 1}, {Key: "infokey", Value: 1}}, false)

	createIndexIfNotExists(mongoClient, MempoolTransferPinsCollection, "fromaddress_pinid_1", bson.D{{Key: "fromaddress", Value: 1}, {Key: "pinid", Value: 1}}, true)
	createIndexIfNotExists(mongoClient, MempoolTransferPinsCollection, "fromaddress_1", bson.D{{Key: "fromaddress", Value: 1}}, false)
	createIndexIfNotExists(mongoClient, MempoolTransferPinsCollection, "toaddress_1", bson.D{{Key: "toaddress", Value: 1}}, false)
	createIndexIfNotExists(mongoClient, MempoolTransferPinsCollection, "txhash_1", bson.D{{Key: "txhash", Value: 1}}, false)

	//mrc20
	createIndexIfNotExists(mongoClient, Mrc20TickCollection, "mrc20id_1", bson.D{{Key: "mrc20id", Value: 1}}, true)
	createIndexIfNotExists(mongoClient, Mrc20UtxoCollection, "mrc20id_txpoint_verify_1", bson.D{{Key: "mrc20id", Value: 1}, {Key: "txpoint", Value: 1}, {Key: "index", Value: 1}}, true)
}

func (mg *Mongodb) Count() (count pin.PinCount) {
	count = pin.PinCount{}
	count.Pin, _ = mongoClient.Collection(PinsCollection).CountDocuments(context.TODO(), bson.M{})
	gp1 := bson.D{{Key: "$group", Value: bson.D{{Key: "_id", Value: "$genesisheight"}}}}
	data, err := mongoClient.Collection(PinsCollection).Aggregate(context.TODO(), mongo.Pipeline{gp1})
	if err != nil {
		return
	}
	var bc []bson.M
	err = data.All(context.TODO(), &bc)
	if err != nil {
		return
	}
	count.Block = int64(len(bc))
	count.MetaId, _ = mongoClient.Collection(MetaIdInfoCollection).CountDocuments(context.TODO(), bson.M{})
	return
}

func (mg *Mongodb) GeneratorFind(generator database.Generator) (data []map[string]interface{}, err error) {
	var opts *options.FindOptions
	if generator.Action == "get" {
		opts = options.Find()
	}
	if generator.Action == "get" && generator.Limit > 0 {
		opts.SetSkip(generator.Cursor).SetLimit(generator.Limit)
	}
	if generator.Action == "get" && len(generator.Sort) == 2 {
		s := -1
		if generator.Sort[1] == "asc" {
			s = 1
		}
		opts.SetSort(bson.D{{Key: generator.Sort[0], Value: s}})
	}
	var conditions bson.A
	if len(generator.Filters) > 0 {
		for _, f := range generator.Filters {
			conditions = append(conditions, getCondition(f))
		}
	}
	var filter bson.D
	if generator.FilterRelation == "or" {
		filter = bson.D{{Key: "$or", Value: conditions}}
	} else {
		filter = bson.D{{Key: "$and", Value: conditions}}
	}
	if generator.Action == "get" {
		var result *mongo.Cursor
		result, err = mongoClient.Collection(generator.Collection).Find(context.TODO(), filter, opts)
		if err != nil {
			return
		}
		err = result.All(context.TODO(), &data)
	}
	if generator.Action == "count" {
		var count int64
		count, err = mongoClient.Collection(generator.Collection).CountDocuments(context.TODO(), filter)
		if err != nil {
			return
		}
		data = append(data, map[string]interface{}{"count": count})
	}
	if generator.Action == "sum" {
		//fmt.Sprintf("$%s", generator.Field[0])

		pipeline := mongo.Pipeline{
			{{"$match", filter}},
			{{"$group", bson.D{{"_id", nil}, {"total", bson.D{{"$sum", fmt.Sprintf("$%s", generator.Field[0])}}}}}},
		}
		var cur *mongo.Cursor
		cur, err = mongoClient.Collection(generator.Collection).Aggregate(context.Background(), pipeline)
		if err != nil {
			return
		}
		defer cur.Close(context.Background())
		for cur.Next(context.Background()) {
			var result bson.M
			err = cur.Decode(&result)
			if err != nil {
				return
			}
			data = append(data, map[string]interface{}{"total": result["total"]})
		}
	}
	return
}
func getCondition(filter database.GeneratorFilter) bson.D {
	switch filter.Operator {
	case "=":
		return bson.D{{Key: filter.Key, Value: filter.Value}}
	case ">":
		return bson.D{{Key: "$gt", Value: bson.D{{Key: filter.Key, Value: filter.Value}}}}
	case ">=":
		return bson.D{{Key: "$gte", Value: bson.D{{Key: filter.Key, Value: filter.Value}}}}
	case "<":
		return bson.D{{Key: "$lt", Value: bson.D{{Key: filter.Key, Value: filter.Value}}}}
	case "<=":
		return bson.D{{Key: "$lte", Value: bson.D{{Key: filter.Key, Value: filter.Value}}}}
	default:
		return bson.D{{Key: filter.Key, Value: filter.Value}}
	}

}
func createPinsView() {
	views, err := mongoClient.ListCollectionNames(context.Background(), bson.M{"name": PinsView})
	if err != nil {
		return
	}
	if len(views) == 0 {
		mongoClient.CreateView(
			context.Background(),
			PinsView,
			PinsCollection,
			bson.A{
				bson.D{{Key: "$unionWith", Value: MempoolPinsCollection}},
			},
		)
	}
}
