package metaso

import (
	"context"
	"fmt"
	"log"
	"manindexer/common"
	"reflect"
	"time"

	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

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
	}
	mongoClient = client.Database(mg.DbName)
	createIndex(mongoClient)
}
func createIndex(mongoClient *mongo.Database) {
	//Tweet
	createIndexIfNotExists(mongoClient, TweetCollection, "pinid_1", bson.D{{Key: "id", Value: 1}}, true)
	createIndexIfNotExists(mongoClient, TweetCollection, "output_1", bson.D{{Key: "output", Value: 1}}, false)
	createIndexIfNotExists(mongoClient, TweetCollection, "path_1", bson.D{{Key: "path", Value: 1}}, false)
	createIndexIfNotExists(mongoClient, TweetCollection, "chainname_1", bson.D{{Key: "chainname", Value: 1}}, false)
	createIndexIfNotExists(mongoClient, TweetCollection, "timestamp_1", bson.D{{Key: "timestamp", Value: 1}}, false)
	createIndexIfNotExists(mongoClient, TweetCollection, "metaid_1", bson.D{{Key: "metaid", Value: 1}}, false)
	createIndexIfNotExists(mongoClient, TweetCollection, "creatormetaid_1", bson.D{{Key: "creatormetaid", Value: 1}}, false)
	createIndexIfNotExists(mongoClient, TweetCollection, "number_1", bson.D{{Key: "number", Value: 1}}, false)
	createIndexIfNotExists(mongoClient, TweetCollection, "operation_1", bson.D{{Key: "operation", Value: 1}}, false)

	createIndexIfNotExists(mongoClient, TweetLikeCollection, "pinid_1", bson.D{{Key: "pinid", Value: 1}}, true)
	createIndexIfNotExists(mongoClient, TweetLikeCollection, "liketopinid_1", bson.D{{Key: "liketopinid", Value: 1}}, false)
	createIndexIfNotExists(mongoClient, TweetLikeCollection, "createaddress_1", bson.D{{Key: "createaddress", Value: 1}}, false)
	createIndexIfNotExists(mongoClient, TweetLikeCollection, "createmetaid_1", bson.D{{Key: "createmetaid", Value: 1}}, false)
	createIndexIfNotExists(mongoClient, TweetLikeCollection, "islike_1", bson.D{{Key: "islike", Value: 1}}, false)
	createIndexIfNotExists(mongoClient, TweetLikeCollection, "timestamp_1", bson.D{{Key: "timestamp", Value: 1}}, false)

	createIndexIfNotExists(mongoClient, TweetCommentCollection, "pinid_1", bson.D{{Key: "pinid", Value: 1}}, true)
	createIndexIfNotExists(mongoClient, TweetCommentCollection, "commentpinid_1", bson.D{{Key: "commentpinid", Value: 1}}, false)
	createIndexIfNotExists(mongoClient, TweetCommentCollection, "createaddress_1", bson.D{{Key: "createaddress", Value: 1}}, false)
	createIndexIfNotExists(mongoClient, TweetCommentCollection, "createmetaid_1", bson.D{{Key: "createmetaid", Value: 1}}, false)
	createIndexIfNotExists(mongoClient, TweetCommentCollection, "islike_1", bson.D{{Key: "islike", Value: 1}}, false)
	createIndexIfNotExists(mongoClient, TweetCommentCollection, "timestamp_1", bson.D{{Key: "timestamp", Value: 1}}, false)

}
func createIndexIfNotExists(mongoClient *mongo.Database, collectionName, indexName string, keys bson.D, unique bool) error {
	exists, err := checkIndexExists(mongoClient, collectionName, indexName)
	if err != nil {
		return err
	}
	if !exists {
		collection := mongoClient.Collection(collectionName)
		indexModel := mongo.IndexModel{
			Keys:    keys,
			Options: options.Index().SetName(indexName).SetUnique(unique),
		}
		_, err := collection.Indexes().CreateOne(context.Background(), indexModel)
		if err != nil {
			return err
		}
		//fmt.Printf("Index %s created successfully\n", indexName)
	}
	return nil
}
func checkIndexExists(mongoClient *mongo.Database, collectionName, indexName string) (bool, error) {
	collection := mongoClient.Collection(collectionName)
	indexView := collection.Indexes()
	cursor, err := indexView.List(context.Background())
	if err != nil {
		return false, err
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var indexKey bson.M
		if err := cursor.Decode(&indexKey); err != nil {
			return false, err
		}
		if indexKey["name"] == indexName {
			return true, nil
		}
	}
	return false, nil
}

type Decimal decimal.Decimal

func (d Decimal) DecodeValue(dc bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	decimalType := reflect.TypeOf(decimal.Decimal{})
	if !val.IsValid() || !val.CanSet() || val.Type() != decimalType {
		return bsoncodec.ValueDecoderError{
			Name:     "decimalDecodeValue",
			Types:    []reflect.Type{decimalType},
			Received: val,
		}
	}

	var value decimal.Decimal
	switch vr.Type() {
	case bsontype.Decimal128:
		dec, err := vr.ReadDecimal128()
		if err != nil {
			return err
		}
		value, err = decimal.NewFromString(dec.String())
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("received invalid BSON type to decode into decimal.Decimal: %s", vr.Type())
	}

	val.Set(reflect.ValueOf(value))
	return nil
}

func (d Decimal) EncodeValue(ec bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
	decimalType := reflect.TypeOf(decimal.Decimal{})
	if !val.IsValid() || val.Type() != decimalType {
		return bsoncodec.ValueEncoderError{
			Name:     "decimalEncodeValue",
			Types:    []reflect.Type{decimalType},
			Received: val,
		}
	}

	dec := val.Interface().(decimal.Decimal)
	dec128, err := primitive.ParseDecimal128(dec.String())
	if err != nil {
		return err
	}

	return vw.WriteDecimal128(dec128)
}
