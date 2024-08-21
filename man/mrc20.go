package man

import (
	"encoding/json"
	"fmt"
	"log"
	"manindexer/mrc20"
	"manindexer/pin"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/txscript"
	"github.com/shopspring/decimal"
)

func Mrc20Handle(chainName string, height int64, mrc20List []*pin.PinInscription, mrc20TransferPinTx map[string]struct{}, txInList []string, isMempool bool) {
	validator := Mrc20Validator{}
	var mrc20UtxoList []mrc20.Mrc20Utxo

	var mrc20TrasferList []*mrc20.Mrc20Utxo
	//var deployHandleList []*pin.PinInscription
	var mintHandleList []*pin.PinInscription
	var transferHandleList []*pin.PinInscription
	for _, pinNode := range mrc20List {
		switch pinNode.Path {
		case "/ft/mrc20/deploy":
			//deployHandleList = append(deployHandleList, pinNode)
			//Prioritize handling deploy
			deployResult := deployHandle(pinNode)
			if len(deployResult) > 0 {
				mrc20UtxoList = append(mrc20UtxoList, deployResult...)
			}
		case "/ft/mrc20/mint":
			mintHandleList = append(mintHandleList, pinNode)
		case "/ft/mrc20/transfer":
			transferHandleList = append(transferHandleList, pinNode)
		}
	}

	for _, pinNode := range mintHandleList {
		mrc20Pin, err := CreateMrc20MintPin(pinNode, &validator, false)
		if err == nil {
			mrc20Pin.Chain = pinNode.ChainName
			mrc20UtxoList = append(mrc20UtxoList, mrc20Pin)
		}
	}
	changedTick := make(map[string]int64)
	if len(mrc20UtxoList) > 0 {
		DbAdapter.SaveMrc20Pin(mrc20UtxoList)
		for _, item := range mrc20UtxoList {
			if item.MrcOption != "deploy" {
				changedTick[item.Mrc20Id] += 1
			}
		}
	}

	//CatchNativeMrc20Transfer
	handleNativTransfer(chainName, height, mrc20TransferPinTx, txInList, isMempool)
	// mrc20transferCheck, err := DbAdapter.GetMrc20UtxoByOutPutList(txInList, isMempool)
	// if err == nil && len(mrc20transferCheck) > 0 {
	// 	mrc20TrasferList := IndexerAdapter[chainName].CatchNativeMrc20Transfer(height, mrc20transferCheck, mrc20TransferPinTx)
	// 	if len(mrc20TrasferList) > 0 {
	// 		DbAdapter.UpdateMrc20Utxo(mrc20TrasferList, isMempool)
	// 	}
	// }

	mrc20TrasferList = transferHandle(transferHandleList)
	if len(mrc20TrasferList) > 0 {
		//DbAdapter.UpdateMrc20Utxo(mrc20TrasferList, false)
		for _, item := range mrc20TrasferList {
			if item.MrcOption != "deploy" {
				changedTick[item.Mrc20Id] += 1
			}
		}
	}
	//CatchNativeMrc20Transfer Agin
	handleNativTransfer(chainName, height, mrc20TransferPinTx, txInList, isMempool)
	//update holders,txCount
	for id, txNum := range changedTick {
		go DbAdapter.UpdateMrc20TickHolder(id, txNum)
	}
}
func handleNativTransfer(chainName string, height int64, mrc20TransferPinTx map[string]struct{}, txInList []string, isMempool bool) {
	mrc20transferCheck, err := DbAdapter.GetMrc20UtxoByOutPutList(txInList, isMempool)
	if err == nil && len(mrc20transferCheck) > 0 {
		mrc20TrasferList := IndexerAdapter[chainName].CatchNativeMrc20Transfer(height, mrc20transferCheck, mrc20TransferPinTx)
		if len(mrc20TrasferList) > 0 {
			DbAdapter.UpdateMrc20Utxo(mrc20TrasferList, isMempool)
		}
	}
}
func transferHandle(transferHandleList []*pin.PinInscription) (mrc20UtxoList []*mrc20.Mrc20Utxo) {
	validator := Mrc20Validator{}
	maxTimes := len(transferHandleList)
	successMap := make(map[string]struct{})
	for i := 0; i < maxTimes; i++ {
		if len(successMap) >= maxTimes {
			break
		}
		for _, pinNode := range transferHandleList {
			if len(successMap) >= maxTimes {
				break
			}
			if _, ok := successMap[pinNode.Id]; ok {
				continue
			}
			transferPinList, _ := CreateMrc20TransferUtxo(pinNode, &validator, false)
			if len(transferPinList) > 0 {
				mrc20UtxoList = append(mrc20UtxoList, transferPinList...)
				successMap[pinNode.Id] = struct{}{}
				DbAdapter.UpdateMrc20Utxo(mrc20UtxoList, false)
			}
		}
	}
	return
}
func deployHandle(pinNode *pin.PinInscription) (mrc20UtxoList []mrc20.Mrc20Utxo) {
	var deployList []mrc20.Mrc20DeployInfo
	validator := Mrc20Validator{}
	//for _, pinNode := range deployHandleList {
	mrc20Pin, preMineUtxo, info, err := CreateMrc20DeployPin(pinNode, &validator)
	if err == nil {
		if mrc20Pin.Mrc20Id != "" {
			mrc20Pin.Chain = pinNode.ChainName
			mrc20UtxoList = append(mrc20UtxoList, mrc20Pin)
		}
		if preMineUtxo.Mrc20Id != "" {
			mrc20UtxoList = append(mrc20UtxoList, preMineUtxo)
		}
		if info.Tick != "" && info.Mrc20Id != "" {
			deployList = append(deployList, info)
		}
	}
	//}
	if len(deployList) > 0 {
		DbAdapter.SaveMrc20Tick(deployList)
	}
	return
}
func CreateMrc20DeployPin(pinNode *pin.PinInscription, validator *Mrc20Validator) (mrc20Utxo mrc20.Mrc20Utxo, preMineUtxo mrc20.Mrc20Utxo, info mrc20.Mrc20DeployInfo, err error) {
	var df mrc20.Mrc20Deploy
	err = json.Unmarshal(pinNode.ContentBody, &df)
	if err != nil {
		return
	}
	premineCount := int64(0)
	if df.PremineCount != "" {
		premineCount, err = strconv.ParseInt(df.PremineCount, 10, 64)
		if err != nil {
			return
		}
	}
	mintCount, err := strconv.ParseInt(df.MintCount, 10, 64)
	if err != nil {
		return
	}
	if mintCount < 0 {
		mintCount = int64(0)
	}
	amtPerMint, err := strconv.ParseInt(df.AmtPerMint, 10, 64)
	if err != nil {
		return
	}
	if amtPerMint < 0 {
		amtPerMint = int64(0)
	}
	//premineCount
	if mintCount < premineCount {
		return
	}
	premineAddress, pointValue, err1 := validator.Deploy(pinNode.ContentBody, pinNode)
	if err1 != nil {
		//mrc20Utxo.Verify = false
		//mrc20Utxo.Msg = err1.Error()
		return
	}
	info.Tick = strings.ToUpper(df.Tick)
	info.TokenName = df.TokenName
	info.Decimals = df.Decimals
	info.AmtPerMint = df.AmtPerMint
	info.PremineCount = uint64(premineCount)
	info.MintCount = uint64(mintCount)
	info.BeginHeight = df.BeginHeight
	info.EndHeight = df.EndHeight
	info.Metadata = df.Metadata
	info.DeployType = df.DeployType
	info.PinCheck = df.PinCheck
	info.PayCheck = df.PayCheck
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
	mrc20Utxo.FromAddress = pinNode.CreateAddress
	mrc20Utxo.ToAddress = pinNode.Address
	mrc20Utxo.TxPoint = pinNode.Output
	mrc20Utxo.PinContent = string(pinNode.ContentBody)
	mrc20Utxo.Timestamp = pinNode.Timestamp
	mrc20Utxo.PointValue = uint64(pinNode.OutputValue)
	mrc20Utxo.Verify = true

	if premineAddress != "" && premineCount > 0 {
		preMineUtxo.Verify = true
		//preMineUtxo.PinId = pinNode.Id
		preMineUtxo.BlockHeight = pinNode.GenesisHeight
		preMineUtxo.MrcOption = "pre-mint"
		preMineUtxo.FromAddress = pinNode.Address
		preMineUtxo.ToAddress = premineAddress
		preMineUtxo.TxPoint = fmt.Sprintf("%s:%d", pinNode.GenesisTransaction, 1)
		//mrc20Utxo.PinContent = string(pinNode.ContentBody)
		preMineUtxo.Timestamp = pinNode.Timestamp
		preMineUtxo.PointValue = uint64(pointValue)
		preMineUtxo.Mrc20Id = info.Mrc20Id
		preMineUtxo.Tick = info.Tick
		preMineUtxo.Chain = pinNode.ChainName
		//preMineUtxo.AmtChange = premineCount * amtPerMint
		num := strconv.FormatInt(premineCount*amtPerMint, 10)
		preMineUtxo.AmtChange, _ = decimal.NewFromString(num)
		info.TotalMinted = uint64(premineCount)
	}
	return
}

