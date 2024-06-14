package api

import (
	"manindexer/api/respond"
	"manindexer/man"
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
}

func allTick(ctx *gin.Context) {
	page, err := strconv.ParseInt(ctx.Query("page"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, respond.ErrParameterError)
	}
	size, err := strconv.ParseInt(ctx.Query("size"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, respond.ErrParameterError)
	}
	order := ctx.Query("order")
	total, list, err := man.DbAdapter.GetMrc20TickPageList(page, size, order)
	if err != nil || list == nil {
		if err == mongo.ErrNoDocuments {
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
