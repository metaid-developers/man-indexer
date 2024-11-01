package metaname

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func Api(r *gin.Engine) {
	accessGroup := r.Group("/api/metaname")
	accessGroup.Use(CorsMiddleware())
	accessGroup.GET("/list", list)
	accessGroup.GET("/info", info)
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

func list(ctx *gin.Context) {
	size, err := strconv.ParseInt(ctx.Query("size"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, ApiError(-1, "size error"))
		return
	}
	if size == 0 {
		size = 10
	}
	list, total, err := getNewest(ctx.Query("lastId"), size, "_id")
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
	info, history, err := getInfo(ctx.Query("name"))
	if err != nil {
		ctx.JSON(http.StatusOK, ApiError(-1, "service exception"))
		return
	}
	ctx.JSON(http.StatusOK, ApiSuccess(1, "ok", gin.H{"info": info, "history": history}))
}
