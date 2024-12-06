package mrc721

import (
	"encoding/json"
	"errors"
	"strings"
)

type Mrc721Validator struct {
}

func (validator *Mrc721Validator) Collection(collectionName string, pinNode *Mrc721Pin) (collection *Mrc721CollectionDescPin, err error) {
	//Check op
	if pinNode.Operation != "create" {
		err = errors.New(ErrOperation)
		return
	}
	//Check JSON content.
	var json1 map[string]interface{}
	content := strings.ToLower(string(pinNode.ContentBody))
	err = json.Unmarshal([]byte(content), &json1)
	if err != nil {
		err = errors.New(ErrPinContent)
		return
	}
	err = json.Unmarshal([]byte(content), &collection)
	if err != nil {
		err = errors.New(ErrPinContent)
		return
	}
	if json1["totalsupply"] == nil {
		collection.TotalSupply = -1
	}
	if json1["royaltyrate"] == nil {
		collection.RoyaltyRate = 5
	}
	//Check totalSupply
	if collection.TotalSupply < -1 || collection.TotalSupply > 1e12 {
		err = errors.New(ErrTotalSupply)
		return
	}
	//Check royaltyRate
	if collection.RoyaltyRate < 0 || collection.RoyaltyRate > 20 {
		err = errors.New(ErrRoyaltyRate)
		return
	}
	//Check for unique collection_name.
	find, err := GetMrc721Collection(collectionName, "")
	if err == nil && find != nil {
		err = errors.New(ErrCollectionExist)
		return
	}
	return
}
func (validator *Mrc721Validator) ItemDesc(collectionName string, pinNode *Mrc721Pin) (itemDesc *Mrc721ItemDescList, collectionPinId string, err error) {
	//Check op
	if pinNode.Operation != "create" {
		err = errors.New(ErrOperation)
		return
	}
	//Check JSON content.
	content := strings.ToLower(string(pinNode.ContentBody))
	err = json.Unmarshal([]byte(content), &itemDesc)
	if err != nil {
		err = errors.New(ErrPinContent)
		return
	}
	find, err := GetMrc721Collection(collectionName, "")
	if err != nil || find == nil {
		err = errors.New(ErrCollectionNotExist)
		return
	}
	collectionPinId = find.PinId
	return
}
func (validator *Mrc721Validator) Item(pinNode *Mrc721Pin, curBlockItemCount *map[string]int64, collections *map[string]Mrc721CollectionDescPin) (item *Mrc721ItemDescPin, collectionPinId string, err error) {
	//Check op
	if pinNode.Operation != "create" {
		err = errors.New(ErrOperation)
		return
	}
	pathLow := strings.ToLower(pinNode.Path)
	pathArray := strings.Split(pathLow, "/")
	collectionName := pathArray[3]
	//Check collection
	var collection Mrc721CollectionDescPin
	var ok bool
	if collection, ok = (*collections)[collectionName]; !ok {
		err = errors.New(ErrCollectionNotExist)
		return
	}
	//Check totalSupply
	(*curBlockItemCount)[collectionName] += 1
	v := (*curBlockItemCount)[collectionName]
	t := collection.TotalNum
	if collection.TotalSupply != -1 && collection.TotalSupply < v+t {
		err = errors.New(ErrTotalSupplyEexceeds)
		return
	}
	var mrc721Item Mrc721ItemDescPin
	mrc721Item.Address = pinNode.Address
	mrc721Item.CollectionName = collectionName
	mrc721Item.CollectionPinId = collection.PinId
	mrc721Item.Content = pinNode.ContentBody
	mrc721Item.CreateTime = pinNode.Timestamp
	mrc721Item.ItemPinId = pinNode.Id
	mrc721Item.ItemPinNumber = pinNode.Number
	mrc721Item.MetaId = pinNode.MetaId
	mrc721Item.ContentType = pinNode.ContentType
	mrc721Item.ContentTypeDetect = pinNode.ContentTypeDetect
	mrc721Item.ContentString = pinNode.ContentSummary
	item = &mrc721Item
	return
}
