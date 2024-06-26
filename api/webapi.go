package api

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"manindexer/common"
	"manindexer/man"
	"manindexer/pin"
	"net/http"
	"strconv"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func formatRootId(rootId string) string {
	if len(rootId) < 6 {
		return ""
	}
	//return fmt.Sprintf("%s...%s", rootId[0:3], rootId[len(rootId)-3:])
	return rootId[0:6]
}
func popLevelCount(pop string) string {
	lv, _ := man.IndexerAdapter.PopLevelCount(pop)
	if lv == -1 {
		return "--"
	}
	return fmt.Sprintf("Lv%d", lv)
}
func popStrShow(pop string) string {
	_, lastStr := man.IndexerAdapter.PopLevelCount(pop)
	return lastStr[0:8] + "..."
}
func CorsMiddleware() gin.HandlerFunc {
	return func(context *gin.Context) {
		method := context.Request.Method

		context.Header("Access-Control-Allow-Origin", "*")
		context.Header("Access-Control-Allow-Credentials", "true")
		context.Header("Access-Control-Allow-Headers", "*")
		context.Header("Access-Control-Allow-Methods", "GET,HEAD,POST,PUT,DELETE,OPTIONS")
		context.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")

		if method == "OPTIONS" {
			context.AbortWithStatus(http.StatusNoContent)
		}
		context.Next()
	}
}
func Start(f embed.FS) {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard
	r := gin.Default()
	funcMap := template.FuncMap{
		"formatRootId":  formatRootId,
		"popLevelCount": popLevelCount,
		"popStrShow":    popStrShow,
	}
	//use embed.FS
	fp, _ := fs.Sub(f, "web/static")
	r.StaticFS("/assets", http.FS(fp))
	tmpl := template.Must(template.New("").Funcs(funcMap).ParseFS(f, "web/template/**/*"))
	r.SetHTMLTemplate(tmpl)
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	r.Use(cors.New(config))
	//r.LoadHTMLGlob("./web/template/**/*")
	//r.Static("/assets", "./web/static")
	r.GET("/", home)
	r.GET("/pin/list/:page", pinPageList)
	r.GET("/metaid/:page", metaid)
	r.GET("/blocks/:page", blocks)
	r.GET("/mempool/:page", mempool)
	r.GET("/block/:height", block)
	r.GET("/pin/:number", pinshow)
	r.GET("/search/:key", searchshow)
	r.GET("/tx/:txid", tx)
	r.GET("/node/:rootid", node)
	r.GET("/content/:number", content)
	r.GET("/stream/:number", stream)

	//btc json api
	btcJsonApi(r)
	log.Println(common.Config.Web.Port)
	if common.Config.Web.KeyFile != "" && common.Config.Web.PemFile != "" {
		r.RunTLS(common.Config.Web.Port, common.Config.Web.PemFile, common.Config.Web.KeyFile)
	} else {
		r.Run(common.Config.Web.Port)
	}

}

// index page
func home(ctx *gin.Context) {
	list, err := man.DbAdapter.GetPinPageList(1, 100)
	if err != nil {
		ctx.String(200, "fail")
	}
	var msg []*pin.PinMsg
	for _, p := range list {
		pmsg := &pin.PinMsg{Content: p.ContentSummary, Number: p.Number, Operation: p.Operation, Id: p.Id, Type: p.ContentTypeDetect, Path: p.Path, Pop: p.Pop, MetaId: p.MetaId}
		msg = append(msg, pmsg)
	}
	count := man.DbAdapter.Count()
	ctx.HTML(200, "home/index.html", gin.H{"Pins": msg, "Count": &count, "Active": "index", "NextPage": 2, "PrePage": 0})
}
func pinPageList(ctx *gin.Context) {
	page, err := strconv.ParseInt(ctx.Param("page"), 10, 64)
	if err != nil {
		ctx.String(200, "fail")
		return
	}
	list, err := man.DbAdapter.GetPinPageList(page, 100)
	if err != nil {
		ctx.String(200, "fail")
	}
	var msg []*pin.PinMsg
	for _, p := range list {
		pmsg := &pin.PinMsg{Content: p.ContentSummary, Number: p.Number, Operation: p.Operation, Id: p.Id, Type: p.ContentTypeDetect, Path: p.Path, Pop: p.Pop}
		msg = append(msg, pmsg)
	}
	count := man.DbAdapter.Count()
	prePage := page - 1
	nextPage := page + 1
	if len(msg) == 0 {
		nextPage = 0
	}
	if prePage <= 0 {
		prePage = 0
	}
	ctx.HTML(200, "home/index.html", gin.H{"Pins": msg, "Count": &count, "Active": "index", "NextPage": nextPage, "PrePage": prePage})
}

