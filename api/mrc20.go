package api

import (
	"manindexer/api/respond"
	"manindexer/man"
	"manindexer/mrc20"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func mrc20JsonApi(r *gin.Engine) {
	mrc20Group := r.Group("/api/mrc20")
	mrc20Group.Use(CorsMiddleware())
	mrc20Group.GET("/tick/all", allTick)
	mrc20Group.GET("/tick/info/:id", getTickInfoById)
	mrc20Group.GET("/tick/address", getHistoryByAddress)
	mrc20Group.GET("/tick/history", getHistoryById)
	mrc20Group.GET("/address/balance/:address", getBalanceByAddress)
	mrc20Group.GET("/tx/history", getHistoryByTx)
	mrc20Group.GET("/address/shovel/list", getShovelListByAddress)
	mrc20Group.GET("/shovel/used", getUsedShovelListByTickId)
}

func allTick(ctx *gin.Context) {
	cursor, err := strconv.ParseInt(ctx.Query("cursor"), 10, 64)
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
	completed := ctx.Query("completed")
	orderType := ctx.Query("orderType")
	total, list, err := man.DbAdapter.GetMrc20TickPageList(cursor, size, order, completed, orderType)
	if err != nil || list == nil {
		if err == mongo.ErrNoDocuments || len(list) == 0 {
			ctx.JSON(http.StatusOK, respond.ErrNoDataFound)
		} else {
			ctx.JSON(http.StatusOK, respond.ErrServiceError)
		}
		return
	}
	ctx.JSON(http.StatusOK, respond.ApiSuccess(1, "ok", gin.H{"list": list, "total": total}))
}
func getTickInfoById(ctx *gin.Context) {
	info, err := man.DbAdapter.GetMrc20TickInfo(ctx.Param("id"))
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
func getHistoryByAddress(ctx *gin.Context) {
	cursor, err := strconv.ParseInt(ctx.Query("cursor"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, respond.ErrParameterError)
		return
	}
	size, err := strconv.ParseInt(ctx.Query("size"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, respond.ErrParameterError)
		return
	}
	tickId := ctx.Query("tickId")
	address := ctx.Query("address")
	status := ctx.Query("status")
	verify := ctx.Query("verify")

	list, total, err := man.DbAdapter.GetHistoryByAddress(tickId, address, cursor, size, status, verify)
	if err != nil || list == nil {
		if err == mongo.ErrNoDocuments || len(list) == 0 {
			ctx.JSON(http.StatusOK, respond.ErrNoDataFound)
		} else {
			ctx.JSON(http.StatusOK, respond.ErrServiceError)
		}
		return
	}
	ctx.JSON(http.StatusOK, respond.ApiSuccess(1, "ok", gin.H{"list": list, "total": total}))
}
func getHistoryById(ctx *gin.Context) {
	cursor, err := strconv.ParseInt(ctx.Query("cursor"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, respond.ErrParameterError)
		return
	}
	size, err := strconv.ParseInt(ctx.Query("size"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, respond.ErrParameterError)
		return
	}
	tickId := ctx.Query("tickId")
	list, total, err := man.DbAdapter.GetMrc20HistoryPageList(tickId, false, cursor, size)
	if err != nil || list == nil {
		if err == mongo.ErrNoDocuments || len(list) == 0 {
			ctx.JSON(http.StatusOK, respond.ErrNoDataFound)
		} else {
			ctx.JSON(http.StatusOK, respond.ErrServiceError)
		}
		return
	}
	ctx.JSON(http.StatusOK, respond.ApiSuccess(1, "ok", gin.H{"list": list, "total": total}))
}

func getBalanceByAddress(ctx *gin.Context) {
	address := ctx.Param("address")
	cursor, err := strconv.ParseInt(ctx.Query("cursor"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, respond.ErrParameterError)
		return
	}
	size, err := strconv.ParseInt(ctx.Query("size"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, respond.ErrParameterError)
		return
	}
	list, total, err := man.DbAdapter.GetMrc20BalanceByAddress(address, cursor, size)
	if err != nil || list == nil {
		if err == mongo.ErrNoDocuments || len(list) == 0 {
			ctx.JSON(http.StatusOK, respond.ErrNoDataFound)
		} else {
			ctx.JSON(http.StatusOK, respond.ErrServiceError)
		}
		return
	}
	ctx.JSON(http.StatusOK, respond.ApiSuccess(1, "ok", gin.H{"list": list, "total": total}))
}
func getHistoryByTx(ctx *gin.Context) {
	cursor, err := strconv.ParseInt(ctx.Query("cursor"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, respond.ErrParameterError)
		return
	}
	size, err := strconv.ParseInt(ctx.Query("size"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, respond.ErrParameterError)
		return
	}
	index, err := strconv.ParseInt(ctx.Query("index"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, respond.ErrParameterError)
		return
	}
	txId := ctx.Query("txId")
	list, total, err := man.DbAdapter.GetHistoryByTx(txId, index, cursor, size)
	if err != nil || list == nil {
		if err == mongo.ErrNoDocuments || len(list) == 0 {
			ctx.JSON(http.StatusOK, respond.ErrNoDataFound)
		} else {
			ctx.JSON(http.StatusOK, respond.ErrServiceError)
		}
		return
	}
	ctx.JSON(http.StatusOK, respond.ApiSuccess(1, "ok", gin.H{"list": list, "total": total}))
}
func getShovelListByAddress(ctx *gin.Context) {
	cursor, err := strconv.ParseInt(ctx.Query("cursor"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, respond.ErrParameterError)
		return
	}
	size, err := strconv.ParseInt(ctx.Query("size"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, respond.ErrParameterError)
		return
	}
	tickId := ctx.Query("tickId")
	address := ctx.Query("address")
	//fmt.Println(cursor, size, tickId, address)
	info, err := man.DbAdapter.GetMrc20TickInfo(tickId)
	if err != nil {
		ctx.JSON(http.StatusOK, respond.ErrNoDataFound)
		return
	}
	lv := int(0)
	path := ""
	query := ""
	key := ""
	value := ""
	operator := ""
	if info.Qual.Lv != "" {
		lv, _ = strconv.Atoi(info.Qual.Lv)
	}
	if info.Qual.Path != "" {
		path, query, key, operator, value = mrc20.PathParse(info.Qual.Path)
		if path != "" && query != "" {
			if key == "" && operator == "" && value == "" {
				query = query[2 : len(query)-2]
			}
		} else if path == "" {
			path = info.Qual.Path
		}
	}

	list, total, err := man.DbAdapter.GetShovelListByAddress(address, tickId, info.Qual.Creator, lv, path, query, key, operator, value, cursor, size)
	if err != nil || list == nil {
		if err == mongo.ErrNoDocuments || len(list) == 0 {
			ctx.JSON(http.StatusOK, respond.ErrNoDataFound)
		} else {
			ctx.JSON(http.StatusOK, respond.ErrServiceError)
		}
		return
	}
	ctx.JSON(http.StatusOK, respond.ApiSuccess(1, "ok", gin.H{"list": list, "total": total}))

	//ctx.String(200, "ok")
}
func getUsedShovelListByTickId(ctx *gin.Context) {
	cursor, err := strconv.ParseInt(ctx.Query("cursor"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, respond.ErrParameterError)
		return
	}
	size, err := strconv.ParseInt(ctx.Query("size"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, respond.ErrParameterError)
		return
	}
	address := ctx.Query("address")
	tickId := ctx.Query("tickId")
	list, total, err := man.DbAdapter.GetUsedShovelIdListByAddress(address, tickId, cursor, size)
	if err != nil || list == nil {
		if err == mongo.ErrNoDocuments || len(list) == 0 {
			ctx.JSON(http.StatusOK, respond.ErrNoDataFound)
		} else {
			ctx.JSON(http.StatusOK, respond.ErrServiceError)
		}
		return
	}
	ctx.JSON(http.StatusOK, respond.ApiSuccess(1, "ok", gin.H{"list": list, "total": total}))

}
