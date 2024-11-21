package man

import (
	"encoding/json"
	"errors"
	"manindexer/basicprotocols/metaaccess"
	"manindexer/pin"
	"strconv"
)

type MetaAccessValidator struct {
}

func (validator *MetaAccessValidator) AccessControl(pinNode *pin.PinInscription) (data metaaccess.AccessControl, err error) {
	err = json.Unmarshal(pinNode.ContentBody, &data)
	if err != nil {
		return
	}
	if data.ManPubkey == "" || data.CreatorPubkey == "" || data.EncryptedKey == "" {
		err = errors.New(metaaccess.ErrPinContent + ",key null")
		return
	}
	if data.PayCheck != nil && *data.PayCheck != (metaaccess.AccessControlPayCheck{}) {
		if data.PayCheck.AccType == "" || data.PayCheck.Amount == "" || data.PayCheck.PayTo == "" {
			err = errors.New(metaaccess.ErrPinContent + ",payCheck error")
			return
		}
		if data.PayCheck.AccType != "chainCoin" && data.PayCheck.Ticker == "" {
			err = errors.New(metaaccess.ErrPinContent + ",payCheck ticker error")
			return
		}
		if data.ControlPath != "" && data.PayCheck.ValidPeriod == "" {
			err = errors.New(metaaccess.ErrPinContent + ",payCheck validPeriod error")
			return
		}
		var payAmt float64
		payAmt, err = strconv.ParseFloat(data.PayCheck.Amount, 64)
		if err != nil {
			err = errors.New(metaaccess.ErrPinContent + ",payAmt error")
			return
		}
		if payAmt < 0 {
			err = errors.New(metaaccess.ErrPinContent + ",payAmt error")
			return
		}
	}
	if data.HoldCheck != nil && *data.HoldCheck != (metaaccess.AccessControlHoldCheck{}) {
		if data.HoldCheck.AccType == "" || data.HoldCheck.Amount == "" {
			err = errors.New(metaaccess.ErrPinContent + ",holdCheck error")
			return
		}
		if data.HoldCheck.AccType != "chainCoin" && data.HoldCheck.Ticker == "" {
			err = errors.New(metaaccess.ErrPinContent + ",holdCheck tiker error")
			return
		}
		var holdAmt float64
		holdAmt, err = strconv.ParseFloat(data.HoldCheck.Amount, 64)
		if err != nil {
			err = errors.New(metaaccess.ErrPinContent + ",holdAmt error")
			return
		}
		if holdAmt < 0 {
			err = errors.New(metaaccess.ErrPinContent + ",holdAmt error")
			return
		}
	}
	if data.HoldCheck != nil && *data.HoldCheck == (metaaccess.AccessControlHoldCheck{}) && *data.PayCheck == (metaaccess.AccessControlPayCheck{}) {
		err = errors.New(metaaccess.ErrPinContent + ",check null")
		return
	}
	data.PinId = pinNode.Id
	data.Address = pinNode.Address
	data.MetaId = pinNode.MetaId
	return
}
func (validator *MetaAccessValidator) AccessPass(pinNode *pin.PinInscription) (data []metaaccess.AccessPassData, err error) {
	//check content
	var pass metaaccess.AccessPass
	err = json.Unmarshal(pinNode.ContentBody, &pass)
	if err != nil {
		return
	}
	if pass.AccessControlID == "" {
		err = errors.New(metaaccess.ErrPinContent)
		return
	}
	//get accesscontrol info
	info, err := DbAdapter.GetControlById(pass.AccessControlID, false)
	if err != nil {
		err = errors.New(metaaccess.ErrGetContronlPin)
		return
	}
	//payCheck
	//info.PayCheck.Amount
	// txResult, err := ChainAdapter[pinNode.ChainName].GetTransaction(pinNode.GenesisTransaction)
	// if err != nil {
	// 	err = errors.New(metaaccess.ErrGetPassTx)
	// 	return
	// }
	// tx := txResult.(*btcutil.Tx)

	for _, contentPinId := range info.ControlPins {
		data = append(data, validator.createPassData(info, "", contentPinId, pinNode))
	}
	if info.ControlPath != "" {
		data = append(data, validator.createPassData(info, info.ControlPath, "", pinNode))
	}
	return
}
func (validator *MetaAccessValidator) createPassData(info *metaaccess.AccessControl, controlPath string, contentPinId string, pinNode *pin.PinInscription) (data metaaccess.AccessPassData) {
	if info == nil {
		return
	}
	if info.PayCheck != nil && info.HoldCheck != nil && *info.PayCheck != (metaaccess.AccessControlPayCheck{}) && *info.HoldCheck != (metaaccess.AccessControlHoldCheck{}) {
		data.CheckMode = "payAndHold"
		data.ValidPeriod, _ = strconv.ParseInt(info.PayCheck.ValidPeriod, 10, 64)
	} else if info.PayCheck != nil && *info.PayCheck != (metaaccess.AccessControlPayCheck{}) {
		data.CheckMode = "pay"
		data.ValidPeriod, _ = strconv.ParseInt(info.PayCheck.ValidPeriod, 10, 64)
	} else if info.HoldCheck != nil && *info.HoldCheck != (metaaccess.AccessControlHoldCheck{}) {
		data.CheckMode = "hold"
	}
	data.PinId = pinNode.Id
	data.CreatorAddress = info.Address
	data.BuyerAddress = pinNode.Address
	data.CreatorMetaId = info.MetaId
	data.BuyerMetaId = pinNode.MetaId
	data.ControlId = info.PinId
	data.ContentPinId = contentPinId
	data.ControlPath = controlPath
	data.ManPubkey = info.ManPubkey
	data.CreatorPubkey = info.CreatorPubkey
	data.EncryptedKey = info.EncryptedKey
	return
}
