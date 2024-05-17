package mongodb

import (
	"context"
	"fmt"
	"log"
	"manindexer/common"
	"manindexer/database"
	"manindexer/pin"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	PinsCollection           string = "pins"
	MempoolPinsCollection    string = "mempoolpins"
	MetaIdInfoCollection     string = "metaid"
	PinTreeCatalogCollection string = "pintree"
	// Mrc20PinCollection       string = "mrc20pins"
	// Mrc20TickCollection      string = "mrc20ticks"
	// Mrc20MintShovel          string = "mrc20shovel"
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
	createIndexIfNotExists(mongoClient, PinsCollection, "id_1", bson.D{{Key: "id", Value: 1}}, true)
	createIndexIfNotExists(mongoClient, PinsCollection, "number_1", bson.D{{Key: "number", Value: 1}}, true)
	createIndexIfNotExists(mongoClient, PinsCollection, "address_status_1", bson.D{{Key: "address", Value: 1}, {Key: "status", Value: 1}}, false)

	createIndexIfNotExists(mongoClient, MempoolPinsCollection, "id_1", bson.D{{Key: "id", Value: 1}}, true)

	createIndexIfNotExists(mongoClient, MetaIdInfoCollection, "address_1", bson.D{{Key: "address", Value: 1}}, true)
	createIndexIfNotExists(mongoClient, MetaIdInfoCollection, "roottxid_1", bson.D{{Key: "roottxid", Value: 1}}, false)

	createIndexIfNotExists(mongoClient, PinTreeCatalogCollection, "treepath_1", bson.D{{Key: "treepath", Value: 1}}, true)
	createIndexIfNotExists(mongoClient, PinTreeCatalogCollection, "roottxid_1", bson.D{{Key: "roottxid", Value: 1}}, false)

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
		return bson.D{{Key: "$lt", Value: bson.D{{Key: filter.Key, Value: filter.Value}}}}
	case "<":
		return bson.D{{Key: "$gte", Value: bson.D{{Key: filter.Key, Value: filter.Value}}}}
	case "<=":
		return bson.D{{Key: "$lte", Value: bson.D{{Key: filter.Key, Value: filter.Value}}}}
	default:
		return bson.D{{Key: filter.Key, Value: filter.Value}}
	}

}
