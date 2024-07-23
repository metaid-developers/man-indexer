package mrc20_service

import (
	"encoding/json"
	"errors"
	"github.com/shopspring/decimal"
)

type Mrc20DataItem struct {
	Id     string `json:"id"`
	Amount string `json:"amount"`
	Vout   int64  `json:"vout"`
}

func MakeTransferPayload(tickId string, transferMrc20s []*TransferMrc20, mrc20Outs []*Mrc20OutInfo) (string, error) {
	var (
		payload       string           = ""
		totalAmountDe decimal.Decimal  = decimal.New(0, 0)
		dataItems     []*Mrc20DataItem = make([]*Mrc20DataItem, 0)
	)
	for _, v := range transferMrc20s {
		mrc20AmountDe, err := decimal.NewFromString(v.Mrc20Amount)
		if err != nil {
			return "", err
		}
		totalAmountDe = totalAmountDe.Add(mrc20AmountDe)
	}

	for i, v := range mrc20Outs {
		dataItem := &Mrc20DataItem{
			Id:     tickId,
			Amount: v.Amount,
			Vout:   int64(i + 1),
		}
		dataItems = append(dataItems, dataItem)
	}

	if result, err := json.Marshal(dataItems); err != nil {
		return "", errors.New("Json Str Parse Err: " + err.Error())
	} else {
		payload = string(result)
	}

	return payload, nil
}

type Mrc20DeployData struct {
	Tick         string      `json:"tick"`
	TokenName    string      `json:"tokenName"`
	Decimals     string      `json:"decimals"`
	AmtPerMint   string      `json:"amtPerMint"`
	MintCount    string      `json:"mintCount"`
	PremineCount string      `json:"premineCount"`
	BeginBlock   string      `json:"beginBlock"`
	EndBlock     string      `json:"endBlock"`
	Metadata     string      `json:"metadata"`
	PinCheck     interface{} `json:"pinCheck"`
	PayCheck     *PayCheck   `json:"payCheck"`
}

type PayCheck struct {
	PayAmount string `json:"payAmount"`
	PayTo     string `json:"payTo"`
}

func MakeDeployPayloadForIdCoins(tick, tokenName, metadata, payTo, payAmount, mintCount, amountPerMint, premineCount, beginBlock, endBlock, decimals string,
	pinCheckCreator, pinCheckPath, pinCheckCount, pinCheckLvl string) (string, *Mrc20DeployData, int64) {
	var (
		payload         string           = ""
		totalSupply     int64            = 0
		mrc20DeployData *Mrc20DeployData = &Mrc20DeployData{
			Tick:         tick,
			TokenName:    tokenName,
			Decimals:     decimals,
			AmtPerMint:   amountPerMint,
			MintCount:    mintCount,
			PremineCount: premineCount,
			BeginBlock:   beginBlock,
			EndBlock:     endBlock,
			Metadata:     metadata,
			PinCheck: map[string]interface{}{
				"creator": pinCheckCreator,
				"path":    pinCheckPath,
				"lvl":     pinCheckLvl,
				"count":   pinCheckCount,
			},
			PayCheck: &PayCheck{
				PayAmount: payAmount,
				PayTo:     payTo,
			},
		}
	)

	amtPerMintDe, _ := decimal.NewFromString(mrc20DeployData.AmtPerMint)
	mintCountDe, _ := decimal.NewFromString(mrc20DeployData.MintCount)
	totalSupply = amtPerMintDe.Mul(mintCountDe).IntPart()

	payloadByte, _ := json.Marshal(mrc20DeployData)
	payload = string(payloadByte)
	return payload, mrc20DeployData, totalSupply
}
