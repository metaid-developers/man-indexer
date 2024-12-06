package api

import (
	"manindexer/api/respond"
	"net/http"
	"strconv"

	"manindexer/basicprotocols/mrc721"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func mrc721JsonApi(r *gin.Engine) {
	mrc20Group := r.Group("/api/mrc721")
	mrc20Group.Use(CorsMiddleware())
	mrc20Group.GET("/collection/pageList", collectionPageList)
	mrc20Group.GET("/collection/info", collectionInfo)
	mrc20Group.GET("/collection/items/pageList", itemPageList)
	mrc20Group.GET("/address/collection", addressCollection)
	mrc20Group.GET("/address/item", addressItem)
	mrc20Group.GET("/item/info", itemInfo)
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
	list, total, err := mrc721.GetMrc721CollectionList([]string{}, cousor, size, true)
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

	data, err := mrc721.GetMrc721Collection(name, pinId)
	if err == nil && data == nil {
		ctx.JSON(http.StatusOK, respond.ErrNoDataFound)
		return
	} else if err == mongo.ErrNoDocuments {
		ctx.JSON(http.StatusOK, respond.ErrNoDataFound)
		return
	} else if err != nil {
		ctx.JSON(http.StatusOK, respond.ErrServiceError)
		return
	}
	ctx.JSON(http.StatusOK, respond.ApiSuccess(1, "ok", data))
}
func itemPageList(ctx *gin.Context) {
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
	list, total, err := mrc721.GetMrc721ItemList("", ctx.Query("pinid"), []string{}, cousor, size, true)
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
func addressCollection(ctx *gin.Context) {
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

	list, total, err := mrc721.GetMrc721CollectionByAddress(ctx.Query("address"), cousor, size, true)
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

func addressItem(ctx *gin.Context) {
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

	list, total, err := mrc721.GetMrc721ItemByAddress(ctx.Query("address"), ctx.Query("pinId"), cousor, size, true)
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
func itemInfo(ctx *gin.Context) {
	pinId := ctx.Query("pinId")
	if pinId == "" {
		ctx.JSON(http.StatusOK, respond.ErrParameterError)
		return
	}
	data, err := mrc721.GetMrc721Item(pinId)
	if err == nil && data == nil {
		ctx.JSON(http.StatusOK, respond.ErrNoDataFound)
		return
	} else if err == mongo.ErrNoDocuments {
		ctx.JSON(http.StatusOK, respond.ErrNoDataFound)
		return
	} else if err != nil {
		ctx.JSON(http.StatusOK, respond.ErrServiceError)
		return
	}
	ctx.JSON(http.StatusOK, respond.ApiSuccess(1, "ok", data))
}
