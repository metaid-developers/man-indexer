package mrc721

import (
	"strings"
)

func (m721 *Mrc721) PinHandle(pinList []*Mrc721Pin) {
	validator := Mrc721Validator{}
	var itemList []*Mrc721ItemDescPin
	var itemDescList []*Mrc721ItemDescPin
	curBlockItemCount := make(map[string]int64)
	collections := make(map[string]Mrc721CollectionDescPin)
	var itemPinList []*Mrc721Pin
	var itemDescPinList []*Mrc721Pin
	nameList := make(map[string]struct{})
	for _, pinNode := range pinList {
		pathLow := strings.ToLower(pinNode.Path)
		pathArray := strings.Split(pathLow, "/")
		if len(pathLow) < 11 {
			continue
		}
		if pathLow[0:11] != "/nft/mrc721" || len(pathArray) < 4 {
			continue
		}
		//collectionName := url.PathEscape(pathArray[3])
		collectionName := pathArray[3]
		op := ""
		if len(pathArray) > 4 {
			op = pathArray[4]
		}
		switch op {
		case "collection_desc":
			collection, err := m721.collectionHandle(collectionName, pinNode, validator)
			if err == nil && pinNode.Number != -1 {
				SaveMrc721Collection(collection)
			}
		case "item_desc":
			itemDescPinList = append(itemDescPinList, pinNode)
		default:
			nameList[collectionName] = struct{}{}
			itemPinList = append(itemPinList, pinNode)
		}
	}
	if len(nameList) > 0 {
		keys := make([]string, 0, len(nameList))
		for k := range nameList {
			keys = append(keys, k)
		}
		collectionList, _, err := GetMrc721CollectionList(keys, 0, 100000, false)
		if err == nil && len(collectionList) > 0 {
			for _, cocollection := range collectionList {
				collections[cocollection.CollectionName] = *cocollection
			}
		}
	}

	for _, pinNode := range itemPinList {
		item, err := m721.itemHandle(pinNode, validator, &curBlockItemCount, &collections)
		if err == nil {
			itemList = append(itemList, item)
		}
	}
	if len(itemList) > 0 {
		SaveMrc721Item(itemList)
	}
	for _, pinNode := range itemDescPinList {
		list, err := m721.itemDescHandle(pinNode, validator)
		if err == nil && len(list) > 0 {
			itemDescList = append(itemDescList, list...)
		}
	}
	if len(itemDescList) > 0 {
		UpdateMrc721ItemDesc(itemDescList)
	}
	if len(nameList) > 0 {
		keys := make([]string, 0, len(nameList))
		for k := range nameList {
			keys = append(keys, k)
		}
		BatchUpdateMrc721CollectionCount(keys)
	}
}

func (m721 *Mrc721) collectionHandle(collectionName string, pinNode *Mrc721Pin, validator Mrc721Validator) (collection *Mrc721CollectionDescPin, err error) {
	collection, err = validator.Collection(collectionName, pinNode)
	collection.Address = pinNode.Address
	collection.CollectionName = collectionName
	collection.CreateTime = pinNode.Timestamp
	collection.MetaId = pinNode.MetaId
	collection.PinId = pinNode.Id
	return
}
func (m721 *Mrc721) itemDescHandle(pinNode *Mrc721Pin, validator Mrc721Validator) (itemList []*Mrc721ItemDescPin, err error) {
	pathLow := strings.ToLower(pinNode.Path)
	pathArray := strings.Split(pathLow, "/")
	collectionName := pathArray[3]
	itemDesc, _, err := validator.ItemDesc(collectionName, pinNode)
	if err != nil {
		return
	}
	for _, item := range itemDesc.Items {
		var itemPin Mrc721ItemDescPin
		itemPin.DescPinId = pinNode.Id
		itemPin.ItemPinId = item.PinId
		itemPin.Name = item.Name
		itemPin.Desc = item.Desc
		itemPin.Cover = item.Cover
		itemPin.Metadata = item.Metadata
		itemList = append(itemList, &itemPin)
	}
	return
}
func (m721 *Mrc721) itemHandle(pinNode *Mrc721Pin, validator Mrc721Validator, curBlockItemCount *map[string]int64, collections *map[string]Mrc721CollectionDescPin) (itemDesc *Mrc721ItemDescPin, err error) {
	itemDesc, _, err = validator.Item(pinNode, curBlockItemCount, collections)
	return
}
