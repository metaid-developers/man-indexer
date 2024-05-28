package man

import (
	"encoding/json"
	"manindexer/mrc20"
	"manindexer/pin"
	"strconv"
)

func Mrc20Handle(mrc20List []*pin.PinInscription) {
	validator := Mrc20Validator{}
	var mrc20PinList []mrc20.Mrc20Pin
	var deployList []mrc20.Mrc20DeployInfo
	for _, pinNode := range mrc20List {
		switch pinNode.Path {
		case "/mrc20/deploy":
			mrc20Pin, info, err := CreateMrc20DeployPin(pinNode, &validator)
			if err == nil {
				mrc20PinList = append(mrc20PinList, mrc20Pin)
				deployList = append(deployList, info)
			}
		case "/mrc20/mint":
			mrc20Pin, err := CreateMrc20MintPin(pinNode, &validator)
			if err == nil {
				mrc20PinList = append(mrc20PinList, mrc20Pin)
			}
		case "/mrc20/transfer":
			transferPinList, err := CreateMrc20TransferPin(pinNode, &validator)
			if err == nil {
				mrc20PinList = append(mrc20PinList, transferPinList...)
			}
		}
	}
	if len(mrc20PinList) > 0 {
		DbAdapter.SaveMrc20Pin(mrc20PinList)
	}
	if len(deployList) > 0 {
		DbAdapter.SaveMrc20Tick(deployList)
	}
}

func CreateMrc20DeployPin(pinNode *pin.PinInscription, validator *Mrc20Validator) (mrc20Pin mrc20.Mrc20Pin, info mrc20.Mrc20DeployInfo, err error) {
	err = json.Unmarshal(pinNode.ContentBody, &info)
	if err != nil {
		return
	}
	info.Mrc20Id = pinNode.Id
	info.PinNumber = pinNode.Number
	mrc20Pin.Tick = info.Tick
	mrc20Pin.Mrc20Id = pinNode.Id
	mrc20Pin.PinId = pinNode.Id
	mrc20Pin.PinNum = pinNode.Number
	mrc20Pin.BlockHeight = pinNode.GenesisHeight
	mrc20Pin.MrcOption = "deploy"
	mrc20Pin.FromAddress = pinNode.Address
	mrc20Pin.ToAddress = pinNode.Address
	mrc20Pin.PinTxHash = pinNode.GenesisTransaction
	mrc20Pin.Content = string(pinNode.ContentBody)
	err1 := validator.Deploy(pinNode.ContentBody)
	if err1 != nil {
		mrc20Pin.Verify = false
		mrc20Pin.ErrorMsg = err1.Error()
	} else {
		mrc20Pin.Verify = true
	}
	return
}

func CreateMrc20MintPin(pinNode *pin.PinInscription, validator *Mrc20Validator) (mrc20Pin mrc20.Mrc20Pin, err error) {
	var content mrc20.Mrc20MintData
	err = json.Unmarshal(pinNode.ContentBody, &content)
	if err != nil {
		return
	}
	mrc20Pin.Verify = true
	mrc20Pin.PinId = pinNode.Id
	mrc20Pin.PinNum = pinNode.Number
	mrc20Pin.BlockHeight = pinNode.GenesisHeight
	mrc20Pin.MrcOption = "mint"
	mrc20Pin.FromAddress = pinNode.Address
	mrc20Pin.ToAddress = pinNode.Address
	mrc20Pin.PinTxHash = pinNode.GenesisTransaction
	mrc20Pin.Content = string(pinNode.ContentBody)
	info, shovel, err1 := validator.Mint(content, pinNode)
	if info != (mrc20.Mrc20DeployInfo{}) {
		mrc20Pin.Mrc20Id = info.Mrc20Id
		mrc20Pin.Tick = info.Tick
	}
	if err1 != nil {
		mrc20Pin.Mrc20Id = info.Mrc20Id
		mrc20Pin.Verify = false
		mrc20Pin.ErrorMsg = err1.Error()
	} else {
		DbAdapter.AddMrc20Shovel(shovel, pinNode.Id)
		DbAdapter.UpdateMrc20TickInfo(info.Mrc20Id, info.TotalMinted+1)
		mrc20Pin.AmtChange, _ = strconv.ParseInt(info.AmtPerMint, 10, 64)
	}

	return
}
func CreateMrc20TransferPin(pinNode *pin.PinInscription, validator *Mrc20Validator) (mrc20PinList []mrc20.Mrc20Pin, err error) {
	var content []mrc20.Mrc20TranferData
	err = json.Unmarshal(pinNode.ContentBody, &content)
	if err != nil {
		return
	}
	mrc20Pin := mrc20.Mrc20Pin{}
	mrc20Pin.Verify = true
	mrc20Pin.PinId = pinNode.Id
	mrc20Pin.PinNum = pinNode.Number
	mrc20Pin.BlockHeight = pinNode.GenesisHeight
	mrc20Pin.MrcOption = "transfer"
	mrc20Pin.FromAddress = pinNode.Address
	mrc20Pin.ToAddress = pinNode.Address
	mrc20Pin.PinTxHash = pinNode.GenesisTransaction
	mrc20Pin.Content = string(pinNode.ContentBody)
	tickNameMap, err1 := validator.Transfer(content, pinNode)
	if err1 != nil {
		mrc20Pin.Verify = false
		mrc20Pin.ErrorMsg = err1.Error()
		for k, v := range tickNameMap {
			mrc20Pin.Mrc20Id = k
			mrc20Pin.Tick = v
		}
		mrc20PinList = append(mrc20PinList, mrc20Pin)
		return
	}
	for _, item := range content {
		send := mrc20Pin
		recive := mrc20Pin
		v, _ := strconv.ParseInt(item.Amount, 10, 64)
		send.AmtChange = -1 * v
		send.Tick = tickNameMap[item.Id]
		send.Mrc20Id = item.Id
		recive.AmtChange = v
		recive.ToAddress = item.Addr
		recive.Mrc20Id = item.Id
		recive.Tick = tickNameMap[item.Id]
		mrc20PinList = append(mrc20PinList, send)
		mrc20PinList = append(mrc20PinList, recive)
	}
	return
}
