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