func mempool(ctx *gin.Context) {
	page, err := strconv.ParseInt(ctx.Param("page"), 10, 64)
	if err != nil {
		ctx.String(200, "fail")
		return
	}
	list, err := man.DbAdapter.GetMempoolPinPageList(page, 100)
	if err != nil {
		ctx.String(200, "fail")
		return
	}
	var msg []*pin.PinMsg
	for _, p := range list {
		pmsg := &pin.PinMsg{Content: p.ContentSummary, Number: p.Number, Operation: p.Operation, Id: p.Id, Type: p.ContentTypeDetect, Path: p.Path, MetaId: p.MetaId}
		msg = append(msg, pmsg)
	}
	count := man.DbAdapter.Count()
	prePage := page - 1
	nextPage := page + 1
	if len(msg) == 0 {
		nextPage = 0
	}
	if prePage <= 0 {
		prePage = 0
	}
	ctx.HTML(200, "home/mempool.html", gin.H{"Pins": msg, "Count": &count, "Active": "mempool", "NextPage": nextPage, "PrePage": prePage})
}

// metaid page
func metaid(ctx *gin.Context) {
	page, err := strconv.ParseInt(ctx.Param("page"), 10, 64)
	if err != nil {
		ctx.String(200, "fail")
		return
	}
	list, err := man.DbAdapter.GetMetaIdPageList(page, 100)
	if err != nil {
		ctx.String(200, "fail")
		return
	}
	prePage := page - 1
	nextPage := page + 1
	if len(list) == 0 {
		nextPage = 0
	}
	if prePage <= 0 {
		prePage = 0
	}
	ctx.HTML(200, "home/metaid.html", gin.H{"List": list, "Active": "metaid", "NextPage": nextPage, "PrePage": prePage})
}

// pinshow
func pinshow(ctx *gin.Context) {
	pinMsg, err := man.DbAdapter.GetPinByNumberOrId(ctx.Param("number"))
	if err != nil || pinMsg == nil {
		ctx.String(200, "fail")
		return
	}
	pinMsg.ContentBody = []byte{}
	ctx.HTML(200, "home/pin.html", pinMsg)
}

