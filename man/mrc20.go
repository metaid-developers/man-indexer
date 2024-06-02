package man

import (
	"encoding/json"
	"manindexer/mrc20"
	"manindexer/pin"
	"strconv"
)

func Mrc20Handle(mrc20List []*pin.PinInscription) {
	validator := Mrc20Validator{}
	var mrc20UtxoList []mrc20.Mrc20Utxo
	var deployList []mrc20.Mrc20DeployInfo
	for _, pinNode := range mrc20List {
		switch pinNode.Path {
		case "/ft/mrc20/deploy":
			mrc20Pin, info, err := CreateMrc20DeployPin(pinNode, &validator)
			if err == nil {
				mrc20UtxoList = append(mrc20UtxoList, mrc20Pin)
				deployList = append(deployList, info)
			}
		case "/ft/mrc20/mint":
			mrc20Pin, err := CreateMrc20MintPin(pinNode, &validator)
			if err == nil {
				mrc20UtxoList = append(mrc20UtxoList, mrc20Pin)
			}
		case "/ft/mrc20/transfer":
			transferPinList, err := CreateMrc20TransferUtxo(pinNode, &validator)
			if err == nil {
				mrc20UtxoList = append(mrc20UtxoList, transferPinList...)
			}
		}
	}
	if len(mrc20UtxoList) > 0 {
		DbAdapter.SaveMrc20Pin(mrc20UtxoList)
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
	mrc20Utxo.Tick = info.Tick
	mrc20Utxo.Mrc20Id = pinNode.Id
	mrc20Utxo.PinId = pinNode.Id
	mrc20Utxo.BlockHeight = pinNode.GenesisHeight
	mrc20Utxo.MrcOption = "deploy"
	mrc20Utxo.FromAddress = pinNode.Address
	mrc20Utxo.ToAddress = pinNode.Address
	mrc20Utxo.TxPoint = pinNode.Output
	mrc20Utxo.PinContent = string(pinNode.ContentBody)
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
	info, shovel, err1 := validator.Mint(content, pinNode)
	if info != (mrc20.Mrc20DeployInfo{}) {
		mrc20Utxo.Mrc20Id = info.Mrc20Id
		mrc20Utxo.Tick = info.Tick
	}
	if err1 != nil {
		mrc20Utxo.Mrc20Id = info.Mrc20Id
		mrc20Utxo.Verify = false
		mrc20Utxo.ErrorMsg = err1.Error()
	} else {
		DbAdapter.AddMrc20Shovel(shovel, pinNode.Id)
		DbAdapter.UpdateMrc20TickInfo(info.Mrc20Id, info.TotalMinted+1)
		mrc20Utxo.AmtChange, _ = strconv.ParseInt(info.AmtPerMint, 10, 64)
	}

	return
}
func CreateMrc20TransferUtxo(pinNode *pin.PinInscription, validator *Mrc20Validator) (mrc20UtxoList []mrc20.Mrc20Utxo, err error) {
	var content []mrc20.Mrc20TranferData
	err = json.Unmarshal(pinNode.ContentBody, &content)
	if err != nil {
		return
	}
	mrc20Utxo := mrc20.Mrc20Utxo{}
	mrc20Utxo.Verify = true
	mrc20Utxo.PinId = pinNode.Id
	mrc20Utxo.BlockHeight = pinNode.GenesisHeight
	mrc20Utxo.MrcOption = "transfer"
	mrc20Utxo.FromAddress = pinNode.Address
	mrc20Utxo.ToAddress = pinNode.Address
	//TODO
	mrc20Utxo.TxPoint = pinNode.GenesisTransaction
	mrc20Utxo.PinContent = string(pinNode.ContentBody)
	tickNameMap, err1 := validator.Transfer(content, pinNode)
	if err1 != nil {
		mrc20Utxo.Verify = false
		mrc20Utxo.ErrorMsg = err1.Error()
		for k, v := range tickNameMap {
			mrc20Utxo.Mrc20Id = k
			mrc20Utxo.Tick = v
		}
		mrc20UtxoList = append(mrc20UtxoList, mrc20Utxo)
		return
	}
	for _, item := range content {
		send := mrc20Utxo
		recive := mrc20Utxo
		v, _ := strconv.ParseInt(item.Amount, 10, 64)
		send.AmtChange = -1 * v
		send.Tick = tickNameMap[item.Id]
		send.Mrc20Id = item.Id
		recive.AmtChange = v
		recive.ToAddress = item.Addr
		recive.Mrc20Id = item.Id
		recive.Tick = tickNameMap[item.Id]
		mrc20UtxoList = append(mrc20UtxoList, send)
		mrc20UtxoList = append(mrc20UtxoList, recive)
	}
	return
}
func Mrc20NativeTransferHandle(sendList []*mrc20.Mrc20Utxo, reciveAddressList map[string]*string, txPointList map[string]*string) (mrc20UtxoList []mrc20.Mrc20Utxo, err error) {

	return
}
