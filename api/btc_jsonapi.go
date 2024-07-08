package api

import (
	"manindexer/api/respond"
	"manindexer/common"
	"manindexer/database"
	"manindexer/man"
	"manindexer/pin"
	"net/http"
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
	btcGroup.GET("/info/metaid/:metaId", getInfoByMetaId)
	btcGroup.GET("/getAllPinByPath", getAllPinByPath)
	btcGroup.POST("/generalQuery", generalQuery)
	btcGroup.GET("/pin/ByOutput/:output", getPinByOutput)
	btcGroup.GET("/follow/record", getFollowRecord)
	btcGroup.GET("/metaid/followerList/:metaid", getFollowerListByMetaId)
	btcGroup.GET("/metaid/followingList/:metaid", getFollowingListByMetaId)
	btcGroup.POST("/getAllPinByPathAndMetaId", getAllPinByPathAndMetaId)
	btcGroup.POST("/metaid/dataValue", getDataValueByMetaIdList)
}

func metaidList(ctx *gin.Context) {
	page, err := strconv.ParseInt(ctx.Query("page"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, respond.ErrParameterError)
		return
	}
	size, err := strconv.ParseInt(ctx.Query("size"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, respond.ErrParameterError)
		return
	}
	order := ctx.Query("order")
	list, err := man.DbAdapter.GetMetaIdPageList(page, size, order)
	if err != nil || list == nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusOK, respond.ErrNoDataFound)
		} else {
			ctx.JSON(http.StatusOK, respond.ErrServiceError)
		}
		return
	}
	count := man.DbAdapter.Count()
	ctx.JSON(http.StatusOK, respond.ApiSuccess(1, "ok", gin.H{"list": list, "count": &count}))
}
func pinList(ctx *gin.Context) {
	page, err := strconv.ParseInt(ctx.Query("page"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, respond.ErrParameterError)
		return
	}
	size, err := strconv.ParseInt(ctx.Query("size"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, respond.ErrParameterError)
		return
	}
	list, err := man.DbAdapter.GetPinPageList(page, size)
	if err != nil || list == nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusOK, respond.ErrNoDataFound)
		} else {
			ctx.JSON(http.StatusOK, respond.ErrServiceError)
		}
		return
	}
	var msg []*pin.PinMsg
	for _, p := range list {
		pmsg := &pin.PinMsg{
			Content: p.ContentSummary, Number: p.Number, Operation: p.Operation,
			Id: p.Id, Type: p.ContentTypeDetect, Path: p.Path, MetaId: p.MetaId,
			Pop: p.Pop, ChainName: p.ChainName,
			InitialOwner: p.InitialOwner, Address: p.Address, CreateAddress: p.CreateAddress,
		}
		msg = append(msg, pmsg)
	}
	count := man.DbAdapter.Count()
	ctx.JSON(http.StatusOK, respond.ApiSuccess(1, "ok", gin.H{"Pins": msg, "Count": &count, "Active": "index"}))
}
func mempoolList(ctx *gin.Context) {
	page, err := strconv.ParseInt(ctx.Query("page"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, respond.ErrParameterError)
		return
	}
	size, err := strconv.ParseInt(ctx.Query("size"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, respond.ErrParameterError)
		return
	}
	list, err := man.DbAdapter.GetMempoolPinPageList(page, size)
	if err != nil || list == nil {
		if err == mongo.ErrNoDocuments || list == nil {
			ctx.JSON(http.StatusOK, respond.ErrNoDataFound)
		} else {
			ctx.JSON(http.StatusOK, respond.ErrServiceError)
		}
		return
	}
	var msg []*pin.PinMsg
	for _, p := range list {
		pmsg := &pin.PinMsg{Content: p.ContentSummary, Number: p.Number, Operation: p.Operation, Id: p.Id, Type: p.ContentTypeDetect, Path: p.Path, MetaId: p.MetaId}
		msg = append(msg, pmsg)
	}
	count := man.DbAdapter.Count()
	ctx.JSON(http.StatusOK, respond.ApiSuccess(1, "ok", gin.H{"Pins": msg, "Count": &count, "Active": "mempool"}))
}
func nodeList(ctx *gin.Context) {
	page, err := strconv.ParseInt(ctx.Query("page"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, respond.ErrParameterError)
		return
	}
	size, err := strconv.ParseInt(ctx.Query("size"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, respond.ErrParameterError)
		return
	}
	rootid := ctx.Query("rootid")
	list, total, err := man.DbAdapter.GetMetaIdPin(rootid, page, size)
	if err != nil || list == nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusOK, respond.ErrNoDataFound)
		} else {
			ctx.JSON(http.StatusOK, respond.ErrServiceError)
		}
		return
	}
	ctx.JSON(http.StatusOK, respond.ApiSuccess(1, "ok", gin.H{"RootId": rootid, "Total": total, "Pins": list}))
}