// searchshow
func searchshow(ctx *gin.Context) {
	pinMsg, err := man.DbAdapter.GetPinByMeatIdOrId(ctx.Param("key"))
	if err != nil || pinMsg == nil {
		ctx.HTML(200, "home/search.html", pinMsg)
		return
	}
	pinMsg.ContentBody = []byte{}
	ctx.HTML(200, "home/search.html", gin.H{"Key": ctx.Param("key"), "Data": pinMsg})
}
func content(ctx *gin.Context) {
	p, err := man.DbAdapter.GetPinByNumberOrId(ctx.Param("number"))
	if err != nil || p == nil {
		ctx.String(200, "fail")
		return
	}
	if p.ContentType == "application/mp4" {
		//ctx.Data(200, "application/octet-stream", p.ContentBody)
		ctx.Header("Content-Type", "text/html; charset=utf-8")
		ctx.String(200, `<video controls autoplay muted src="/stream/`+p.Id+`"></viedo>`)
	} else {
		ctx.String(200, string(p.ContentBody))
	}
}
func stream(ctx *gin.Context) {
	p, err := man.DbAdapter.GetPinByNumberOrId(ctx.Param("number"))
	if err != nil || p == nil {
		ctx.String(200, "fail")
		return
	}
	ctx.Data(200, "application/octet-stream", p.ContentBody)
}
func blocks(ctx *gin.Context) {
	page, err := strconv.ParseInt(ctx.Param("page"), 10, 64)
	if err != nil {
		ctx.String(200, "fail")
		return
	}
	list, err := man.DbAdapter.GetPinPageList(page, 100)
	if err != nil {
		ctx.String(200, "fail")
		return
	}
	msgMap := make(map[int64][]*pin.PinMsg)
	var msgList []int64
	for _, p := range list {
		pmsg := &pin.PinMsg{Content: p.ContentSummary, Number: p.Number, Id: p.Id, Type: p.ContentTypeDetect, Height: p.GenesisHeight}
		if _, ok := msgMap[pmsg.Height]; ok {
			msgMap[pmsg.Height] = append(msgMap[pmsg.Height], pmsg)
		} else {
			msgMap[pmsg.Height] = []*pin.PinMsg{pmsg}
			msgList = append(msgList, pmsg.Height)
		}
	}
	prePage := page - 1
	nextPage := page + 1
	if len(list) == 0 {
		nextPage = 0
	}
	if prePage <= 0 {
		prePage = 0
	}
	ctx.HTML(200, "home/blocks.html", gin.H{"msgMap": msgMap, "msgList": msgList, "Active": "blocks", "NextPage": nextPage, "PrePage": prePage})
}

func block(ctx *gin.Context) {
	height, err := strconv.ParseInt(ctx.Param("height"), 10, 64)
	if err != nil {
		ctx.String(200, "fail")
		return
	}
	list, total, err := man.DbAdapter.GetBlockPin(height, 20)
	if err != nil {
		ctx.String(200, "fail")
		return
	}
	var pins []*pin.PinMsg
	for _, p := range list {
		pmsg := &pin.PinMsg{Content: p.ContentSummary, Number: p.Number, Id: p.Id, Type: p.ContentTypeDetect}
		pins = append(pins, pmsg)
	}
	block := man.ChainAdapter.GetBlockMsg(height)
	msg := gin.H{
		"Pins":   pins,
		"PinNum": total,
		"Height": height,
		"Block":  block,
	}
	ctx.HTML(200, "home/block.html", &msg)
}

type txMsgOutput struct {
	Id      string
	Value   int64
	Script  string
	Address string
}

func tx(ctx *gin.Context) {
	txid := ctx.Param("txid")
	trst, err := man.ChainAdapter.GetTransaction(txid)
	if err != nil {
		ctx.String(200, "fail")
		return
	}
	tx := trst.(*btcutil.Tx)
	var outList []*txMsgOutput
	for i, out := range tx.MsgTx().TxOut {
		id := fmt.Sprintf("%s:%d", tx.Hash().String(), i)
		address := man.IndexerAdapter.GetAddress(out.PkScript)
		outList = append(outList, &txMsgOutput{Id: id, Value: out.Value, Script: string(out.PkScript), Address: address})
	}
	msg := gin.H{
		"TxHash":    tx.Hash().String(),
		"InputNum":  len(tx.MsgTx().TxIn),
		"OutPutNum": len(tx.MsgTx().TxOut),
		"TxIn":      tx.MsgTx().TxIn,
		"TxOut":     outList,
	}
	ctx.HTML(200, "home/tx.html", msg)
}

func node(ctx *gin.Context) {
	rootid := ctx.Param("rootid")
	list, total, err := man.DbAdapter.GetMetaIdPin(rootid, 1, 200)
	if err != nil {
		ctx.String(200, "fail")
		return
	}
	ctx.HTML(200, "home/node.html", &gin.H{"RootId": rootid, "Total": total, "Pins": list})
}
