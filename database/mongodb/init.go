package mongodb

import (
	"context"
	"fmt"
	"manindexer/common"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func protocolsInit(mongoClient *mongo.Database) {
	protocols := common.Config.Protocols
	if len(protocols) > 0 {
		for name, config := range protocols {
			for _, index := range config.Indexes {
				indexName := fmt.Sprintf("index_%s_%s", name, strings.Join(index.Fields, "_"))
				keys := bson.D{}
				for _, key := range index.Fields {
					keys = append(keys, bson.E{Key: key, Value: 1})
				}
				createIndexIfNotExists(mongoClient, strings.ToLower(name), indexName, keys, index.Unique)
			}
		}
	}
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