func CreateMrc20MintPin(pinNode *pin.PinInscription, validator *Mrc20Validator, mempool bool) (mrc20Utxo mrc20.Mrc20Utxo, err error) {
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
	mrc20Utxo.PointValue = uint64(pinNode.OutputValue)
	info, shovelList, toAddress, vout, err1 := validator.Mint(content, pinNode)
	if toAddress != "" {
		mrc20Utxo.ToAddress = toAddress
		mrc20Utxo.TxPoint = fmt.Sprintf("%s:%d", pinNode.GenesisTransaction, vout)
	}
	if info != (mrc20.Mrc20DeployInfo{}) {
		mrc20Utxo.Mrc20Id = info.Mrc20Id
		mrc20Utxo.Tick = info.Tick
	}
	if mempool {
		mrc20Utxo.Mrc20Id = info.Mrc20Id
		mrc20Utxo.AmtChange, _ = decimal.NewFromString(info.AmtPerMint)
		return
	}
	if err1 != nil {
		mrc20Utxo.Mrc20Id = info.Mrc20Id
		mrc20Utxo.Verify = false
		mrc20Utxo.Msg = err1.Error()
	} else {
		if len(shovelList) > 0 {
			DbAdapter.AddMrc20Shovel(shovelList, pinNode.Id, mrc20Utxo.Mrc20Id)
		}
		DbAdapter.UpdateMrc20TickInfo(info.Mrc20Id, mrc20Utxo.TxPoint, uint64(info.TotalMinted)+1)
		//mrc20Utxo.AmtChange, _ = strconv.ParseInt(info.AmtPerMint, 10, 64)
		mrc20Utxo.AmtChange, _ = decimal.NewFromString(info.AmtPerMint)
	}

	return
}