// get pin by id
func getPinById(ctx *gin.Context) {
	pinMsg, err := man.DbAdapter.GetPinByNumberOrId(ctx.Param("numberOrId"))
	if err != nil || pinMsg == nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusOK, respond.ErrNoPinFound)
		} else {
			ctx.JSON(http.StatusOK, respond.ErrServiceError)
		}
		return
	}
	//pinMsg.ContentBody = []byte{}
	pinMsg.ContentSummary = string(pinMsg.ContentBody)
	pinMsg.PopLv, _ = pin.PopLevelCount(pinMsg.ChainName, pinMsg.Pop)
	pinMsg.Preview = common.Config.Web.Host + "/pin/" + pinMsg.Id
	pinMsg.Content = common.Config.Web.Host + "/content/" + pinMsg.Id
	check, err := man.DbAdapter.GetMempoolTransferById(pinMsg.Id)
	if err == nil && check != nil {
		pinMsg.Status = -9
	}
	ctx.JSON(http.StatusOK, respond.ApiSuccess(1, "ok", pinMsg))
}
func getPinByOutput(ctx *gin.Context) {
	pinMsg, err := man.DbAdapter.GetPinByOutput(ctx.Param("output"))
	if err != nil || pinMsg == nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusOK, respond.ErrNoPinFound)
		} else {
			ctx.JSON(http.StatusOK, respond.ErrServiceError)
		}
		return
	}
	//pinMsg.ContentBody = []byte{}
	pinMsg.ContentSummary = string(pinMsg.ContentBody)
	pinMsg.Preview = common.Config.Web.Host + "/pin/" + pinMsg.Id
	pinMsg.Content = common.Config.Web.Host + "/content/" + pinMsg.Id
	pinMsg.PopLv, _ = pin.PopLevelCount(pinMsg.ChainName, pinMsg.Pop)
	ctx.JSON(http.StatusOK, respond.ApiSuccess(1, "ok", pinMsg))
}

func blockList(ctx *gin.Context) {
	page, err := strconv.ParseInt(ctx.Query("page"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, respond.ErrParameterError)
		return
	}
	size, err := strconv.ParseInt(ctx.Query("size"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, respond.ErrParameterError)
		return
	}
	list, err := man.DbAdapter.GetPinPageList(page, size)
	if err != nil || list == nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusOK, respond.ErrNoDataFound)
		} else {
			ctx.JSON(http.StatusOK, respond.ErrServiceError)
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
	ctx.JSON(http.StatusOK, respond.ApiSuccess(1, "ok", gin.H{"msgMap": msgMap, "msgList": msgList, "Active": "blocks"}))
}

