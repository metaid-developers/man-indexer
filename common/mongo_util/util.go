package mongo_util

import (
	"context"
	"fmt"
	"reflect"

	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateIndexIfNotExists(mongoClient *mongo.Database, collectionName, indexName string, keys bson.D, unique bool) error {
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
func CreateIndexWithFilterIfNotExists(mongoClient *mongo.Database, collectionName, indexName string, keys bson.D, unique bool, filter bson.D) error {
	exists, err := checkIndexExists(mongoClient, collectionName, indexName)
	if err != nil {
		return err
	}
	if !exists {
		collection := mongoClient.Collection(collectionName)
		indexModel := mongo.IndexModel{
			Keys:    keys,
			Options: options.Index().SetName(indexName).SetUnique(unique).SetPartialFilterExpression(filter),
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
