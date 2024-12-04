package metaso

import (
	"context"
	"encoding/json"
	"manindexer/database/mongodb"
	"manindexer/pin"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TweetWithLike struct {
	Tweet
	Like []string `json:"like"`
}

func getNewest(lastId string, size int64, listType string, metaid string, followed string) (listData []*TweetWithLike, total int64, err error) {
	var list []*Tweet
	filter := bson.D{}
	totalFilter := bson.D{}
	if lastId != "" {
		var objectId primitive.ObjectID
		objectId, err = primitive.ObjectIDFromHex(lastId)
		if err != nil {
			return
		}
		filter = append(filter, bson.E{Key: "_id", Value: bson.D{{Key: "$lt", Value: objectId}}})
	}
	if metaid != "" && followed == "1" {
		followList, err1 := getAddressFollowing(metaid)
		if err1 != nil || len(followList) == 0 {
			err = nil
			return
		}
		totalFilter = append(totalFilter, bson.E{Key: "createmetaid", Value: bson.D{{Key: "$in", Value: followList}}})
		filter = append(filter, bson.E{Key: "createmetaid", Value: bson.D{{Key: "$in", Value: followList}}})
	} else if metaid != "" && followed == "" {
		filter = append(filter, bson.E{Key: "createmetaid", Value: metaid})
		totalFilter = append(totalFilter, bson.E{Key: "createmetaid", Value: metaid})
	}
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: listType, Value: -1}})
	findOptions.SetLimit(size)
	result, err := mongoClient.Collection(BuzzView).Find(context.TODO(), filter, findOptions)
	if err != nil {
		return
	}
	err = result.All(context.TODO(), &list)
	if err == mongo.ErrNoDocuments {
		err = nil
	}
	var pinIdList []string
	for _, item := range list {
		item.Content = string(item.ContentBody)
		item.ContentBody = nil
		pinIdList = append(pinIdList, item.Id)
	}

	mempoolList, err := getBuzzMempoolCount(pinIdList)
	if err == nil {
		for _, item := range list {
			for _, data := range mempoolList {
				if item.Id == data.Target && data.Path == "/protocols/paylike" {
					item.LikeCount += 1
				}
				if item.Id == data.Target && data.Path == "/protocols/paycomment" {
					item.CommentCount += 1
				}
			}
		}
	}
	likeMap, err := batchGetPayLike(pinIdList)
	if err == nil {
		for _, item := range list {
			if v, ok := likeMap[item.Id]; ok {
				listData = append(listData, &TweetWithLike{Tweet: *item, Like: v})
			} else {
				listData = append(listData, &TweetWithLike{Tweet: *item, Like: []string{}})
			}
		}
	}
	total, err = mongoClient.Collection(BuzzView).CountDocuments(context.TODO(), totalFilter)

	return
}

func getBuzzMempoolCount(pinIdList []string) (mempoolData []MempoolData, err error) {
	filter := bson.D{{Key: "target", Value: bson.D{{Key: "$in", Value: pinIdList}}}}
	resultMempool, err := mongoClient.Collection(MetaSoMempoolCollection).Find(context.TODO(), filter)
	if err != nil {
		return
	}
	resultMempool.All(context.TODO(), &mempoolData)
	return
}

