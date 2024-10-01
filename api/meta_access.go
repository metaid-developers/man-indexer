package api

import (
	"manindexer/api/respond"
	"manindexer/common"
	"manindexer/man"
	"manindexer/metaaccess"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func metaAccessJsonApi(r *gin.Engine) {
	accessGroup := r.Group("/api/access")
	accessGroup.Use(CorsMiddleware())
	accessGroup.POST("/decrypt", accessContentDecrypt)
	accessGroup.GET("/getPubKey", getPubKey)
}
func getPubKey(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, respond.ApiSuccess(1, "ok", common.Config.MetaSo.Pubkey))
}

type decryptReq struct {
	Address     string `json:"address"`
	Timestamp   int64  `json:"timestamp"`
	Sign        string `json:"sign"`
	PinId       string `json:"pinId"`
	ControlPath string `json:"controlPath"`
}

func accessContentDecrypt(ctx *gin.Context) {
	var req decryptReq
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, respond.ErrParameterError)
		return
	}
	passData, err := man.DbAdapter.CheckAccessPass(req.Address, req.PinId, req.ControlPath)
	if err != nil || passData == nil {
		ctx.JSON(http.StatusOK, respond.ApiError(404, "pass check error."))
		return
	}
	pinNode, err := man.DbAdapter.GetPinByNumberOrId(req.PinId)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusOK, respond.ErrNoPinFound)
		} else {
			ctx.JSON(http.StatusOK, respond.ErrServiceError)
		}
		return
	}
	result, err := metaaccess.DecryptionPin(pinNode.ContentBody, common.Config.MetaSo.Prikey, passData.CreatorPubkey, passData.EncryptedKey)
	if err != nil {
		ctx.JSON(http.StatusOK, respond.ApiError(404, "decryptionPin fail."))
		return
	}
	ctx.JSON(http.StatusOK, respond.ApiSuccess(1, "ok", string(result)))
}
