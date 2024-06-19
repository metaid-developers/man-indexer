package man

import (
	"encoding/json"
	"fmt"
	"manindexer/mrc20"
	"manindexer/pin"
	"strconv"
)

func Mrc20Handle(mrc20List []*pin.PinInscription) {
	validator := Mrc20Validator{}
	var mrc20UtxoList []mrc20.Mrc20Utxo
	var deployList []mrc20.Mrc20DeployInfo
	var mrc20TrasferList []*mrc20.Mrc20Utxo
	for _, pinNode := range mrc20List {
		switch pinNode.Path {
		case "/ft/mrc20/deploy":
			mrc20Pin, info, err := CreateMrc20DeployPin(pinNode, &validator)
			if err == nil {
				mrc20Pin.Chain = pinNode.ChainName
				mrc20UtxoList = append(mrc20UtxoList, mrc20Pin)
				deployList = append(deployList, info)
			}
		case "/ft/mrc20/mint":
			mrc20Pin, err := CreateMrc20MintPin(pinNode, &validator)
			if err == nil {
				mrc20Pin.Chain = pinNode.ChainName
				mrc20UtxoList = append(mrc20UtxoList, mrc20Pin)
			}
		case "/ft/mrc20/transfer":
			transferPinList, err := CreateMrc20TransferUtxo(pinNode, &validator)
			if err == nil && len(transferPinList) > 0 {
				mrc20TrasferList = append(mrc20TrasferList, transferPinList...)
			}
		}
	}
	if len(mrc20UtxoList) > 0 {
		DbAdapter.SaveMrc20Pin(mrc20UtxoList)
	}
	if len(mrc20TrasferList) > 0 {
		DbAdapter.UpdateMrc20Utxo(mrc20TrasferList)
	}

	if len(deployList) > 0 {
		DbAdapter.SaveMrc20Tick(deployList)
	}
}

func CreateMrc20DeployPin(pinNode *pin.PinInscription, validator *Mrc20Validator) (mrc20Utxo mrc20.Mrc20Utxo, info mrc20.Mrc20DeployInfo, err error) {
	err = json.Unmarshal(pinNode.ContentBody, &info)
	if err != nil {
		return
	}
	info.Mrc20Id = pinNode.Id
	info.PinNumber = pinNode.Number
	info.Chain = pinNode.ChainName
	mrc20Utxo.Tick = info.Tick
	mrc20Utxo.Mrc20Id = pinNode.Id
	mrc20Utxo.PinId = pinNode.Id
	mrc20Utxo.BlockHeight = pinNode.GenesisHeight
	mrc20Utxo.MrcOption = "deploy"
	mrc20Utxo.FromAddress = pinNode.Address
	mrc20Utxo.ToAddress = pinNode.Address
	mrc20Utxo.TxPoint = pinNode.Output
	mrc20Utxo.PinContent = string(pinNode.ContentBody)
	mrc20Utxo.Timestamp = pinNode.Timestamp
	err1 := validator.Deploy(pinNode.ContentBody)
	if err1 != nil {
		mrc20Utxo.Verify = false
		mrc20Utxo.ErrorMsg = err1.Error()
	} else {
		mrc20Utxo.Verify = true
	}
	return
}

