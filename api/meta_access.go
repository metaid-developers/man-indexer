package api

import (
	"encoding/json"
	"manindexer/api/respond"
	"manindexer/basicprotocols/metaaccess"
	"manindexer/basicprotocols/metaso"
	"manindexer/common"
	"manindexer/database/mongodb"
	"manindexer/man"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/mongo"
)

func metaAccessJsonApi(r *gin.Engine) {
	accessGroup := r.Group("/api/access")
	accessGroup.Use(CorsMiddleware())
	accessGroup.POST("/decrypt", accessContentDecrypt)
	accessGroup.GET("/getPubKey", getPubKey)
	accessGroup.GET("/getControlByContentPin", getControlByContentPin)
}
func getPubKey(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, respond.ApiSuccess(1, "ok", common.Config.MetaSo.Pubkey))
}

type decryptReq struct {
	Address      string `json:"address"`
	Timestamp    int64  `json:"timestamp"`
	PublicKey    string `json:"publicKey"`
	Sign         string `json:"sign"`
	PinId        string `json:"pinId"`
	ControlPath  string `json:"controlPath"`
	ControlPinId string `json:"controlPinId"`
}

func accessContentDecrypt(ctx *gin.Context) {
	var req decryptReq
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, respond.ErrParameterError)
		return
	}
	err := metaaccess.CheckSign(req.PublicKey, common.Config.MetaSo.Prikey, req.Timestamp, req.Address, req.Sign)
	if err != nil {
		ctx.JSON(http.StatusOK, respond.ApiError(401, "sign check error."))
		return
	}
	control, err := man.DbAdapter.GetControlById(req.ControlPinId, false)
	if err != nil {
		ctx.JSON(http.StatusOK, respond.ApiError(401, "control Pin null."))
		return
	}
	//control.PayCheck
	status := "purchased"
	if control.PayCheck != nil && *control.PayCheck != (metaaccess.AccessControlPayCheck{}) {
		passData, err := man.DbAdapter.CheckAccessPass(req.Address, req.PinId, req.ControlPath)
		if err != nil || passData == nil {
			find, _ := mongodb.CheckAccessPassInMempool(req.Address, req.ControlPinId)
			if find {
				status = "mempool"
			} else {
				status = "unpurchased"
			}
			ctx.JSON(http.StatusOK, respond.ApiSuccess(1, "ok", gin.H{"status": status, "contentResult": nil, "filesResult": nil}))
			//ctx.JSON(http.StatusOK, respond.ApiError(404, "pass check error."))
			return
		}
	}
	if control.HoldCheck != nil && *control.HoldCheck != (metaaccess.AccessControlHoldCheck{}) {
		if control.HoldCheck.AccType == "mrc20" && control.HoldCheck.Ticker != "" {
			totalAmt, err := mongodb.GetTickBalance(control.HoldCheck.Ticker, req.Address)
			if err != nil {
				ctx.JSON(http.StatusOK, respond.ApiError(401, "hold check,getTickBalance error"))
				return
			}
			passAmt, err := decimal.NewFromString(control.HoldCheck.Amount)
			if err != nil {
				ctx.JSON(http.StatusOK, respond.ApiError(401, "hold check,Amount value error"))
				return
			}
			if totalAmt.Cmp(passAmt) == -1 {
				ctx.JSON(http.StatusOK, respond.ApiError(402, "Not enough ticks held."))
				return
			}
		}

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
	var data metaso.PayBuzz
	err = json.Unmarshal(pinNode.ContentBody, &data)
	if err != nil {
		ctx.JSON(http.StatusOK, respond.ApiError(404, "paybuzz json check error."))
		return
	}
	var encryptFiles [][]byte
	if len(data.EncryptFiles) > 0 {
		for _, item := range data.EncryptFiles {
			pinId := strings.ReplaceAll(item, "metafile://", "")
			pinNode, err := man.DbAdapter.GetPinByNumberOrId(pinId)
			if err == nil {
				encryptFiles = append(encryptFiles, pinNode.ContentBody)
			}
		}
	}
	contentResult, filesResult, err := metaaccess.DecryptionPin([]byte(data.EncryptContent), encryptFiles, common.Config.MetaSo.Prikey, control.CreatorPubkey, control.EncryptedKey)
	if err != nil {
		ctx.JSON(http.StatusOK, respond.ApiError(404, "decryptionPin fail."+err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, respond.ApiSuccess(1, "ok", gin.H{"status": status, "contentResult": string(contentResult), "filesResult": filesResult}))
}

func getControlByContentPin(ctx *gin.Context) {
	if ctx.Query("pinId") == "" {
		ctx.JSON(http.StatusOK, respond.ApiError(404, "pinId is null."))
		return
	}
	data, err := man.DbAdapter.GetControlById(ctx.Query("pinId"), true)
	if err != nil {
		ctx.JSON(http.StatusOK, respond.ApiError(404, "no data"))
		return
	}
	ctx.JSON(http.StatusOK, respond.ApiSuccess(1, "ok", data))
}
