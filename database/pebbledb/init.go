package pebbledb

import (
	"encoding/json"
	"fmt"
	"log"
	"manindexer/common"
	"manindexer/database"
	"manindexer/pin"
	"strings"

	"github.com/cockroachdb/pebble"
)

const (
	NumberLog                string = "numberLog"
	PinsCollection           string = "pins"
	PinsNumber               string = "pinsnumber"
	PinsPath                 string = "pinsPath"
	BlockPins                string = "blockPins"
	AddressPins              string = "addressPins"
	PinRootId                string = "pinRootId"
	MetaIdInfoCollection     string = "metaid"
	MetaIdNumber             string = "metaidNumber"
	MempoolPinsCollection    string = "mempoolpins"
	MempoolPathPins          string = "mempoolPathPins"
	MempoolMetaIdInfo        string = "mempoolMetaIdInfo"
	MempoolRootPin           string = "mempoolRootPin"
	PinTreeCatalogCollection string = "pintree"
)

type Pebble struct{}
type Logger struct{}

func (ml Logger) Infof(format string, args ...interface{}) {}
func (ml Logger) Fatalf(format string, args ...interface{}) {
	log.Println(format, args)
}
func (ml Logger) Errorf(format string, args ...interface{}) {
	log.Println(format, args)
}

var Pb map[string]*pebble.DB
var PbProtocols map[string]*pebble.DB
var ProtocolsKey map[string]string

func (pb *Pebble) InitDatabase() {
	Pb = make(map[string]*pebble.DB, 14)
	open(NumberLog)
	open(PinsCollection)
	open(PinsNumber)
	open(PinsPath)
	open(BlockPins)
	open(AddressPins)
	open(PinRootId)
	open(MetaIdInfoCollection)
	open(MetaIdNumber)
	open(PinTreeCatalogCollection)
	open(MempoolPinsCollection)
	open(MempoolPathPins)
	open(MempoolMetaIdInfo)
	open(MempoolRootPin)
	protocolsInit()
}
func protocolsInit() {
	protocols := common.Config.Protocols
	PbProtocols = make(map[string]*pebble.DB, len(protocols))
	ProtocolsKey = make(map[string]string, len(protocols))
	if len(protocols) > 0 {
		for name, p := range protocols {
			openProtocols(strings.ToLower(name))
			ProtocolsKey[strings.ToLower(name)] = p.Key
		}
	}
}
func open(dbName string) (err error) {
	lg := Logger{}
	dbPath := common.Config.Pebble.Dir
	var db *pebble.DB
	db, err = pebble.Open(dbPath+"/"+dbName, &pebble.Options{Logger: lg})
	if err != nil {
		log.Printf("Pebble %s init error\n", dbName)
	} else {
		Pb[dbName] = db
	}
	return
}
func openProtocols(dbName string) (err error) {
	lg := Logger{}
	dbPath := common.Config.Pebble.Dir
	var db *pebble.DB
	db, err = pebble.Open(dbPath+"/protocols/"+dbName, &pebble.Options{Logger: lg})
	if err != nil {
		log.Printf("Pebble %s init error\n", dbName)
	} else {
		PbProtocols[dbName] = db
	}
	return
}
func (pb *Pebble) Count() (count pin.PinCount) {
	return
}

func (pb *Pebble) GeneratorFind(generator database.Generator) (data []map[string]interface{}, err error) {
	dbName := strings.ToLower(generator.Collection)
	key := strings.ToLower(generator.Filters[0].Value.(string))
	page := generator.Cursor + 1
	limit := generator.Limit
	if dbName == "" || key == "" {
		return
	}
	if PbProtocols[dbName] == nil {
		fmt.Printf("%s is not exit", dbName)
		return
	}
	p := prefixIterOptions([]byte(key))
	iter, err := PbProtocols[dbName].NewIter(&p)
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
	for ; iter.Valid() && count < limit; iter.Prev() {
		value := iter.Value()
		d := make(map[string]interface{})
		err := json.Unmarshal(value, &d)
		if err != nil {
			continue
		}
		data = append(data, d)
		count++
	}
	return
}