func getAddressFollowing(metaid string) (list []string, err error) {
	filterA := bson.M{"followmetaid": metaid, "status": true}
	result, err := mongoClient.Collection(mongodb.FollowCollection).Find(context.TODO(), filterA)
	if err != nil {
		return
	}
	var followData []*pin.FollowData //pin.FollowData
	err = result.All(context.TODO(), &followData)
	for _, item := range followData {
		list = append(list, item.MetaId)
	}
	return
}
func batchGetPayLike(pinIdList []string) (list map[string][]string, err error) {
	list = make(map[string][]string)
	filter1 := bson.D{{Key: "liketopinid", Value: bson.D{{Key: "$in", Value: pinIdList}}}}
	result, err := mongoClient.Collection(TweetLikeCollection).Find(context.TODO(), filter1)
	var likeList []*TweetLike
	if err == nil {
		result.All(context.TODO(), &likeList)
	}
	for _, like := range likeList {
		if like.IsLike != "1" {
			continue
		}
		list[like.LikeToPinId] = append(list[like.LikeToPinId], like.CreateMetaid)
	}
	//mempool
	filter2 := bson.D{{Key: "target", Value: bson.D{{Key: "$in", Value: pinIdList}}}, {Key: "path", Value: "/protocols/paylike"}}
	resultMempool, err := mongoClient.Collection(MetaSoMempoolCollection).Find(context.TODO(), filter2)
	if err == nil {
		var mempoolData []MempoolData
		resultMempool.All(context.TODO(), &mempoolData)
		for _, data := range mempoolData {
			if data.IsCancel == 1 {
				if v, ok := list[data.Target]; ok {
					list[data.Target] = deleteSlice(v, data.CreateMetaId)
				}
			} else {
				list[data.Target] = append(list[data.Target], data.CreateMetaId)
			}
		}
	}
	return
}
func deleteSlice(s []string, elem string) []string {
	r := s[:0]
	for _, v := range s {
		if v != elem {
			r = append(r, v)
		}
	}
	return r
}
func getInfo(pinId string) (tweet *Tweet, comments []*TweetComment, like []*TweetLike, err error) {
	filter := bson.D{{Key: "id", Value: pinId}}
	err = mongoClient.Collection(BuzzView).FindOne(context.TODO(), filter, nil).Decode(&tweet)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err = nil
		}
		return
	}
	tweet.Content = string(tweet.ContentBody)
	tweet.ContentBody = nil
	filter2 := bson.D{{Key: "commentpinid", Value: pinId}}

	result, err := mongoClient.Collection(TweetCommentCollection).Find(context.TODO(), filter2)
	if err == nil {
		result.All(context.TODO(), &comments)
	}

	filter3 := bson.D{{Key: "liketopinid", Value: pinId}}
	result2, err := mongoClient.Collection(TweetLikeCollection).Find(context.TODO(), filter3)
	if err == nil {
		result2.All(context.TODO(), &like)
	}
	//mempool
	filter4 := bson.D{{Key: "target", Value: pinId}}
	resultMempool, err := mongoClient.Collection(MetaSoMempoolCollection).Find(context.TODO(), filter4)
	if err != nil {
		return
	}
	var mempoolData []MempoolData
	resultMempool.All(context.TODO(), &mempoolData)
	for _, data := range mempoolData {
		if data.Path == "/protocols/paylike" {
			var likeData TweetLike
			err := json.Unmarshal([]byte(data.Content), &likeData)
			if err == nil {
				like = append(like, &likeData)
				tweet.LikeCount += 1
			}
		} else if data.Path == "/protocols/paycomment" {
			var commentData TweetComment
			err := json.Unmarshal([]byte(data.Content), &commentData)
			if err == nil {
				comments = append(comments, &commentData)
				tweet.CommentCount += 1
			}
		}
	}
	return
}
func getBlockInfo(height int64, host string, cursor int64, size int64, orderby string) (list []*HostData, err error) {
	var filter primitive.D
	if height > 0 {
		filter = bson.D{{Key: "blockHeight", Value: height}}
	} else {
		filter = bson.D{{Key: "host", Value: host}}
	}
	if orderby == "" {
		orderby = "txCount"
	}
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: orderby, Value: -1}})
	findOptions.SetSkip(cursor).SetLimit(size)
	result, err := mongoClient.Collection(HostDataCollection).Find(context.TODO(), filter, findOptions)
	if err != nil {
		return
	}
	err = result.All(context.TODO(), &list)
	if err == mongo.ErrNoDocuments {
		err = nil
	}
	return
}
func getTickByAddress(address string, tickType string) (list []*Mrc20DeployInfo, err error) {
	filter := bson.D{{Key: "address", Value: address}}
	if tickType == "idcoins" {
		filter = append(filter, bson.E{Key: "idcoin", Value: 1})
	}
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "tick", Value: 1}})
	result, err := mongoClient.Collection(MetasoTickCollection).Find(context.TODO(), filter, findOptions)
	if err != nil {
		return
	}
	err = result.All(context.TODO(), &list)
	if err == mongo.ErrNoDocuments {
		err = nil
	}
	return
}
func getMempoolFollow(metaid string) (list []*string, err error) {
	filter := bson.D{{Key: "target", Value: metaid}, {Key: "path", Value: "/follow"}}
	resultMempool, err := mongoClient.Collection(MetaSoMempoolCollection).Find(context.TODO(), filter)
	if err != nil {
		return
	}
	var mempoolData []MempoolData
	resultMempool.All(context.TODO(), &mempoolData)
	for _, data := range mempoolData {
		list = append(list, &data.Content)
	}
	return
}
