package man

import (
	"encoding/json"
	"fmt"
	"log"
	"manindexer/mrc20"
	"manindexer/pin"
	"strconv"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/txscript"
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
			transferPinList, _ := CreateMrc20TransferUtxo(pinNode, &validator)
			if len(transferPinList) > 0 {
				mrc20TrasferList = append(mrc20TrasferList, transferPinList...)
			}
		}
	}
	changedTick := make(map[string]int64)
	if len(mrc20UtxoList) > 0 {
		DbAdapter.SaveMrc20Pin(mrc20UtxoList)
		for _, item := range mrc20UtxoList {
			changedTick[item.Mrc20Id] += 1
		}
	}
	if len(mrc20TrasferList) > 0 {
		DbAdapter.UpdateMrc20Utxo(mrc20TrasferList)
		for _, item := range mrc20TrasferList {
			changedTick[item.Mrc20Id] += 1
		}
	}

	if len(deployList) > 0 {
		DbAdapter.SaveMrc20Tick(deployList)
	}
	//update holders,txCount
	for id, txNum := range changedTick {
		go DbAdapter.UpdateMrc20TickHolder(id, txNum)
	}
}

func CreateMrc20DeployPin(pinNode *pin.PinInscription, validator *Mrc20Validator) (mrc20Utxo mrc20.Mrc20Utxo, info mrc20.Mrc20DeployInfo, err error) {
	var df mrc20.Mrc20Deploy
	err = json.Unmarshal(pinNode.ContentBody, &df)
	if err != nil {
		return
	}
	info.Tick = df.Tick
	info.TokenName = df.TokenName
	info.Decimals = df.Decimals
	info.AmtPerMint = df.AmtPerMint
	info.PremineCount, _ = strconv.ParseInt(df.PremineCount, 10, 64)
	info.MintCount, _ = strconv.ParseInt(df.MintCount, 10, 64)
	info.Blockheight = df.Blockheight
	info.Metadata = df.Metadata
	info.DeployType = df.DeployType
	info.Qual = df.Qual
	info.DeployTime = pinNode.Timestamp

	info.Mrc20Id = pinNode.Id
	info.PinNumber = pinNode.Number
	info.Chain = pinNode.ChainName
	info.Address = pinNode.Address
	info.MetaId = pinNode.MetaId
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
	mrc20Utxo.PointValue = pinNode.OutputValue
	err1 := validator.Deploy(pinNode.ContentBody)
	if err1 != nil {
		mrc20Utxo.Verify = false
		mrc20Utxo.Msg = err1.Error()
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
	mrc20Utxo.PointValue = pinNode.OutputValue
	info, shovelList, err1 := validator.Mint(content, pinNode)
	if info != (mrc20.Mrc20DeployInfo{}) {
		mrc20Utxo.Mrc20Id = info.Mrc20Id
		mrc20Utxo.Tick = info.Tick
	}
	if err1 != nil {
		mrc20Utxo.Mrc20Id = info.Mrc20Id
		mrc20Utxo.Verify = false
		mrc20Utxo.Msg = err1.Error()
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
		mrc20UtxoList = sendAllAmountToFirstOutput(pinNode, "Transfer JSON format error")
		return
	}
	//check
	toAddress, utxoList, outputValueList, msg, firstIdx, err1 := validator.Transfer(content, pinNode)
	//if err1 != nil && err1.Error() != "valueErr" {
	if err1 != nil {
		mrc20UtxoList = sendAllAmountToFirstOutput(pinNode, msg)
		return
	}
	// if err1 != nil && err1.Error() == "valueErr" {
	// 	for i, utxo := range utxoList {
	// 		mrc20Utxo := mrc20.Mrc20Utxo{}
	// 		mrc20Utxo.Mrc20Id = utxo.Mrc20Id
	// 		mrc20Utxo.Tick = utxo.Tick
	// 		mrc20Utxo.Verify = true
	// 		mrc20Utxo.PinId = pinNode.Id
	// 		mrc20Utxo.BlockHeight = pinNode.GenesisHeight
	// 		mrc20Utxo.MrcOption = "data-transfer"
	// 		mrc20Utxo.FromAddress = utxo.ToAddress
	// 		mrc20Utxo.ToAddress = pinNode.Address
	// 		mrc20Utxo.Chain = pinNode.ChainName
	// 		mrc20Utxo.Timestamp = pinNode.Timestamp
	// 		mrc20Utxo.TxPoint = fmt.Sprintf("%s:%d", pinNode.GenesisTransaction, pinNode.Offset)
	// 		mrc20Utxo.PinContent = string(pinNode.ContentBody)
	// 		mrc20Utxo.Index = i
	// 		mrc20Utxo.AmtChange = utxo.AmtChange
	// 		mrc20Utxo.Msg = msg
	// 		mrc20Utxo.PointValue = pinNode.OutputValue
	// 		mrc20UtxoList = append(mrc20UtxoList, &mrc20Utxo)
	// 	}
	// 	return
	// }
	address := make(map[string]string)
	name := make(map[string]string)
	inputAmtMap := make(map[string]int64)

	for _, utxo := range utxoList {
		address[utxo.Mrc20Id] = utxo.ToAddress
		name[utxo.Mrc20Id] = utxo.Tick
		//Spent the input UTXO
		amt := utxo.AmtChange * -1
		mrc20Utxo := mrc20.Mrc20Utxo{TxPoint: utxo.TxPoint, Index: utxo.Index, Mrc20Id: utxo.Mrc20Id, Verify: true, Status: -1, AmtChange: amt}
		mrc20UtxoList = append(mrc20UtxoList, &mrc20Utxo)
		inputAmtMap[utxo.Mrc20Id] += utxo.AmtChange
	}
	outputAmtMap := make(map[string]int64)
	x := 0
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
		mrc20Utxo.TxPoint = fmt.Sprintf("%s:%d", pinNode.GenesisTransaction, item.Vout)
		mrc20Utxo.PinContent = string(pinNode.ContentBody)
		mrc20Utxo.Index = x
		mrc20Utxo.PointValue = outputValueList[item.Vout]
		mrc20Utxo.AmtChange, _ = strconv.ParseInt(item.Amount, 10, 64)
		outputAmtMap[item.Id] += mrc20Utxo.AmtChange
		mrc20Utxo.Timestamp = pinNode.Timestamp
		mrc20UtxoList = append(mrc20UtxoList, &mrc20Utxo)
		x += 1
	}
	//Check if the input exceeds the output.
	for id, inputAmt := range inputAmtMap {
		if inputAmt > outputAmtMap[id] {
			find := false
			for _, utxo := range mrc20UtxoList {
				if utxo.Mrc20Id == id && utxo.ToAddress == toAddress[0] {
					utxo.AmtChange += (inputAmt - outputAmtMap[id])
					utxo.Msg = "The total input amount is greater than the output amount"
					find = true
				}
			}
			if find {
				continue
			}
			mrc20Utxo := mrc20.Mrc20Utxo{}
			mrc20Utxo.Mrc20Id = id
			mrc20Utxo.Tick = name[id]
			mrc20Utxo.Verify = true
			mrc20Utxo.PinId = pinNode.Id
			mrc20Utxo.BlockHeight = pinNode.GenesisHeight
			mrc20Utxo.MrcOption = "data-transfer"
			mrc20Utxo.FromAddress = address[id]
			mrc20Utxo.ToAddress = toAddress[0]
			mrc20Utxo.Chain = pinNode.ChainName
			mrc20Utxo.Timestamp = pinNode.Timestamp
			mrc20Utxo.TxPoint = fmt.Sprintf("%s:%d", pinNode.GenesisTransaction, firstIdx)
			mrc20Utxo.PointValue = outputValueList[firstIdx]
			mrc20Utxo.PinContent = string(pinNode.ContentBody)
			mrc20Utxo.Index = x
			mrc20Utxo.AmtChange = inputAmt - outputAmtMap[id]
			mrc20Utxo.Msg = "The total input amount is greater than the output amount"
			mrc20UtxoList = append(mrc20UtxoList, &mrc20Utxo)
			x += 1
		}
	}
	return
}
func sendAllAmountToFirstOutput(pinNode *pin.PinInscription, msg string) (mrc20UtxoList []*mrc20.Mrc20Utxo) {
	tx, err := ChainAdapter[pinNode.ChainName].GetTransaction(pinNode.GenesisTransaction)
	if err != nil {
		log.Println("GetTransaction:", err)
		return
	}
	txb := tx.(*btcutil.Tx)
	toAddress := ""
	idx := 0
	value := int64(0)
	for i, out := range txb.MsgTx().TxOut {
		class, addresses, _, _ := txscript.ExtractPkScriptAddrs(out.PkScript, ChainParams)
		if class.String() != "nulldata" && class.String() != "nonstandard" && len(addresses) > 0 {
			toAddress = addresses[0].String()
			idx = i
			value = out.Value
			break
		}
	}
	if toAddress == "" {
		return
	}
	var inputList []string
	for _, in := range txb.MsgTx().TxIn {
		s := fmt.Sprintf("%s:%d", in.PreviousOutPoint.Hash.String(), in.PreviousOutPoint.Index)
		inputList = append(inputList, s)
	}
	list, err := DbAdapter.GetMrc20UtxoByOutPutList(inputList)
	if err != nil {
		//log.Println("GetMrc20UtxoByOutPutList:", err)
		return
	}
	utxoList := make(map[string]*mrc20.Mrc20Utxo)
	for _, item := range list {
		//Spent the input UTXO
		amt := item.AmtChange * -1
		mrc20Utxo := mrc20.Mrc20Utxo{TxPoint: item.TxPoint, Index: item.Index, Mrc20Id: item.Mrc20Id, Verify: true, Status: -1, AmtChange: amt}
		mrc20UtxoList = append(mrc20UtxoList, &mrc20Utxo)
		if v, ok := utxoList[item.Mrc20Id]; ok {
			v.AmtChange += item.AmtChange
		} else {
			utxoList[item.Mrc20Id] = &mrc20.Mrc20Utxo{
				Mrc20Id:     item.Mrc20Id,
				Tick:        item.Tick,
				Verify:      true,
				PinId:       pinNode.Id,
				BlockHeight: pinNode.GenesisHeight,
				MrcOption:   "data-transfer",
				FromAddress: pinNode.Address,
				ToAddress:   toAddress,
				Chain:       pinNode.ChainName,
				Timestamp:   pinNode.Timestamp,
				TxPoint:     fmt.Sprintf("%s:%d", pinNode.GenesisTransaction, idx),
				PointValue:  value,
				PinContent:  string(pinNode.ContentBody),
				Index:       0,
				AmtChange:   item.AmtChange,
				Msg:         msg,
			}
		}

	}
	for _, mrc20Utxo := range utxoList {
		mrc20UtxoList = append(mrc20UtxoList, mrc20Utxo)
	}
	return
}
func Mrc20NativeTransferHandle(sendList []*mrc20.Mrc20Utxo, reciveAddressList map[string]*string, txPointList map[string]*string) (mrc20UtxoList []mrc20.Mrc20Utxo, err error) {

	return
}