// get Pin Utxo Count By Address
func getPinUtxoCountByAddress(ctx *gin.Context) {
	if ctx.Param("address") == "" {
		ctx.JSON(http.StatusOK, respond.ErrAddressIsEmpty)
	}
	utxoNum, utxoSum, err := man.DbAdapter.GetPinUtxoCountByAddress(ctx.Param("address"))
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusOK, respond.ErrNoDataFound)
		} else {
			ctx.JSON(http.StatusOK, respond.ErrServiceError)
		}
		return
	}
	ctx.JSON(http.StatusOK, respond.ApiSuccess(1, "ok", gin.H{"utxoNum": utxoNum, "utxoSum": utxoSum}))
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
			ctx.JSON(http.StatusOK, respond.ErrNoPinFound)
		} else {
			ctx.JSON(http.StatusOK, respond.ErrServiceError)
		}
		return
	}
	//get mempool transfer pin
	memTransRecive := make(map[string]struct{})
	memTransSend := make(map[string]struct{})
	mempoolTransferList, err := man.DbAdapter.GetMempoolTransfer(ctx.Param("address"), "")
	if err == nil {
		for _, transfer := range mempoolTransferList {
			if transfer.FromAddress == ctx.Param("address") {
				memTransSend[transfer.PinId] = struct{}{}
			} else if transfer.ToAddress == ctx.Param("address") {
				memTransRecive[transfer.PinId] = struct{}{}
			}
		}
		total -= int64(len(memTransSend))
	}
	var result []*pin.PinInscription
	if cursor == 0 && len(memTransRecive) > 0 {
		var idList []string
		for k := range memTransRecive {
			idList = append(idList, k)
		}
		list, err := man.DbAdapter.GetPinListByIdList(idList)
		if err == nil && len(list) > 0 {
			for _, p := range list {
				p.Status = -9
				p.ContentBody = []byte{}
				p.Preview = common.Config.Web.Host + "/pin/" + p.Id
				p.Content = common.Config.Web.Host + "/content/" + p.Id
				p.PopLv, _ = pin.PopLevelCount(p.ChainName, p.Pop)
				result = append(result, p)
			}
		}
		total += int64(len(list))
	}
	var fixPinList []*pin.PinInscription
	for _, pinNode := range pinList {
		_, ok := memTransSend[pinNode.Id]
		if ok {
			continue
		}
		pinNode.ContentBody = []byte{}
		pinNode.Preview = common.Config.Web.Host + "/pin/" + pinNode.Id
		pinNode.Content = common.Config.Web.Host + "/content/" + pinNode.Id
		pinNode.PopLv, _ = pin.PopLevelCount(pinNode.ChainName, pinNode.Pop)
		fixPinList = append(fixPinList, pinNode)
	}
	result = append(result, fixPinList...)
	if cnt == "true" {
		ctx.JSON(http.StatusOK, respond.ApiSuccess(1, "ok", gin.H{"list": result, "total": total}))
	} else {
		ctx.JSON(http.StatusOK, respond.ApiSuccess(1, "ok", result))
	}

}

// get child node by id
func getChildNodeById(ctx *gin.Context) {
	pinList, err := man.DbAdapter.GetChildNodeById(ctx.Param("pinId"))
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusOK, respond.ErrNoChildFound)
		} else {
			ctx.JSON(http.StatusOK, respond.ErrServiceError)
		}
		return
	}
	for _, pin := range pinList {
		pin.ContentBody = []byte{}
	}
	ctx.JSON(http.StatusOK, respond.ApiSuccess(1, "ok", pinList))
}

// get parent node by id
func getParentNodeById(ctx *gin.Context) {
	pinMsg, err := man.DbAdapter.GetParentNodeById(ctx.Param("pinId"))
	if err != nil || pinMsg == nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusOK, respond.ErrNoNodeFound)
		} else {
			ctx.JSON(http.StatusOK, respond.ErrServiceError)
		}
		return
	}
	pinMsg.ContentBody = []byte{}
	ctx.JSON(http.StatusOK, respond.ApiSuccess(1, "ok", pinMsg))
}

type metaInfo struct {
	*pin.MetaIdInfo
	Unconfirmed string `json:"unconfirmed"`
}

