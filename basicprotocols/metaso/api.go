package metaso

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func Api(r *gin.Engine) {
	accessGroup := r.Group("/social/buzz")
	accessGroup.Use(CorsMiddleware())
	accessGroup.GET("/newest", newest)
	accessGroup.GET("/hot", hot)
	accessGroup.GET("/info", info)
	hostGroup := r.Group("/host")
	hostGroup.Use(CorsMiddleware())
	hostGroup.GET("/block/sync-newest", syncNewest)
	hostGroup.GET("/block/info", blockInfo)
	hostGroup.GET("/info", hostInfo)
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

type ApiResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"message"`
	Data interface{} `json:"data"`
}

func ApiError(code int, msg string) (res *ApiResponse) {
	return &ApiResponse{Code: code, Msg: msg}
}
func ApiNullData(code int, msg string) (res *ApiResponse) {
	return &ApiResponse{Code: code, Msg: msg, Data: []string{}}
}
func ApiSuccess(code int, msg string, data interface{}) (res *ApiResponse) {
	return &ApiResponse{Code: code, Msg: msg, Data: data}
}

func newest(ctx *gin.Context) {
	size, err := strconv.ParseInt(ctx.Query("size"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, ApiError(-1, "size error"))
		return
	}
	if size == 0 {
		size = 10
	}
	list, total, err := getNewest(ctx.Query("lastId"), size, "_id", ctx.Query("metaid"), ctx.Query("followed"))
	lastId := ""
	if len(list) > 0 {
		lastId = list[len(list)-1].MogoID.Hex()
	}
	if err != nil {
		fmt.Println(err)
		ctx.JSON(http.StatusOK, ApiError(-1, "service exception."))
		return
	}
	ctx.JSON(http.StatusOK, ApiSuccess(1, "ok", gin.H{"list": list, "total": total, "lastId": lastId}))
}
func hot(ctx *gin.Context) {
	size, err := strconv.ParseInt(ctx.Query("size"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, ApiError(-1, "size error"))
		return
	}
	if size == 0 {
		size = 10
	}
	list, total, err := getNewest(ctx.Query("lastId"), size, "hot", "", "")
	lastId := ""
	if len(list) > 0 {
		lastId = list[len(list)-1].MogoID.Hex()
	}
	if err != nil {
		ctx.JSON(http.StatusOK, ApiError(-1, "service exception."))
		return
	}
	ctx.JSON(http.StatusOK, ApiSuccess(1, "ok", gin.H{"list": list, "total": total, "lastId": lastId}))
}
func info(ctx *gin.Context) {
	tweet, comments, like, err := getInfo(ctx.Query("pinId"))
	if err != nil {
		ctx.JSON(http.StatusOK, ApiError(-1, "service exception"))
		return
	}
	ctx.JSON(http.StatusOK, ApiSuccess(1, "ok", gin.H{"tweet": tweet, "comments": comments, "like": like}))
}
func syncNewest(ctx *gin.Context) {
	_, height := getSyncHeight()
	ctx.JSON(http.StatusOK, ApiSuccess(1, "ok", height))
}
func blockInfo(ctx *gin.Context) {
	height, err := strconv.ParseInt(ctx.Query("height"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, ApiError(-1, "query height error"))
		return
	}
	cursor, err := strconv.ParseInt(ctx.Query("cursor"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, ApiError(-1, "query cursor error"))
		return
	}
	size, err := strconv.ParseInt(ctx.Query("size"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, ApiError(-1, "query size error"))
		return
	}

	list, err := getBlockInfo(height, "", cursor, size, ctx.Query("orderby"))
	if err != nil {
		ctx.JSON(http.StatusOK, ApiError(-1, "service exception"))
		return
	}
	ctx.JSON(http.StatusOK, ApiSuccess(1, "ok", list))
}
func hostInfo(ctx *gin.Context) {
	cursor, err := strconv.ParseInt(ctx.Query("cursor"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, ApiError(-1, "query cursor error"))
		return
	}
	size, err := strconv.ParseInt(ctx.Query("size"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, ApiError(-1, "query size error"))
		return
	}
	list, err := getBlockInfo(0, ctx.Query("host"), cursor, size, ctx.Query("orderby"))
	if err != nil {
		ctx.JSON(http.StatusOK, ApiError(-1, "service exception"))
		return
	}
	ctx.JSON(http.StatusOK, ApiSuccess(1, "ok", list))
}