func CreateMrc20MintPin(pinNode *pin.PinInscription, validator *Mrc20Validator) (mrc20Utxo mrc20.Mrc20Utxo, err error) {
	var content mrc20.Mrc20MintData
	err = json.Unmarshal(pinNode.ContentBody, &content)
	if err != nil {
		return
	}
	mrc20Utxo.Verify = true
	mrc20Utxo.PinId = pinNode.Id
	mrc20Utxo.BlockHeight = pinNode.GenesisHeight
	mrc20Utxo.MrcOption = "mint"
	mrc20Utxo.FromAddress = pinNode.Address
	mrc20Utxo.ToAddress = pinNode.Address
	mrc20Utxo.TxPoint = pinNode.Output
	mrc20Utxo.PinContent = string(pinNode.ContentBody)
	mrc20Utxo.Timestamp = pinNode.Timestamp
	info, shovelList, err1 := validator.Mint(content, pinNode)
	if info != (mrc20.Mrc20DeployInfo{}) {
		mrc20Utxo.Mrc20Id = info.Mrc20Id
		mrc20Utxo.Tick = info.Tick
	}
	if err1 != nil {
		mrc20Utxo.Mrc20Id = info.Mrc20Id
		mrc20Utxo.Verify = false
		mrc20Utxo.ErrorMsg = err1.Error()
	} else {
		DbAdapter.AddMrc20Shovel(shovelList, pinNode.Id)
		DbAdapter.UpdateMrc20TickInfo(info.Mrc20Id, info.TotalMinted+1)
		mrc20Utxo.AmtChange, _ = strconv.ParseInt(info.AmtPerMint, 10, 64)
	}

	return
}
func CreateMrc20TransferUtxo(pinNode *pin.PinInscription, validator *Mrc20Validator) (mrc20UtxoList []*mrc20.Mrc20Utxo, err error) {
	var content []mrc20.Mrc20TranferData
	err = json.Unmarshal(pinNode.ContentBody, &content)
	if err != nil {
		return
	}
	//check
	toAddress, utxoList, err1 := validator.Transfer(content, pinNode)
	if err1 != nil && err1.Error() != "valueErr" {
		return
	}
	if err1 != nil && err1.Error() == "valueErr" {
		for i, utxo := range utxoList {
			mrc20Utxo := mrc20.Mrc20Utxo{}
			mrc20Utxo.Mrc20Id = utxo.Mrc20Id
			mrc20Utxo.Tick = utxo.Tick
			mrc20Utxo.Verify = true
			mrc20Utxo.PinId = pinNode.Id
			mrc20Utxo.BlockHeight = pinNode.GenesisHeight
			mrc20Utxo.MrcOption = "data-transfer"
			mrc20Utxo.FromAddress = utxo.ToAddress
			mrc20Utxo.ToAddress = pinNode.Address
			mrc20Utxo.Chain = pinNode.ChainName
			mrc20Utxo.Timestamp = pinNode.Timestamp
			mrc20Utxo.TxPoint = fmt.Sprintf("%s:%d", pinNode.GenesisTransaction, pinNode.Offset)
			mrc20Utxo.PinContent = string(pinNode.ContentBody)
			mrc20Utxo.Index = i
			mrc20Utxo.AmtChange = utxo.AmtChange
			mrc20UtxoList = append(mrc20UtxoList, &mrc20Utxo)
		}
		return
	}
	address := make(map[string]string)
	name := make(map[string]string)
	for _, utxo := range utxoList {
		address[utxo.Mrc20Id] = utxo.ToAddress
		name[utxo.Mrc20Id] = utxo.Tick
		mrc20Utxo := mrc20.Mrc20Utxo{TxPoint: utxo.TxPoint, Index: utxo.Index, Mrc20Id: utxo.Mrc20Id, Verify: true, Status: -1}
		mrc20UtxoList = append(mrc20UtxoList, &mrc20Utxo)
	}
	for _, item := range content {
		mrc20Utxo := mrc20.Mrc20Utxo{}
		mrc20Utxo.Mrc20Id = item.Id
		mrc20Utxo.Tick = name[item.Id]
		mrc20Utxo.Verify = true
		mrc20Utxo.PinId = pinNode.Id
		mrc20Utxo.BlockHeight = pinNode.GenesisHeight
		mrc20Utxo.MrcOption = "data-transfer"
		mrc20Utxo.FromAddress = address[item.Id]
		mrc20Utxo.ToAddress = toAddress[item.Vout]
		mrc20Utxo.Chain = pinNode.ChainName
		mrc20Utxo.TxPoint = fmt.Sprintf("%s:%d", pinNode.GenesisTransaction, pinNode.Offset)
		mrc20Utxo.PinContent = string(pinNode.ContentBody)
		mrc20Utxo.Index = item.Vout
		mrc20Utxo.AmtChange = item.Amount
		mrc20Utxo.Timestamp = pinNode.Timestamp
		mrc20UtxoList = append(mrc20UtxoList, &mrc20Utxo)
	}
	return
}
func Mrc20NativeTransferHandle(sendList []*mrc20.Mrc20Utxo, reciveAddressList map[string]*string, txPointList map[string]*string) (mrc20UtxoList []mrc20.Mrc20Utxo, err error) {

	return
}