func getInfoByAddress(ctx *gin.Context) {
	metaid, unconfirmed, err := man.DbAdapter.GetMetaIdInfo(ctx.Param("address"), true, "")
	if err != nil {
		ctx.JSON(http.StatusOK, respond.ErrServiceError)
		return
	}
	if metaid == nil {
		metaid = &pin.MetaIdInfo{MetaId: common.GetMetaIdByAddress(ctx.Param("address")), Address: ctx.Param("address")}
		//ctx.JSON(200, apiError(100, "no metaid found."))
		ctx.JSON(http.StatusOK, respond.ApiSuccess(1, "ok", metaInfo{metaid, ""}))
		return
	}
	if metaid.Address == "" {
		metaid.Address = ctx.Param("address")
	}
	if metaid.MetaId == "" {
		metaid.MetaId = common.GetMetaIdByAddress(ctx.Param("address"))
	}
	ctx.JSON(http.StatusOK, respond.ApiSuccess(1, "ok", metaInfo{metaid, unconfirmed}))
}

func getInfoByMetaId(ctx *gin.Context) {
	metaid, unconfirmed, err := man.DbAdapter.GetMetaIdInfo("", true, ctx.Param("metaId"))
	if err != nil {
		ctx.JSON(http.StatusOK, respond.ErrServiceError)
		return
	}
	if metaid == nil {
		metaid = &pin.MetaIdInfo{MetaId: ctx.Param("metaId"), Address: ""}
		//ctx.JSON(200, apiError(100, "no metaid found."))
		ctx.JSON(http.StatusOK, respond.ApiSuccess(1, "ok", metaInfo{metaid, ""}))
		return
	}

	if metaid.MetaId == "" {
		metaid.MetaId = ctx.Param("metaId")
	}
	ctx.JSON(http.StatusOK, respond.ApiSuccess(1, "ok", metaInfo{metaid, unconfirmed}))
}
func generalQuery(ctx *gin.Context) {
	var g database.Generator
	if err := ctx.BindJSON(&g); err != nil {
		ctx.JSON(http.StatusOK, respond.ErrParameterError)
		return
	}
	ret, err := man.DbAdapter.GeneratorFind(g)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusOK, respond.ErrNoResultFound)
		} else {
			ctx.JSON(http.StatusOK, respond.ErrServiceError)
		}
		return
	}
	ctx.JSON(http.StatusOK, respond.ApiSuccess(1, "ok", ret))
}
func getAllPinByPath(ctx *gin.Context) {
	page, err := strconv.ParseInt(ctx.Query("page"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, respond.ApiError(101, "page parameter error"))
		return
	}
	limit, err := strconv.ParseInt(ctx.Query("limit"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, respond.ApiError(101, "limit parameter error"))
		return
	}
	if ctx.Query("path") == "" {
		ctx.JSON(http.StatusOK, respond.ApiError(101, "parentPath parameter error"))
		return
	}
	pinList1, total, err := man.DbAdapter.GetAllPinByPath(page, limit, ctx.Query("path"), []string{})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusOK, respond.ErrNoPinFound)
		} else {
			ctx.JSON(http.StatusOK, respond.ErrServiceError)
		}
		return
	}
	var pinList []*pin.PinInscription
	for _, pinNode := range pinList1 {
		pinNode.ContentSummary = string(pinNode.ContentBody)
		pinList = append(pinList, pinNode)
	}
	ctx.JSON(http.StatusOK, respond.ApiSuccess(1, "ok", gin.H{"list": pinList, "total": total}))
}

// getAllPinByPathAndMetaId
type pinQuery struct {
	Page       int64    `json:"page"`
	Size       int64    `json:"size"`
	Path       string   `json:"path"`
	MetaIdList []string `json:"metaIdList"`
}

