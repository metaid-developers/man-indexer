package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo/options"
)

func (mg *Mongodb) BatchAddPinTree(data []interface{}) (err error) {
	ordered := false
	option := options.InsertManyOptions{Ordered: &ordered}
	_, err = mongoClient.Collection(PinTreeCatalogCollection).InsertMany(context.TODO(), data, &option)
	return
}
