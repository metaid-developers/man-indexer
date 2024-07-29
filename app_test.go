package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"manindexer/adapter/bitcoin"
	"manindexer/common"
	"manindexer/database"
	"manindexer/database/mongodb"
	"manindexer/man"
	"manindexer/pin"
	"strconv"
	"testing"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
)

func TestGetBlock(t *testing.T) {
	common.InitConfig()
	chain := &bitcoin.BitcoinChain{}
	block, err := chain.GetBlock(1)
	fmt.Println(err)
	b := block.(*wire.MsgBlock)
	fmt.Println(b.Header.BlockHash().String())

	txret, err := chain.GetTransaction("798a14129d9697906908046998431ee9e97293bc6c5e8d9d3418f1d944913608")
	fmt.Println(err)
	tx := txret.(*btcutil.Tx)
	fmt.Println("HasWitness", tx.HasWitness())
	for _, out := range tx.MsgTx().TxOut {
		fmt.Println(out.Value)
	}

	indexer := &bitcoin.Indexer{ChainParams: &chaincfg.TestNet3Params}
	pins := indexer.CatchPinsByTx(tx.MsgTx(), 123, 11123232, "", "", 0)
	fmt.Println(len(pins))
	for _, pin := range pins {
		fmt.Println("----------------")
		fmt.Printf("%+v\n", pin)
		//fmt.Println("-----------------\ncontent:", string(inscription.Pin.ContentBody))
		//fmt.Println(hex.EncodeToString(inscription.Pin.ContentBody))
	}
}
func TestGetPin(t *testing.T) {
	txId := "793e32472f85e94cae3ea552c320c362137a84b864d6cda6f342864375f4dbcf"
	chain := &bitcoin.BitcoinChain{}
	txret, err := chain.GetTransaction(txId)
	if err != nil {
		return
	}
	tx := txret.(*btcutil.Tx)
	fmt.Println("HasWitness", tx.HasWitness())
	indexer := &bitcoin.Indexer{ChainParams: &chaincfg.TestNet3Params}
	pins := indexer.CatchPinsByTx(tx.MsgTx(), 0, 0, "", "", 0)
	fmt.Println(pins)
	for _, pin := range pins {
		fmt.Println(string(pin.ContentBody))
	}
}
func TestAddMempoolPin(t *testing.T) {
	dbAdapter := &mongodb.Mongodb{}
	pin, err := dbAdapter.GetPinByNumberOrId("2")
	fmt.Println(err, pin.Address)
	err = dbAdapter.AddMempoolPin(pin)
	fmt.Println(err)
}
func TestDelMempoolPin(t *testing.T) {
	common.InitConfig()
	man.InitAdapter("btc", "mongo", "1", "1")
	man.DeleteMempoolData(2572919, "btc")
}
func TestConfig(t *testing.T) {
	config := common.Config
	fmt.Println(config.Protocols)
	decimals, err := strconv.ParseInt("", 10, 64)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(decimals)
}