func getAllPinByPathAndMetaId(ctx *gin.Context) {
	var q pinQuery
	if err := ctx.BindJSON(&q); err != nil {
		ctx.JSON(http.StatusOK, respond.ErrParameterError)
		return
	}
	pinList1, total, err := man.DbAdapter.GetAllPinByPath(q.Page, q.Size, q.Path, q.MetaIdList)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusOK, respond.ErrNoPinFound)
		} else {
			ctx.JSON(http.StatusOK, respond.ErrServiceError)
		}
		return
	}
	var pinList []*pin.PinInscription
	for _, pinNode := range pinList1 {
		pinNode.ContentSummary = string(pinNode.ContentBody)
		pinList = append(pinList, pinNode)
	}
	ctx.JSON(http.StatusOK, respond.ApiSuccess(1, "ok", gin.H{"list": pinList, "total": total}))
}

// getDataValueByMetaIdList
type stringListQuery struct {
	List []string `json:"list"`
}

func getDataValueByMetaIdList(ctx *gin.Context) {
	var q stringListQuery
	if err := ctx.BindJSON(&q); err != nil {
		ctx.JSON(http.StatusOK, respond.ErrParameterError)
		return
	}
	result, err := man.DbAdapter.GetDataValueByMetaIdList(q.List)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusOK, respond.ErrNoResultFound)
		} else {
			ctx.JSON(http.StatusOK, respond.ErrServiceError)
		}
		return
	}
	ctx.JSON(http.StatusOK, respond.ApiSuccess(1, "ok", result))

}

// getFollowListByMetaId
func getFollowerListByMetaId(ctx *gin.Context) {
	cursorStr := ctx.Query("cursor")
	sizeStr := ctx.Query("size")
	cursor := int64(0)
	size := int64(100)
	myFollow := false
	followDetail := false
	if ctx.Query("followDetail") == "true" {
		followDetail = true
	}
	if cursorStr != "" && sizeStr != "" {
		cursor, _ = strconv.ParseInt(cursorStr, 10, 64)
		size, _ = strconv.ParseInt(sizeStr, 10, 64)
	}
	list, total, err := man.DbAdapter.GetFollowDataByMetaId(ctx.Param("metaid"), myFollow, followDetail, cursor, size)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusOK, respond.ErrNoResultFound)
		} else {
			ctx.JSON(http.StatusOK, respond.ErrServiceError)
		}
		return
	}
	ctx.JSON(http.StatusOK, respond.ApiSuccess(1, "ok", gin.H{"list": list, "total": total}))

}

// getFollowListByMetaId
func getFollowingListByMetaId(ctx *gin.Context) {
	cursorStr := ctx.Query("cursor")
	sizeStr := ctx.Query("size")
	cursor := int64(0)
	size := int64(100)
	myFollow := true
	followDetail := false
	if ctx.Query("followDetail") == "true" {
		followDetail = true
	}
	if cursorStr != "" && sizeStr != "" {
		cursor, _ = strconv.ParseInt(cursorStr, 10, 64)
		size, _ = strconv.ParseInt(sizeStr, 10, 64)
	}
	list, total, err := man.DbAdapter.GetFollowDataByMetaId(ctx.Param("metaid"), myFollow, followDetail, cursor, size)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusOK, respond.ErrNoResultFound)
		} else {
			ctx.JSON(http.StatusOK, respond.ErrServiceError)
		}
		return
	}
	ctx.JSON(http.StatusOK, respond.ApiSuccess(1, "ok", gin.H{"list": list, "total": total}))

}

// getFollowRecord
func getFollowRecord(ctx *gin.Context) {
	metaId := ctx.Query("metaId")
	followMetaId := ctx.Query("followerMetaId")
	if metaId == "" || followMetaId == "" {
		ctx.JSON(http.StatusOK, respond.ErrParameterError)
		return
	}
	info, err := man.DbAdapter.GetFollowRecord(metaId, followMetaId)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusOK, respond.ErrNoResultFound)
		} else {
			ctx.JSON(http.StatusOK, respond.ErrServiceError)
		}
		return
	}
	ctx.JSON(http.StatusOK, respond.ApiSuccess(1, "ok", info))

}
