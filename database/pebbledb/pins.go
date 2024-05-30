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

func (pb *Pebble) GetMaxHeight(chainName string) (height int64, err error) {
	value, closer, err := Pb[NumberLog].Get([]byte("maxBlockHeight"))
	if err != nil {
		if err == pebble.ErrNotFound {
			err = nil
		}
		log.Println(err)
		return
	}
	defer closer.Close()
	height, err = strconv.ParseInt(string(value), 10, 64)
	return
}
func (pb *Pebble) GetMaxNumber() (number int64) {
	value, closer, err := Pb[NumberLog].Get([]byte("maxPinNumer"))
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
func getNumLog(key string) (num int64) {
	value, closer, err := Pb[NumberLog].Get([]byte(key))
	if err != nil && err != pebble.ErrNotFound {
		return
	}
	if err != pebble.ErrNotFound {
		defer closer.Close()
		num, _ = strconv.ParseInt(string(value), 10, 64)
	} else {
		num = 0
	}
	if num < 0 {
		num = 0
	}
	return
}
func addNumLog(key string, num int64) (err error) {
	historyNum := getNumLog(key)
	v := strconv.FormatInt(historyNum+num, 10)
	return Pb[NumberLog].Set([]byte(key), []byte(v), pebble.Sync)
}
func (pb *Pebble) BatchAddPins(pins []interface{}) (err error) {
	batchPins := Pb[PinsCollection].NewBatch()
	batchNumber := Pb[PinsNumber].NewBatch()
	batchAddress := Pb[AddressPins].NewBatch()
	batchPinRoot := Pb[PinRootId].NewBatch()
	batchPath := Pb[PinsPath].NewBatch()
	maxBlockHeight := int64(-1)
	maxPinNumer := int64(-1)
	defer func() {
		batchPins.Close()
		batchNumber.Close()
		batchAddress.Close()
		batchPinRoot.Close()
		batchPath.Close()
	}()
	opts := &pebble.WriteOptions{
		Sync: true,
	}
	blockPins := make(map[string][]string)
	for _, item := range pins {
		pinNode := item.(*pin.PinInscription)
		b, err := json.Marshal(&pinNode)
		if err != nil {
			continue
		}
		var pinId strings.Builder
		pinId.WriteString("@")
		pinId.WriteString(pinNode.Id)
		batchAddress.Merge([]byte(pinNode.Address), []byte(pinId.String()), opts)
		batchPins.Set([]byte(pinNode.Id), b, opts)
		if pinNode.Number > maxPinNumer {
			maxPinNumer = pinNode.Number
		}
		if pinNode.GenesisHeight > maxBlockHeight {
			maxBlockHeight = pinNode.GenesisHeight
		}
		//path data
		if pinNode.Path != "/" && pinNode.Path != "" {
			var pathKey strings.Builder
			pathKey.WriteString(pinNode.Path)
			pathKey.WriteString("_")
			pathKey.WriteString(strconv.FormatInt(pinNode.Number+1000000000, 10))
			batchPath.Set([]byte(pathKey.String()), []byte(pinNode.Id), opts)
			addNumLog(pinNode.Path, 1)
		}
		number := strconv.FormatInt(pinNode.Number, 10)
		batchNumber.Set([]byte(number), []byte(pinNode.Id), opts)
		h := strconv.FormatInt(pinNode.GenesisHeight, 10)
		blockPins[h] = append(blockPins[h], pinNode.Id)
		if pinNode.Operation == "init" {
			batchPinRoot.Set([]byte(pinNode.Address), []byte(pinNode.Id), opts)
		}
	}
	if err = batchPins.Commit(pebble.Sync); err != nil {
		log.Fatalf("Error committing batch: %v", err)
		return
	}
	if err = batchNumber.Commit(pebble.Sync); err != nil {
		log.Fatalf("Error committing batch: %v", err)
		return
	}
	if err = batchAddress.Commit(pebble.Sync); err != nil {
		log.Fatalf("Error committing batch: %v", err)
		return
	}
	if err = batchPinRoot.Commit(pebble.Sync); err != nil {
		log.Fatalf("Error committing batch: %v", err)
		return
	}
	if err = batchPath.Commit(pebble.Sync); err != nil {
		log.Fatalf("Error committing batch: %v", err)
		return
	}

	for k, v := range blockPins {
		s := strings.Join(v, ",")
		Pb[BlockPins].Set([]byte(k), []byte(s), opts)
	}
	if maxBlockHeight >= 0 {
		Pb[NumberLog].Set([]byte("maxBlockHeight"), []byte(strconv.FormatInt(maxBlockHeight, 10)), opts)
	}
	if maxPinNumer >= 0 {
		Pb[NumberLog].Set([]byte("maxPinNumer"), []byte(strconv.FormatInt(maxPinNumer, 10)), opts)
	}
	return
}
func (pb *Pebble) UpdateTransferPin(trasferMap map[string]*pin.PinTransferInfo) (err error) {
	return
}
func (pb *Pebble) BatchUpdatePins(pins []*pin.PinInscription) (err error) {
	return
}
func batchGetPinIdByNumber(numbers [][]byte) (pinIdList [][]byte) {
	getBatch := Pb[PinsNumber].NewIndexedBatch()
	defer getBatch.Close()
	for _, number := range numbers {
		value, closer, err1 := getBatch.Get(number)
		if err1 != nil {
			continue
		}
		pinIdList = append(pinIdList, value)
		closer.Close()
	}
	return
}
func batchGetPinById(ids [][]byte) (pins []*pin.PinInscription) {
	getBatch := Pb[PinsCollection].NewIndexedBatch()
	defer getBatch.Close()
	for _, id := range ids {
		value, closer, err1 := getBatch.Get(id)
		if err1 != nil {
			continue
		}
		var p pin.PinInscription
		err2 := json.Unmarshal(value, &p)
		if err2 != nil {
			continue
		}
		pins = append(pins, &p)
		closer.Close()
	}
	return
}
func (pb *Pebble) GetPinPageList(page int64, size int64) (pins []*pin.PinInscription, err error) {
	last := pb.GetMaxNumber()
	from := (page - 1) * size
	to := page * size
	if from > last {
		return
	}
	if to > last {
		to = last
	}
	var numbers [][]byte
	for i := to; i >= from; i-- {
		numbers = append(numbers, []byte(strconv.FormatInt(i, 10)))
	}
	pinIdList := batchGetPinIdByNumber(numbers)
	pins = batchGetPinById(pinIdList)
	return
}
func (pb *Pebble) GetPinListByIdList(idList []string) (pinList []*pin.PinInscription, err error) {
	var ids [][]byte
	for _, id := range idList {
		ids = append(ids, []byte(id))
	}
	pinList = batchGetPinById(ids)
	return
}
func (pb *Pebble) GetPinListByOutPutList(outputList []string) (pinList []*pin.PinInscription, err error) {
	return
}
func (pb *Pebble) GetPinListByAddress(address string, addressType string, cursor int64, size int64, cnt string, path string) (pins []*pin.PinInscription, total int64, err error) {
	value, close, err := Pb[AddressPins].Get([]byte(address))
	if err != nil {
		return
	}
	defer close.Close()
	var ids [][]byte
	idList := strings.Split(string(value), "@")
	for _, id := range idList {
		if id == "" {
			continue
		}
		ids = append(ids, []byte(id))
	}
	pins = batchGetPinById(ids)
	return
}
func (pb *Pebble) GetPinUtxoCountByAddress(address string) (utxoNum int64, utxoSum int64, err error) {
	return
}

func getPinById(id []byte) (pinNode *pin.PinInscription, err error) {
	value, closer, err := Pb[PinsCollection].Get(id)
	if err != nil {
		if err == pebble.ErrNotFound {
			return getMemPoolPinById(id)
		}
		return
	}
	defer closer.Close()
	var p pin.PinInscription
	err = json.Unmarshal(value, &p)
	pinNode = &p
	return
}
func getMemPoolPinById(id []byte) (pinNode *pin.PinInscription, err error) {
	value, closer, err := Pb[MempoolPinsCollection].Get(id)
	if err != nil {
		return
	}
	defer closer.Close()
	var p pin.PinInscription
	err = json.Unmarshal(value, &p)
	pinNode = &p
	return
}
func getPinIdByNumber(number []byte) (id string, err error) {
	value, closer, err := Pb[PinsNumber].Get(number)
	if err != nil {
		log.Println(err)
		return
	}
	defer closer.Close()
	id = string(value)
	return
}
func (pb *Pebble) GetPinByNumberOrId(numberOrId string) (pinInscription *pin.PinInscription, err error) {
	_, err1 := strconv.ParseInt(numberOrId, 10, 64)
	id := numberOrId
	if err1 == nil {
		id, err = getPinIdByNumber([]byte(numberOrId))
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	return getPinById([]byte(id))
}
func (pb *Pebble) GetPinByOutput(output string) (pinInscription *pin.PinInscription, err error) {
	return
}
func (pb *Pebble) GetPinByMeatIdOrId(key string) (pinInscription *pin.PinInscription, err error) {
	return
}
func (pb *Pebble) GetBlockPin(height int64, size int64) (pins []*pin.PinInscription, total int64, err error) {
	return
}

func (pb *Pebble) GetMetaIdPin(address string, page int64, size int64) (pins []*pin.PinInscription, total int64, err error) {
	return
}
func (pb *Pebble) GetChildNodeById(pinId string) (pins []*pin.PinInscription, err error) {

	return
}
func (pb *Pebble) GetParentNodeById(pinId string) (pinnode *pin.PinInscription, err error) {

	return
}
func keyUpperBound(b []byte) []byte {
	end := make([]byte, len(b))
	copy(end, b)
	for i := len(end) - 1; i >= 0; i-- {
		end[i] = end[i] + 1
		if end[i] != 0 {
			return end[:i+1]
		}
	}
	return nil // no upper-bound
}
func prefixIterOptions(prefix []byte) pebble.IterOptions {
	return pebble.IterOptions{
		LowerBound: prefix,
		UpperBound: keyUpperBound(prefix),
	}
}
func (pb *Pebble) GetAllPinByPath(page, limit int64, path string, metaidList []string) (pins []*pin.PinInscription, total int64, err error) {
	if path == "" {
		return
	}
	pins, memPoolCount, _ := getMemPoolPinsByPath(page, limit, path)
	if memPoolCount >= limit {
		return
	}
	limit = limit - memPoolCount
	p := prefixIterOptions([]byte(path))
	iter, err := Pb[PinsPath].NewIter(&p)
	if err != nil {
		return
	}
	defer iter.Close()
	iter.Last()
	from := (page - 1) * limit
	for i := int64(0); i < from; i++ {
		iter.Prev()
	}
	count := int64(0)
	idList := make([]string, limit)
	for ; iter.Valid() && count < limit; iter.Prev() {
		value := iter.Value()
		idList = append(idList, string(value))
		count++
	}
	idList2 := make([][]byte, limit)
	for _, s := range idList {
		idList2 = append(idList2, []byte(s))
	}
	pins2 := batchGetPinById(idList2)
	pins = append(pins, pins2...)
	total = getNumLog(path) + getNumLog("mem_"+path)
	return
}
func getMemPoolPinsByPath(page, limit int64, path string) (pins []*pin.PinInscription, count int64, err error) {
	if path == "" {
		return
	}
	p := prefixIterOptions([]byte(path))
	iter, err := Pb[MempoolPathPins].NewIter(&p)
	if err != nil {
		return
	}
	defer iter.Close()
	iter.Last()
	from := (page - 1) * limit
	for i := int64(0); i < from; i++ {
		iter.Prev()
	}
	for ; iter.Valid() && count < limit; iter.Prev() {
		value := iter.Value()
		var p pin.PinInscription
		err2 := json.Unmarshal(value, &p)
		if err2 != nil {
			continue
		}
		pins = append(pins, &p)
		count++
	}
	return
}
func (pb *Pebble) BatchAddProtocolData(pins []*pin.PinInscription) (err error) {
	dataMap := make(map[string][]*pin.PinInscription)
	for _, pinItem := range pins {
		keyArr := strings.Split(pinItem.Path, "/")
		key := keyArr[len(keyArr)-1]
		if list, ok := dataMap[key]; ok {
			dataMap[key] = append(list, pinItem)
		} else {
			dataMap[key] = []*pin.PinInscription{pinItem}
		}
	}
	for collectionName, pinList := range dataMap {
		data := getDataByContent(pinList)
		if len(data) > 0 {
			upsertProtocolData(data, collectionName)
		}
	}
	return
}
func upsertProtocolData(data []map[string]interface{}, collectionName string) (err error) {
	batch := PbProtocols[collectionName].NewBatch()
	defer batch.Close()
	opts := &pebble.WriteOptions{
		Sync: true,
	}
	for _, info := range data {
		bytes, _ := json.Marshal(info)
		keyPos := ProtocolsKey[strings.ToLower(collectionName)]
		// if info[keyPos] == nil {
		// 	continue
		// }
		keyData := info[keyPos].(string)
		n := info["pinNumber"].(int64) + 1000000000
		var key strings.Builder
		key.WriteString(keyData)
		key.WriteString("_")
		key.WriteString(strconv.FormatInt(n, 10))
		batch.Set([]byte(key.String()), bytes, opts)
	}
	if err = batch.Commit(pebble.Sync); err != nil {
		log.Printf("Error committing batch: %v", err)
		return
	}
	return
}
func getDataByContent(pinList []*pin.PinInscription) (data []map[string]interface{}) {
	for _, pinItem := range pinList {
		var d map[string]interface{}
		err := json.Unmarshal(pinItem.ContentBody, &d)
		if err == nil {
			d["pinId"] = pinItem.Id
			d["pinNumber"] = pinItem.Number
			d["pinAddress"] = pinItem.Address
			data = append(data, d)
		} else {
			//fmt.Println(err)
		}
	}
	return
}
func (pb *Pebble) GetMemPoolPinByNumberOrId(numberOrId string) (pinInscription *pin.PinInscription, err error) {

	return
}
func getMempoolMaxNumber() (number int64) {
	value, closer, err := Pb[NumberLog].Get([]byte("mempoolMaxNumber"))
	if err != nil {
		if err == pebble.ErrNotFound {
			err = nil
		}
		log.Println(err)
		return
	}
	defer closer.Close()
	number, _ = strconv.ParseInt(string(value), 10, 64)
	number += 1
	return
}
func (pb *Pebble) AddMempoolPin(pin *pin.PinInscription) (err error) {
	opts := &pebble.WriteOptions{
		Sync: true,
	}
	b, err := json.Marshal(pin)
	if err != nil {
		return
	}
	if checkMempoolPinsCollection([]byte(pin.Id)) {
		return
	}
	num := getMempoolMaxNumber()
	err = Pb[MempoolPinsCollection].Set([]byte(pin.Id), b, opts)
	Pb[NumberLog].Set([]byte("mempoolMaxNumber"), []byte(strconv.FormatInt(num, 10)), opts)
	//add MempoolPathPins
	if pin.Path != "/" && pin.Path != "" {
		key := fmt.Sprintf("%s_%s_%d", pin.Path, pin.Id, pin.Timestamp)
		Pb[MempoolPathPins].Set([]byte(key), b, opts)
		addNumLog("mem_"+pin.Path, 1)
	}
	//add MempoolMetaIdInfo
	if pin.OriginalPath == "/info/name" || pin.OriginalPath == "/info/avatar" || pin.Path == "/info/name" || pin.Path == "/info/avatar" {
		key := fmt.Sprintf("%s_%s_%d", pin.Address, pin.Id, pin.Timestamp)
		Pb[MempoolMetaIdInfo].Set([]byte(key), b, opts)

	}
	//add MempoolRootPin
	if pin.Operation == "init" {
		fmt.Println("add root:", pin.Address)
		Pb[MempoolRootPin].Set([]byte(pin.Address), b, opts)
	}
	return
}
func checkMempoolPinsCollection(key []byte) bool {
	_, closer, err := Pb[MempoolPinsCollection].Get(key)
	if err != nil {
		return false
	}
	defer closer.Close()
	return true
}
func (pb *Pebble) GetMempoolPinPageList(page int64, size int64) (pins []*pin.PinInscription, err error) {
	iter, err := Pb[MempoolPinsCollection].NewIter(nil)
	if err != nil {
		return
	}
	defer iter.Close()
	iter.Last()
	from := (page - 1) * size
	for i := int64(0); i < from; i++ {
		iter.Prev()
	}
	count := int64(0)
	for ; iter.Valid() && count < size; iter.Prev() {
		//key := iter.Key()
		value := iter.Value()
		var p pin.PinInscription
		err2 := json.Unmarshal(value, &p)
		if err2 != nil {
			continue
		}
		pins = append(pins, &p)
		count++
	}
	return
}
func getAllMempoolPins(txIds []string) (pins []*pin.PinInscription, err error) {
	getBatch := Pb[MempoolPinsCollection].NewIndexedBatch()
	defer getBatch.Close()
	for _, id := range txIds {
		value, closer, err1 := getBatch.Get([]byte(id))
		if err1 != nil {
			continue
		}
		var p pin.PinInscription
		err2 := json.Unmarshal(value, &p)
		if err2 != nil {
			continue
		}
		pins = append(pins, &p)
		closer.Close()
	}
	return
}
func (pb *Pebble) DeleteMempoolInscription(txIds []string) (err error) {
	batch := Pb[MempoolPinsCollection].NewBatch()
	batchPath := Pb[MempoolPathPins].NewBatch()
	batchMetaId := Pb[MempoolMetaIdInfo].NewBatch()
	defer func() {
		batch.Close()
		batchPath.Close()
		batchMetaId.Close()
	}()
	pins, _ := getAllMempoolPins(txIds)
	for _, pinNode := range pins {
		if pinNode.Path != "/" && pinNode.Path != "" {
			key := fmt.Sprintf("%s_%s_%d", pinNode.Path, pinNode.Id, pinNode.Timestamp)
			batchPath.Delete([]byte(key), pebble.Sync)
			addNumLog("mem_"+pinNode.Path, -1)
		}
		if pinNode.OriginalPath == "/info/name" || pinNode.OriginalPath == "/info/avatar" || pinNode.Path == "/info/name" || pinNode.Path == "/info/avatar" {
			key := fmt.Sprintf("%s_%s_%d", pinNode.Address, pinNode.Id, pinNode.Timestamp)
			batchPath.Delete([]byte(key), pebble.Sync)
		}
		if pinNode.Operation == "init" {
			fmt.Println("del root:", pinNode.Address)
			Pb[MempoolRootPin].Delete([]byte(pinNode.Address), pebble.Sync)
		}
	}
	for _, id := range txIds {
		batch.Delete([]byte(id), pebble.Sync)
	}

	if err = batch.Commit(pebble.Sync); err != nil {
		log.Fatalf("Error committing batch: %v", err)
		//return
	}
	if err = batchPath.Commit(pebble.Sync); err != nil {
		log.Fatalf("Error committing batch: %v", err)
		//return
	}
	return
}
func (pb *Pebble) AddMempoolTransfer(transferData *pin.MemPoolTrasferPin) (err error) {
	return
}
func (pb *Pebble) GetMempoolTransfer(address string, act string) (list []*pin.MemPoolTrasferPin, err error) {
	return
}
func (pb *Pebble) GetMempoolTransferById(pinId string) (result *pin.MemPoolTrasferPin, err error) {
	return
}
