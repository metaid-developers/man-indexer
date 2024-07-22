package cli

import (
	"errors"
	"manindexer/common"
	"manindexer/inscribe/mrc20_service"
	"manindexer/man"
	"manindexer/mrc20"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/shopspring/decimal"
)

func getNetParams() *chaincfg.Params {
	if common.TestNet == "1" {
		return &chaincfg.TestNet3Params
	} else if common.TestNet == "2" {
		return &chaincfg.RegressionNetParams
	} else {
		return &chaincfg.MainNetParams
	}
}

func getMrc20Utxos(address, tickId, needAmount string) ([]*mrc20_service.TransferMrc20, error) {
	list, total, err := man.DbAdapter.GetHistoryByAddress(tickId, address, 0, 1000, "0", "true")
	if err != nil {
		return nil, err
	}
	if total == 0 {
		return nil, errors.New("no mrc20 utxo")
	}
	needAmountDe, err := decimal.NewFromString(needAmount)
	if err != nil {
		return nil, err
	}
	pkScript, err := mrc20_service.AddressToPkScript(getNetParams(), address)
	if err != nil {
		return nil, err
	}
	utxos := make([]*mrc20_service.TransferMrc20, 0)
	totalAmountDe := decimal.NewFromInt(0)
	for _, v := range list {
		txPoint := v.TxPoint
		txPointStrs := strings.Split(txPoint, ":")
		if len(txPointStrs) != 2 {
			continue
		}
		txId := txPointStrs[0]
		index, _ := strconv.ParseInt(txPointStrs[1], 10, 64)
		item := &mrc20_service.TransferMrc20{
			PrivateKeyHex: wallet.GetPrivateKey(),
			Address:       wallet.GetAddress(),
			RedeemScript:  "",
			PkScript:      pkScript,
			OutRaw:        "",
			UtxoTxId:      txId,
			UtxoIndex:     uint32(index),
			UtxoOutValue:  v.PointValue,
			Mrc20Amount:   v.AmtChange.String(),
			Mrc20TickerId: v.Mrc20Id,
		}
		utxos = append(utxos, item)
		totalAmountDe = totalAmountDe.Add(v.AmtChange)
		if totalAmountDe.GreaterThanOrEqual(needAmountDe) {
			break
		}
	}
	if totalAmountDe.LessThan(needAmountDe) {
		return nil, errors.New("insufficient mrc20 utxo")
	}
	return utxos, nil
}

func getShovels(address, tickId string) ([]*mrc20_service.MintPin, []*mrc20_service.PayTo, error) {
	payTos := make([]*mrc20_service.PayTo, 0)
	info, err := man.DbAdapter.GetMrc20TickInfo(tickId, "")
	if err != nil {
		return nil, nil, err
	}
	lv := int(0)
	path := ""
	query := ""
	key := ""
	value := ""
	operator := ""
	count := int64(0)
	if info.PinCheck.Lv != "" {
		lv, _ = strconv.Atoi(info.PinCheck.Lv)
	}
	if info.PinCheck.Path != "" {
		path, query, key, operator, value = mrc20.PathParse(info.PinCheck.Path)
		if path != "" && query != "" {
			if key == "" && operator == "" && value == "" {
				query = query[2 : len(query)-2]
			}
		} else if path == "" {
			path = info.PinCheck.Path
		}
	}
	if info.PayCheck.PayTo != "" {
		payAddress := info.PayCheck.PayTo
		payAmount, _ := strconv.ParseInt(info.PayCheck.PayAmount, 10, 64)
		if payAmount < 546 {
			payAmount = 546
		}
		payTos = append(payTos, &mrc20_service.PayTo{
			Amount:  payAmount,
			Address: payAddress,
		})
	}
	count, _ = strconv.ParseInt(info.PinCheck.Count, 10, 64)
	if count == 0 {
		return nil, payTos, nil
	}
	list, total, err := man.DbAdapter.GetShovelListByAddress(address, tickId, info.PinCheck.Creator, lv, path, query, key, operator, value, 0, 1000)
	if err != nil {
		return nil, nil, err
	}
	if count > total {
		return nil, nil, errors.New("insufficient shovel list")
	}
	pkScript, err := mrc20_service.AddressToPkScript(getNetParams(), address)
	if err != nil {
		return nil, nil, err
	}
	shovelList := make([]*mrc20_service.MintPin, 0)
	for i, v := range list {
		location := v.Location
		locationStrs := strings.Split(location, ":")
		if len(locationStrs) != 3 {
			continue
		}
		txId := locationStrs[0]
		index, _ := strconv.ParseInt(locationStrs[1], 10, 64)

		item := &mrc20_service.MintPin{
			PinId:           v.Id,
			PinUtxoTxId:     txId,
			PinUtxoIndex:    uint32(index),
			PinUtxoOutValue: v.OutputValue,
			PrivateKeyHex:   wallet.GetPrivateKey(),
			Address:         wallet.GetAddress(),
			RedeemScript:    "",
			PkScript:        pkScript,
			OutRaw:          "",
		}
		shovelList = append(shovelList, item)
		if i == int(count-1) {
			break
		}
	}
	return shovelList, payTos, nil
}