func CreateMrc20TransferUtxo(pinNode *pin.PinInscription, validator *Mrc20Validator, isMempool bool) (mrc20UtxoList []*mrc20.Mrc20Utxo, err error) {
	//Check if it has been processed
	find, err1 := DbAdapter.CheckOperationtx(pinNode.GenesisTransaction, isMempool)
	if err1 != nil || find != nil {
		return
	}

	var content []mrc20.Mrc20TranferData
	err = json.Unmarshal(pinNode.ContentBody, &content)
	if err != nil {
		mrc20UtxoList = sendAllAmountToFirstOutput(pinNode, "Transfer JSON format error", isMempool)
		return
	}
	//check
	toAddress, utxoList, outputValueList, msg, firstIdx, err1 := validator.Transfer(content, pinNode, isMempool)
	//if err1 != nil && err1.Error() != "valueErr" {
	if err1 != nil {
		mrc20UtxoList = sendAllAmountToFirstOutput(pinNode, msg, isMempool)
		return
	}
	address := make(map[string]string)
	name := make(map[string]string)
	inputAmtMap := make(map[string]decimal.Decimal)
	var spendUtxoList []*mrc20.Mrc20Utxo
	for _, utxo := range utxoList {
		address[utxo.Mrc20Id] = utxo.ToAddress
		name[utxo.Mrc20Id] = utxo.Tick
		//Spent the input UTXO
		//amt := utxo.AmtChange * -1
		//amt := utxo.AmtChange.Mul(decimal.NewFromInt(-1))
		//amt := utxo.AmtChange
		//mrc20Utxo := mrc20.Mrc20Utxo{TxPoint: utxo.TxPoint, Index: utxo.Index, Mrc20Id: utxo.Mrc20Id, Verify: true, Status: -1, AmtChange: amt}
		//if isMempool {
		mrc20Utxo := *utxo
		mrc20Utxo.Status = -1
		//}
		mrc20Utxo.OperationTx = pinNode.GenesisTransaction
		spendUtxoList = append(spendUtxoList, &mrc20Utxo)
		//inputAmtMap[utxo.Mrc20Id] += utxo.AmtChange
		inputAmtMap[utxo.Mrc20Id] = inputAmtMap[utxo.Mrc20Id].Add(utxo.AmtChange)
	}
	outputAmtMap := make(map[string]decimal.Decimal)
	x := 0
	var reciveUtxoList []*mrc20.Mrc20Utxo
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
		mrc20Utxo.OperationTx = pinNode.GenesisTransaction
		mrc20Utxo.PointValue = uint64(outputValueList[item.Vout])
		//mrc20Utxo.AmtChange, _ = strconv.ParseInt(item.Amount, 10, 64)
		mrc20Utxo.AmtChange, _ = decimal.NewFromString(item.Amount)
		//outputAmtMap[item.Id] += mrc20Utxo.AmtChange
		outputAmtMap[item.Id] = outputAmtMap[item.Id].Add(mrc20Utxo.AmtChange)
		mrc20Utxo.Timestamp = pinNode.Timestamp
		reciveUtxoList = append(reciveUtxoList, &mrc20Utxo)
		x += 1
	}
	//Check if the input exceeds the output.
	for id, inputAmt := range inputAmtMap {
		//inputAmt > outputAmtMap[id]
		if inputAmt.Compare(outputAmtMap[id]) == 1 {
			//if !isMempool {
			// find := false
			// for _, utxo := range mrc20UtxoList {
			// 	vout := strings.Split(utxo.TxPoint, ":")[1]
			// 	if utxo.Mrc20Id == id && utxo.ToAddress == toAddress[0] && vout == "0" {
			// 		//utxo.AmtChange += (inputAmt - outputAmtMap[id])

			// 		diff := inputAmt.Sub(outputAmtMap[id])
			// 		fmt.Println("2===>", diff, utxo.AmtChange)
			// 		utxo.AmtChange = utxo.AmtChange.Add(diff)

			// 		utxo.Msg = "The total input amount is greater than the output amount"
			// 		find = true
			// 	}
			// }
			// if find {
			// 	continue
			// }
			//}
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
			mrc20Utxo.PointValue = uint64(outputValueList[firstIdx])
			mrc20Utxo.PinContent = string(pinNode.ContentBody)
			mrc20Utxo.OperationTx = pinNode.GenesisTransaction
			mrc20Utxo.Index = x
			//mrc20Utxo.AmtChange = inputAmt - outputAmtMap[id]
			mrc20Utxo.AmtChange = inputAmt.Sub(outputAmtMap[id])
			mrc20Utxo.Msg = "The total input amount is greater than the output amount"
			mrc20UtxoList = append(mrc20UtxoList, &mrc20Utxo)
			x += 1
		}
	}
	mrc20UtxoList = append(mrc20UtxoList, spendUtxoList...)
	mrc20UtxoList = append(mrc20UtxoList, reciveUtxoList...)
	return
}
func sendAllAmountToFirstOutput(pinNode *pin.PinInscription, msg string, isMempool bool) (mrc20UtxoList []*mrc20.Mrc20Utxo) {
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
		class, addresses, _, _ := txscript.ExtractPkScriptAddrs(out.PkScript, ChainParams[pinNode.ChainName])
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
	list, err := DbAdapter.GetMrc20UtxoByOutPutList(inputList, isMempool)
	if err != nil {
		//log.Println("GetMrc20UtxoByOutPutList:", err)
		return
	}
	utxoList := make(map[string]*mrc20.Mrc20Utxo)
	for _, item := range list {
		//Spent the input UTXO
		//amt := item.AmtChange * -1
		amt := item.AmtChange.Neg()
		mrc20Utxo := mrc20.Mrc20Utxo{TxPoint: item.TxPoint, Index: item.Index, Mrc20Id: item.Mrc20Id, Verify: true, Status: -1, AmtChange: amt}
		mrc20UtxoList = append(mrc20UtxoList, &mrc20Utxo)
		if v, ok := utxoList[item.Mrc20Id]; ok {
			//v.AmtChange += item.AmtChange
			v.AmtChange = v.AmtChange.Add(item.AmtChange)
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
				PointValue:  uint64(value),
				PinContent:  string(pinNode.ContentBody),
				Index:       0,
				AmtChange:   item.AmtChange,
				Msg:         msg,
				OperationTx: pinNode.GenesisTransaction,
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
