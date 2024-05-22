package api

import (
	"manindexer/common"
	"manindexer/database"
	"manindexer/man"
	"manindexer/pin"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type ApiResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"message"`
	Data interface{} `json:"data"`
}

func btcJsonApi(r *gin.Engine) {
	btcGroup := r.Group("/api")
	btcGroup.Use(CorsMiddleware())
	btcGroup.GET("/metaid/list", metaidList)
	btcGroup.GET("/pin/list", pinList)
	btcGroup.GET("/block/list", blockList)
	btcGroup.GET("/mempool/list", mempoolList)
	btcGroup.GET("/node/list", nodeList)

	btcGroup.GET("/pin/:numberOrId", getPinById)
	btcGroup.GET("/address/pin/utxo/count/:address", getPinUtxoCountByAddress)
	btcGroup.GET("/address/pin/list/:addressType/:address", getPinListByAddress)
	btcGroup.GET("/node/child/:pinId", getChildNodeById)
	btcGroup.GET("/node/parent/:pinId", getParentNodeById)
	btcGroup.GET("/info/address/:address", getInfoByAddress)
	btcGroup.GET("/getAllPinByPath", getAllPinByPath)
	btcGroup.POST("/generalQuery", generalQuery)
	btcGroup.GET("/pin/ByOutput/:output", getPinByOutput)
}
func apiError(code int, msg string) (res *ApiResponse) {
	return &ApiResponse{Code: code, Msg: msg}
}
func apiSuccess(code int, msg string, data interface{}) (res *ApiResponse) {
	return &ApiResponse{Code: code, Msg: msg, Data: data}
}
func metaidList(ctx *gin.Context) {
	page, err := strconv.ParseInt(ctx.Query("page"), 10, 64)
	if err != nil {
		ctx.JSON(200, apiError(404, "Parameter error."))
	}
	size, err := strconv.ParseInt(ctx.Query("size"), 10, 64)
	if err != nil {
		ctx.JSON(200, apiError(404, "Parameter error."))
	}
	list, err := man.DbAdapter.GetMetaIdPageList(page, size)
	if err != nil || list == nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(200, apiError(100, "no data found."))
		} else {
			ctx.JSON(200, apiError(404, "service exception."))
		}
		return
	}
	ctx.JSON(200, apiSuccess(1, "ok", list))
}
func pinList(ctx *gin.Context) {
	page, err := strconv.ParseInt(ctx.Query("page"), 10, 64)
	if err != nil {
		ctx.JSON(200, apiError(404, "Parameter error."))
	}
	size, err := strconv.ParseInt(ctx.Query("size"), 10, 64)
	if err != nil {
		ctx.JSON(200, apiError(404, "Parameter error."))
	}
	list, err := man.DbAdapter.GetPinPageList(page, size)
	if err != nil || list == nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(200, apiError(100, "no data found."))
		} else {
			ctx.JSON(200, apiError(404, "service exception."))
		}
		return
	}
	var msg []*pin.PinMsg
	for _, p := range list {
		pmsg := &pin.PinMsg{Content: p.ContentSummary, Number: p.Number, Operation: p.Operation, Id: p.Id, Type: p.ContentTypeDetect, Path: p.Path, MetaId: p.MetaId, Pop: p.Pop}
		msg = append(msg, pmsg)
	}
	count := man.DbAdapter.Count()
	ctx.JSON(200, apiSuccess(1, "ok", gin.H{"Pins": msg, "Count": &count, "Active": "index"}))
}
func mempoolList(ctx *gin.Context) {
	page, err := strconv.ParseInt(ctx.Query("page"), 10, 64)
	if err != nil {
		ctx.JSON(200, apiError(404, "Parameter error."))
	}
	size, err := strconv.ParseInt(ctx.Query("size"), 10, 64)
	if err != nil {
		ctx.JSON(200, apiError(404, "Parameter error."))
	}
	list, err := man.DbAdapter.GetMempoolPinPageList(page, size)
	if err != nil || list == nil {
		if err == mongo.ErrNoDocuments || list == nil {
			ctx.JSON(200, apiError(100, "no data found."))
		} else {
			ctx.JSON(200, apiError(404, "service exception."))
		}
		return
	}
	var msg []*pin.PinMsg
	for _, p := range list {
		pmsg := &pin.PinMsg{Content: p.ContentSummary, Number: p.Number, Operation: p.Operation, Id: p.Id, Type: p.ContentTypeDetect, Path: p.Path, MetaId: p.MetaId}
		msg = append(msg, pmsg)
	}
	count := man.DbAdapter.Count()
	ctx.JSON(200, apiSuccess(1, "ok", gin.H{"Pins": msg, "Count": &count, "Active": "mempool"}))
}
func nodeList(ctx *gin.Context) {
	page, err := strconv.ParseInt(ctx.Query("page"), 10, 64)
	if err != nil {
		ctx.JSON(200, apiError(404, "Parameter error."))
	}
	size, err := strconv.ParseInt(ctx.Query("size"), 10, 64)
	if err != nil {
		ctx.JSON(200, apiError(404, "Parameter error."))
	}
	rootid := ctx.Query("rootid")
	list, total, err := man.DbAdapter.GetMetaIdPin(rootid, page, size)
	if err != nil || list == nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(200, apiError(100, "no data found."))
		} else {
			ctx.JSON(200, apiError(404, "service exception."))
		}
		return
	}
	ctx.JSON(200, apiSuccess(1, "ok", gin.H{"RootId": rootid, "Total": total, "Pins": list}))
}

// get pin by id
func getPinById(ctx *gin.Context) {
	pinMsg, err := man.DbAdapter.GetPinByNumberOrId(ctx.Param("numberOrId"))
	if err != nil || pinMsg == nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(200, apiError(100, "no pin found."))
		} else {
			ctx.JSON(200, apiError(404, "service exception."))
		}
		return
	}
	//pinMsg.ContentBody = []byte{}
	pinMsg.ContentSummary = string(pinMsg.ContentBody)
	pinMsg.Preview = common.Config.Web.Host + "/pin/" + pinMsg.Id
	pinMsg.Content = common.Config.Web.Host + "/content/" + pinMsg.Id
	ctx.JSON(200, apiSuccess(1, "ok", pinMsg))
}
func getPinByOutput(ctx *gin.Context) {
	pinMsg, err := man.DbAdapter.GetPinByOutput(ctx.Param("output"))
	if err != nil || pinMsg == nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(200, apiError(100, "no pin found."))
		} else {
			ctx.JSON(200, apiError(404, "service exception."))
		}
		return
	}
	//pinMsg.ContentBody = []byte{}
	pinMsg.ContentSummary = string(pinMsg.ContentBody)
	pinMsg.Preview = common.Config.Web.Host + "/pin/" + pinMsg.Id
	pinMsg.Content = common.Config.Web.Host + "/content/" + pinMsg.Id
	pinMsg.PopLv, _ = man.IndexerAdapter.PopLevelCount(pinMsg.Pop)
	ctx.JSON(200, apiSuccess(1, "ok", pinMsg))
}

func blockList(ctx *gin.Context) {
	page, err := strconv.ParseInt(ctx.Query("page"), 10, 64)
	if err != nil {
		ctx.JSON(200, apiError(404, "Parameter error."))
	}
	size, err := strconv.ParseInt(ctx.Query("size"), 10, 64)
	if err != nil {
		ctx.JSON(200, apiError(404, "Parameter error."))
	}
	list, err := man.DbAdapter.GetPinPageList(page, size)
	if err != nil || list == nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(200, apiError(100, "no data found."))
		} else {
			ctx.JSON(200, apiError(404, "service exception."))
		}
		return
	}
	msgMap := make(map[int64][]*pin.PinMsg)
	var msgList []int64
	for _, p := range list {
		pmsg := &pin.PinMsg{Operation: p.Operation, Path: p.Path, Content: p.ContentSummary, Number: p.Number, Id: p.Id, Type: p.ContentTypeDetect, MetaId: p.MetaId, Height: p.GenesisHeight, Pop: p.Pop}
		if _, ok := msgMap[pmsg.Height]; ok {
			msgMap[pmsg.Height] = append(msgMap[pmsg.Height], pmsg)
		} else {
			msgMap[pmsg.Height] = []*pin.PinMsg{pmsg}
			msgList = append(msgList, pmsg.Height)
		}
	}
	ctx.JSON(200, apiSuccess(1, "ok", gin.H{"msgMap": msgMap, "msgList": msgList, "Active": "blocks"}))
}

// get Pin Utxo Count By Address
func getPinUtxoCountByAddress(ctx *gin.Context) {
	if ctx.Param("address") == "" {
		ctx.JSON(200, apiError(100, "address is null"))
	}
	utxoNum, utxoSum, err := man.DbAdapter.GetPinUtxoCountByAddress(ctx.Param("address"))
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(200, apiError(100, "no  data found."))
		} else {
			ctx.JSON(200, apiError(404, "service exception."))
		}
		return
	}
	ctx.JSON(200, apiSuccess(1, "ok", gin.H{"utxoNum": utxoNum, "utxoSum": utxoSum}))
}

// get pin list by address
func getPinListByAddress(ctx *gin.Context) {
	cursorStr := ctx.Query("cursor")
	sizeStr := ctx.Query("size")
	cnt := ctx.Query("cnt")
	path := ctx.Query("path")
	cursor := int64(0)
	size := int64(10000)
	if cursorStr != "" && sizeStr != "" {
		cursor, _ = strconv.ParseInt(cursorStr, 10, 64)
		size, _ = strconv.ParseInt(sizeStr, 10, 64)
	}
	pinList, total, err := man.DbAdapter.GetPinListByAddress(ctx.Param("address"), ctx.Param("addressType"), cursor, size, cnt, path)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(200, apiError(100, "no  pin found."))
		} else {
			ctx.JSON(200, apiError(404, "service exception."))
		}
		return
	}
	for _, pin := range pinList {
		pin.ContentBody = []byte{}
		pin.Preview = common.Config.Web.Host + "/pin/" + pin.Id
		pin.Content = common.Config.Web.Host + "/content/" + pin.Id
	}
	if cnt == "true" {
		ctx.JSON(200, apiSuccess(1, "ok", gin.H{"list": pinList, "total": total}))
	} else {
		ctx.JSON(200, apiSuccess(1, "ok", pinList))
	}

}

// get child node by id
func getChildNodeById(ctx *gin.Context) {
	pinList, err := man.DbAdapter.GetChildNodeById(ctx.Param("pinId"))
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(200, apiError(100, "no child found."))
		} else {
			ctx.JSON(200, apiError(404, "service exception."))
		}
		return
	}
	for _, pin := range pinList {
		pin.ContentBody = []byte{}
	}
	ctx.JSON(200, apiSuccess(1, "ok", pinList))
}

// get parent node by id
func getParentNodeById(ctx *gin.Context) {
	pinMsg, err := man.DbAdapter.GetParentNodeById(ctx.Param("pinId"))
	if err != nil || pinMsg == nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(200, apiError(100, "no node found."))
		} else {
			ctx.JSON(200, apiError(404, "service exception."))
		}
		return
	}
	pinMsg.ContentBody = []byte{}
	ctx.JSON(200, apiSuccess(1, "ok", pinMsg))
}

type metaInfo struct {
	*pin.MetaIdInfo
	Unconfirmed string `json:"unconfirmed"`
}

func getInfoByAddress(ctx *gin.Context) {
	metaid, unconfirmed, err := man.DbAdapter.GetMetaIdInfo(ctx.Param("address"), true)
	if err != nil {
		ctx.JSON(200, apiError(404, "service exception."))
		return
	}
	if metaid == nil {
		metaid = &pin.MetaIdInfo{MetaId: common.GetMetaIdByAddress(ctx.Param("address")), Address: ctx.Param("address")}
		//ctx.JSON(200, apiError(100, "no metaid found."))
		ctx.JSON(200, apiSuccess(1, "ok", metaInfo{metaid, ""}))
		return
	}
	ctx.JSON(200, apiSuccess(1, "ok", metaInfo{metaid, unconfirmed}))
}

func generalQuery(ctx *gin.Context) {
	var g database.Generator
	if err := ctx.BindJSON(&g); err != nil {
		ctx.JSON(200, apiError(404, "request parameter error."))
		return
	}
	ret, err := man.DbAdapter.GeneratorFind(g)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(200, apiError(100, "no result found."))
		} else {
			ctx.JSON(200, apiError(404, "service exception."))
		}
		return
	}
	ctx.JSON(200, apiSuccess(1, "ok", ret))
}
func getAllPinByPath(ctx *gin.Context) {
	page, err := strconv.ParseInt(ctx.Query("page"), 10, 64)
	if err != nil {
		ctx.JSON(200, apiError(101, "page parameter error"))
		return
	}
	limit, err := strconv.ParseInt(ctx.Query("limit"), 10, 64)
	if err != nil {
		ctx.JSON(200, apiError(101, "limit parameter error"))
		return
	}
	if ctx.Query("path") == "" {
		ctx.JSON(200, apiError(101, "parentPath parameter error"))
		return
	}
	pinList1, total, err := man.DbAdapter.GetAllPinByPath(page, limit, ctx.Query("path"))
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(200, apiError(100, "no  pin found."))
		} else {
			ctx.JSON(200, apiError(404, "service exception."))
		}
		return
	}
	var pinList []*pin.PinInscription
	for _, pinNode := range pinList1 {
		pinNode.ContentSummary = string(pinNode.ContentBody)
		pinList = append(pinList, pinNode)
	}
	ctx.JSON(200, apiSuccess(1, "ok", gin.H{"list": pinList, "total": total}))
}