func TestGetDbPin(t *testing.T) {
	common.InitConfig()
	man.InitAdapter("btc", "mongo", "1", "1")
	p, err := man.DbAdapter.GetPinByNumberOrId("999")
	fmt.Println(err)
	//fmt.Println(string(p.ContentBody))
	//contentType := common.DetectContentType(&p.ContentBody)
	//fmt.Println(contentType)
	standardEncoded := base64.StdEncoding.EncodeToString(p.ContentBody)
	fmt.Println(standardEncoded)
}
func TestMongoGeneratorFind(t *testing.T) {
	jsonData := `
	{"collection":"pins","action":"sum","filterRelation":"or","field":["number"],
	"filter":[{"operator":"=","key":"number","value":1},{"operator":"=","key":"number","value":2}],
	"cursor":0,"limit":1,"sort":["number","desc"]
	}
	`
	var g database.Generator
	err := json.Unmarshal([]byte(jsonData), &g)
	fmt.Println(err)
	fmt.Println(g.Action)
	mg := mongodb.Mongodb{}
	ret, err := mg.GeneratorFind(g)
	fmt.Println(err, len(ret))
	if err == nil {
		jsonStr, err1 := json.Marshal(ret)
		if err1 != nil {
			fmt.Println("Error marshalling JSON:", err)
		}
		fmt.Println(string(jsonStr))
	}
}
func TestGetSaveData(t *testing.T) {
	common.InitConfig()
	man.InitAdapter("btc", "mongo", "1", "1")
	pinList, _, _, _, _, mrc20List, _, _, err := man.GetSaveData("btc", 2868996)
	fmt.Println(err, len(pinList), len(mrc20List))
	// var testList []*pin.PinInscription
	// for _, mrc20 := range mrc20List {
	// 	if mrc20.GenesisTransaction == "3f7f5a5b31b97df8d8c568b649ce8e8f38f39db714a8f52ac104b6d2dd889d45" {
	// 		testList = append(testList, mrc20)
	// 	}
	// }
	//man.Mrc20Handle(testList)
	man.Mrc20Handle(mrc20List)
}
func TestCatchData(t *testing.T) {
	common.InitConfig()
	man.InitAdapter("btc", "mongo", "1", "1")
	from := 2868128
	to := 2868128
	// for i := from; i <= to; i++ {
	// 	man.DoIndexerRun("btc", int64(i))
	// }
	man.DoIndexerRun("btc", int64(from))
	man.DoIndexerRun("btc", int64(to))

}
func TestHash(t *testing.T) {
	common.InitConfig()
	add := "tb1qtjqupfjej6a9wu94g374fvnlq6ks9v4am7hwtz"
	h := common.GetMetaIdByAddress(add)
	fmt.Println(add)
	fmt.Println(h)
}
func TestGetOwner(t *testing.T) {
	common.InitConfig()
	man.InitAdapter("btc", "mongo", "1", "1")
	//txResult, err := man.ChainAdapter.GetTransaction("d8373e66a6852331c667c94bdccdac94b4908b7ca47b35a00d90a76ae29eb015")
	//fmt.Println(err)
	//tx := txResult.(*btcutil.Tx)
	//inpitId := "8fb1a5154b013f1efaae82a922e03419d6d765006812e6cf32e7b8709971a8c7:0"
	//man.IndexerAdapter.GetOWnerAddress()
	// index := bitcoin.Indexer{
	// 	ChainParams: &chaincfg.TestNet3Params,
	// 	PopCutNum:   common.Config.Btc.PopCutNum,
	// 	DbAdapter:   &man.DbAdapter,
	// }
	// info, err := index.GetOWnerAddress(inpitId, tx.MsgTx())
	// fmt.Println(err)
	// fmt.Printf("%+v", info)
	// list, err := index.TransferCheck(tx.MsgTx())
	// fmt.Println(err)
	// for _, l := range list {
	// 	fmt.Printf("%+v", l)
	// }
	ll, e := man.DbAdapter.GetMempoolTransfer("tb1q3h9twrcz7s5mz7q2eu6pneex446tp3v5yasnp5", "")
	fmt.Println(e, len(ll))
}
func TestRarityScoreBinary(t *testing.T) {
	str := "00000000000000000000000000354712732267161417502043436707557310655121055015573522441662265776662610002362543123510570022146525640016535265733565315137521366643101110550222"
	//fmt.Println(pin.RarityScoreBinary("000001010101"))
	fmt.Println(pin.RarityScoreBinary("btc", str))

}
func TestMrc721(t *testing.T) {
	common.InitConfig()
	man.InitAdapter("btc", "mongo", "2", "1")
	var pinList []*pin.PinInscription
	m721 := &man.Mrc721{}
	collectionStr := `
	{
	"name":"fullname01",
	"desc":"description",
	"website":"https://the-website-of-the-collection",
	"cover":"metafile://your-nft-cover-pinid",
	"metadata":"any data"
}
	`
	pinNode := pin.PinInscription{
		Operation:   "create",
		Id:          "collection01",
		Path:        "/nft/mrc721/test/collection_desc",
		ContentBody: []byte(collectionStr),
	}
	pinList = append(pinList, &pinNode)
	m721.PinHandle(pinList)

	itemStr := `
	{
		"items":
		[	{
			"pinid":"itemId01",
			"name":"itemId01",
			"desc":"the description of the specific NFT",
			"cover":"metafile://your-nft-cover-pinid",
			"metadata":"any arbitrary data you can place here"
			},
			{
			"pinid":"itemId02",
			"name":"itemId02",
			"desc":"the description of the specific NFT",
			"cover":"metafile://your-nft-cover-pinid",
			"metadata":"any arbitrary data you can place here"
			}
		]
		}
	`
	itemStr2 := `
	{
		"items":
		[	{
			"pinid":"itemId01",
			"name":"itemId01",
			"desc":"the description of the specific NFT",
			"cover":"metafile://your-nft-cover-pinid",
			"metadata":"any arbitrary data you can place here"
			},
			{
			"pinid":"itemId04",
			"name":"itemId04",
			"desc":"the description of the specific NFT",
			"cover":"metafile://your-nft-cover-pinid",
			"metadata":"any arbitrary data you can place here"
			}
		]
		}
	`
	itemPin := pin.PinInscription{
		Operation:   "create",
		Id:          "item01",
		Path:        "/nft/mrc721/test/item_desc",
		ContentBody: []byte(itemStr),
	}
	itemPin2 := pin.PinInscription{
		Operation:   "create",
		Id:          "item01",
		Path:        "/nft/mrc721/test/item_desc",
		ContentBody: []byte(itemStr2),
	}
	var itemPinList []*pin.PinInscription
	itemPinList = append(itemPinList, &itemPin)
	itemPinList = append(itemPinList, &itemPin2)
	m721.PinHandle(itemPinList)

}
func TestMrc721Save(t *testing.T) {
	common.InitConfig()
	man.InitAdapter("btc", "mongo", "2", "1")
	man.DoIndexerRun("btc", int64(237))
}
