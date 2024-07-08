package pebbledb

import (
	"encoding/json"
	"fmt"
	"log"
	"manindexer/pin"
	"strconv"
	"strings"

	"github.com/cockroachdb/pebble"
)

func (pb *Pebble) GetMaxMetaIdNumber() (number int64) {
	value, closer, err := Pb[NumberLog].Get([]byte("maxMetaIdNumer"))
	if err != nil {
		if err == pebble.ErrNotFound {
			err = nil
		}
		log.Println(err)
		return
	}
	defer closer.Close()
	number, _ = strconv.ParseInt(string(value), 10, 64)
	number = number + 1
	return
}

func (pb *Pebble) GetMetaIdInfo(address string, mempool bool, metaid string) (info *pin.MetaIdInfo, unconfirmed string, err error) {

	mempoolInfo, _ := findMetaIdInfoInMempool(address)
	var unconfirmedList []string
	value, closer, err := Pb[MetaIdInfoCollection].Get([]byte(address))
	if err != nil && err != pebble.ErrNotFound {
		return
	}
	if err != pebble.ErrNotFound {
		defer closer.Close()
	}
	if err == pebble.ErrNotFound {
		err = nil
		if mempoolInfo.Number == -1 {
			unconfirmedList = append(unconfirmedList, "number")
			info = &mempoolInfo
		}
	} else {
		var m pin.MetaIdInfo
		err = json.Unmarshal(value, &m)
		info = &m
		if mempoolInfo.Avatar != "" {
			info.Avatar = mempoolInfo.Avatar
			unconfirmedList = append(unconfirmedList, "avatar")
		}
		if mempoolInfo.Name != "" {
			info.Name = mempoolInfo.Name
			unconfirmedList = append(unconfirmedList, "name")
		}
		if mempoolInfo.Bio != "" {
			info.Bio = mempoolInfo.Bio
			unconfirmedList = append(unconfirmedList, "bio")
		}
	}
	if len(unconfirmedList) > 0 {
		unconfirmed = strings.Join(unconfirmedList, ",")
	}
	return
}
func findMetaIdInfoInMempool(address string) (info pin.MetaIdInfo, err error) {
	p := prefixIterOptions([]byte(address))
	iter, err := Pb[MempoolMetaIdInfo].NewIter(&p)
	if err != nil {
		return
	}
	defer iter.Close()
	for iter.First(); iter.Valid(); iter.Next() {
		var pinNode pin.PinInscription
		err := json.Unmarshal(iter.Value(), &pinNode)
		if err != nil {
			continue
		}
		if pinNode.OriginalPath == "/info/name" {
			info.Name = string(pinNode.ContentBody)
		} else if pinNode.OriginalPath == "/info/avatar" {
			info.Avatar = fmt.Sprintf("/content/%s", pinNode.Id)
		} else if pinNode.OriginalPath == "/info/bid" {
			info.Bio = string(pinNode.ContentBody)
		}
	}

	return
}
func (pb *Pebble) BatchUpsertMetaIdInfo(infoList map[string]*pin.MetaIdInfo) (err error) {
	maxMetaIdNumer := int64(-1)
	//batchMetaId := Pb[MetaIdInfoCollection].NewBatch()
	batchNumber := Pb[MetaIdNumber].NewBatch()
	opts := &pebble.WriteOptions{
		Sync: true,
	}
	defer func() {
		//batchMetaId.Close()
		batchNumber.Close()
	}()
	for _, info := range infoList {
		metaId, _, _ := pb.GetMetaIdInfo(info.Address, false, "")
		if metaId == nil {
			metaId = info
		}
		//metaId := info
		if info.Number > 0 {
			metaId.Number = info.Number
			number := strconv.FormatInt(metaId.Number, 10)
			batchNumber.Set([]byte(number), []byte(info.Address), opts)
		}
		if metaId.Number > maxMetaIdNumer {
			maxMetaIdNumer = metaId.Number
		}
		if info.MetaId != "" {
			metaId.MetaId = info.MetaId
		}
		// if info.RootTxId != "" {
		// 	updateInfo = append(updateInfo, bson.E{Key: "roottxid", Value: info.RootTxId})
		// }
		if info.Name != "" {
			metaId.Name = info.Name
		}
		if info.NameId != "" {
			metaId.NameId = info.NameId
		}
		if info.Address != "" {
			metaId.Address = info.Address
		}
		if len(info.Avatar) > 0 {
			metaId.Avatar = info.Avatar
		}
		if len(info.AvatarId) > 0 {
			metaId.AvatarId = info.AvatarId
		}
		if len(info.Bio) > 0 {
			metaId.Bio = info.Bio
		}
		if len(info.BioId) > 0 {
			metaId.BioId = info.BioId
		}
		if len(info.SoulbondToken) > 0 {
			metaId.SoulbondToken = info.SoulbondToken
		}
		b, err := json.Marshal(metaId)
		if err != nil {
			continue
		}
		Pb[MetaIdInfoCollection].Set([]byte(info.Address), b, opts)
	}
	// if err = batchMetaId.Commit(pebble.Sync); err != nil {
	// 	log.Printf("Error committing batch: %v", err)
	// 	return
	// }
	if err = batchNumber.Commit(pebble.Sync); err != nil {
		log.Printf("Error committing batch: %v", err)
		return
	}
	if maxMetaIdNumer >= 0 {
		Pb[NumberLog].Set([]byte("maxMetaIdNumer"), []byte(strconv.FormatInt(maxMetaIdNumer, 10)), opts)
	}
	return
}

func (pb *Pebble) GetMetaIdPageList(page int64, size int64, order string) (pins []*pin.MetaIdInfo, err error) {
	iter, err := Pb[MetaIdInfoCollection].NewIter(nil)
	if err != nil {
		return
	}
	defer iter.Close()
	iter.Last()
	count := 0
	for ; iter.Valid() && count < 10; iter.Prev() {
		//key := iter.Key()
		value := iter.Value()
		var p pin.MetaIdInfo
		err1 := json.Unmarshal(value, &p)
		if err1 != nil {
			continue
		}
		pins = append(pins, &p)
		count++
	}
	return
}

func findRootTxIdInMempool(address string) (rootTxId string, err error) {

	return
}
func (pb *Pebble) BatchUpsertMetaIdInfoAddition(infoList []*pin.MetaIdInfoAdditional) (err error) {
	return
}
func (pb *Pebble) GetDataValueByMetaIdList(list []string) (result []*pin.MetaIdDataValue, err error) {
	return
}
