package api

import (
	"manindexer/api/respond"
	"manindexer/man"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func mrc721JsonApi(r *gin.Engine) {
	mrc20Group := r.Group("/api/mrc721")
	mrc20Group.Use(CorsMiddleware())
	mrc20Group.GET("/collection/pageList", collectionPageList)
	mrc20Group.GET("/collection/info", collectionInfo)
}

func collectionPageList(ctx *gin.Context) {
	cousor, err := strconv.ParseInt(ctx.Query("cousor"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, respond.ErrParameterError)
		return
	}
	size, err := strconv.ParseInt(ctx.Query("size"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, respond.ErrParameterError)
		return
	}
	list, total, err := man.DbAdapter.GetMrc721CollectionList([]string{}, cousor, size, true)
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
func collectionInfo(ctx *gin.Context) {
	name := ctx.Query("name")
	pinId := ctx.Query("pinId")
	if name == "" && pinId == "" {
		ctx.JSON(http.StatusOK, respond.ErrParameterError)
		return
	}

	data, err := man.DbAdapter.GetMrc721Collection(name, pinId)
	if err != nil || data == nil {
		if err != mongo.ErrNoDocuments {
			ctx.JSON(http.StatusOK, respond.ErrServiceError)
		} else {
			ctx.JSON(http.StatusOK, respond.ErrNoDataFound)
		}
		return
	}
	ctx.JSON(http.StatusOK, respond.ApiSuccess(1, "ok", data))
}
