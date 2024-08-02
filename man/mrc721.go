package man

import (
	"manindexer/mrc721"
	"manindexer/pin"
	"strings"
)

type Mrc721 struct{}

func (m721 *Mrc721) PinHandle(pinList []*pin.PinInscription) {
	validator := Mrc721Validator{}
	var itemList []*mrc721.Mrc721ItemDescPin
	var itemDescList []*mrc721.Mrc721ItemDescPin
	curBlockItemCount := make(map[string]int64)
	collections := make(map[string]mrc721.Mrc721CollectionDescPin)
	var itemPinList []*pin.PinInscription
	var itemDescPinList []*pin.PinInscription
	nameList := make(map[string]struct{})
	for _, pinNode := range pinList {
		pathLow := strings.ToLower(pinNode.Path)
		pathArray := strings.Split(pathLow, "/")
		if len(pathLow) < 4 {
			continue
		}
		if pathArray[1] != "nft" || pathArray[2] != "mrc721" {
			continue
		}
		collectionName := pathArray[3]
		op := ""
		if len(pathArray) > 4 {
			op = pathArray[4]
		}
		switch op {
		case "collection_desc":
			collection, err := m721.collectionHandle(collectionName, pinNode, validator)
			if err == nil {
				DbAdapter.SaveMrc721Collection(collection)
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
		collectionList, _, err := DbAdapter.GetMrc721CollectionList(keys, 0, 100000, false)
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
		DbAdapter.SaveMrc721Item(itemList)
	}
	for _, pinNode := range itemDescPinList {
		list, err := m721.itemDescHandle(pinNode, validator)
		if err == nil && len(list) > 0 {
			itemDescList = append(itemDescList, list...)
		}
	}
	if len(itemDescList) > 0 {
		DbAdapter.UpdateMrc721ItemDesc(itemDescList)
	}
	if len(nameList) > 0 {
		keys := make([]string, 0, len(nameList))
		for k := range nameList {
			keys = append(keys, k)
		}
		DbAdapter.BatchUpdateMrc721CollectionCount(keys)
	}
}

func (m721 *Mrc721) collectionHandle(collectionName string, pinNode *pin.PinInscription, validator Mrc721Validator) (collection *mrc721.Mrc721CollectionDescPin, err error) {
	collection, err = validator.Collection(collectionName, pinNode)
	collection.Address = pinNode.Address
	collection.CollectionName = collectionName
	collection.CreateTime = pinNode.Timestamp
	collection.MetaId = pinNode.MetaId
	collection.PinId = pinNode.Id
	return
}
func (m721 *Mrc721) itemDescHandle(pinNode *pin.PinInscription, validator Mrc721Validator) (itemList []*mrc721.Mrc721ItemDescPin, err error) {
	pathLow := strings.ToLower(pinNode.Path)
	pathArray := strings.Split(pathLow, "/")
	collectionName := pathArray[3]
	itemDesc, _, err := validator.ItemDesc(collectionName, pinNode)
	if err != nil {
		return
	}
	for _, item := range itemDesc.Items {
		var itemPin mrc721.Mrc721ItemDescPin
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
func (m721 *Mrc721) itemHandle(pinNode *pin.PinInscription, validator Mrc721Validator, curBlockItemCount *map[string]int64, collections *map[string]mrc721.Mrc721CollectionDescPin) (itemDesc *mrc721.Mrc721ItemDescPin, err error) {
	itemDesc, _, err = validator.Item(pinNode, curBlockItemCount, collections)
	return
}
