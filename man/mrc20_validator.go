package man

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"manindexer/common"
	"manindexer/mrc20"
	"manindexer/pin"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/txscript"
	"github.com/shopspring/decimal"
)

type Mrc20Validator struct {
}

func (validator *Mrc20Validator) Check(pinNode *pin.PinInscription) {

}
func (validator *Mrc20Validator) Deploy(content []byte, pinNode *pin.PinInscription) (string, int64, error) {
	lowerContent := strings.ToLower(string(content))
	var data mrc20.Mrc20DeployLow
	err := json.Unmarshal([]byte(lowerContent), &data)
	if err != nil {
		return "", 0, errors.New(mrc20.ErrDeployContent)
	}
	if len(data.Tick) < 2 || len(data.Tick) > 24 {
		return "", 0, errors.New(mrc20.ErrDeployTickLength)
	}
	if data.TokenName != "" {
		if len(data.TokenName) < 1 || len(data.TokenName) > 48 {
			return "", 0, errors.New(mrc20.ErrDeployTickNameLength)
		}
	}
	decimals := int64(0)
	if data.Decimals != "" {
		decimals, err := strconv.ParseInt(data.Decimals, 10, 64)
		if err != nil {
			return "", 0, err
		}

		if decimals < 0 || decimals > 12 {
			return "", 0, errors.New(mrc20.ErrDeployNum)
		}
	}

	amtPerMint, err := strconv.ParseInt(data.AmtPerMint, 10, 64)
	if err != nil {
		return "", 0, err
	}
	if amtPerMint < 1 || amtPerMint > 1000000000000 {
		return "", 0, errors.New(mrc20.ErrDeployNum)
	}

	mintCount, err := strconv.ParseInt(data.MintCount, 10, 64)
	if err != nil {
		return "", 0, err
	}
	if mintCount < 1 || mintCount > 1000000000000 {
		return "", 0, errors.New(mrc20.ErrDeployNum)
	}

	premineCount := int64(0)
	if data.PremineCount != "" {
		premineCount, err = strconv.ParseInt(data.PremineCount, 10, 64)
		if err != nil {
			return "", 0, err
		}
	}
	if premineCount > mintCount {
		return "", 0, errors.New(mrc20.ErrDeployNum)
	}
	//check tick name
	//ErrDeployTickExists
	tickName := strings.ToUpper(data.Tick)
	info, _ := DbAdapter.GetMrc20TickInfo("", tickName)
	if info != (mrc20.Mrc20DeployInfo{}) {
		if tickName == info.Tick {
			return "", 0, errors.New(mrc20.ErrDeployTickExists)
		}
	}

	if premineCount <= 0 {
		return "", 0, nil
	}
	t := getDigitsCount(amtPerMint*mintCount) + decimals
	if t > 20 {
		return "", 0, errors.New(mrc20.ErrDeployNum)
	}

	tx, err := ChainAdapter[pinNode.ChainName].GetTransaction(pinNode.GenesisTransaction)
	if err != nil {
		return "", 0, errors.New(mrc20.ErrDeployTxGet)
	}
	txb := tx.(*btcutil.Tx)
	//premineCount check
	if len(txb.MsgTx().TxOut) < 2 {
		return "", 0, errors.New("tx error")
	}
	if pinNode.Offset != 0 {
		return "", 0, errors.New("tx error")
	}
	toAddress := ""
	class, addresses, _, _ := txscript.ExtractPkScriptAddrs(txb.MsgTx().TxOut[1].PkScript, ChainParams)
	if class.String() != "nulldata" && class.String() != "nonstandard" && len(addresses) > 0 {
		toAddress = addresses[0].String()
	}
	return toAddress, txb.MsgTx().TxOut[1].Value, nil
}
func getDigitsCount(n int64) int64 {
	return int64(len(strconv.FormatInt(n, 10)))
}
func (validator *Mrc20Validator) Mint(content mrc20.Mrc20MintData, pinNode *pin.PinInscription) (info mrc20.Mrc20DeployInfo, shovelList []string, err error) {
	if content.Id == "" {
		err = errors.New(mrc20.ErrMintTickIdNull)
		return
	}
	//check if indexed

	// if content.Pin == "" {
	// 	err = errors.New(mrc20.ErrMintPinIdNull)
	// 	return
	// }
	info, err = DbAdapter.GetMrc20TickInfo(content.Id, "")
	if err != nil {
		log.Println("GetMrc20TickInfo:", err)
		return
	}
	if info == (mrc20.Mrc20DeployInfo{}) {
		err = errors.New(mrc20.ErrMintTickNotExists)
		return
	}
	if info.Chain != pinNode.ChainName {
		err = errors.New(mrc20.ErrCrossChain)
		return
	}
	//if info.TotalMinted >= (info.MintCount - info.PremineCount) {
	if (info.MintCount - info.TotalMinted) < 1 {
		err = errors.New(mrc20.ErrMintLimit)
		return
	}
	if info.Qual.Count == "" || info.Qual.Count == "0" {
		return
	}
	//count, _ := strconv.ParseInt(info.MintCount, 10, 64)
	height, _ := strconv.ParseInt(info.Blockheight, 10, 64)
	if pinNode.GenesisHeight < height {
		err = errors.New(mrc20.ErrMintHeight)
		return
	}
	// if info.Qual.Lv == "0" {
	// 	return
	// }
	tx, err := ChainAdapter[pinNode.ChainName].GetTransaction(pinNode.GenesisTransaction)
	if err != nil {
		log.Println("GetTransaction:", err)
		return
	}
	txb := tx.(*btcutil.Tx)
	var inputList []string
	//Because the PIN has been transferred,
	//use the output to find the PIN attributes.
	for i := range txb.MsgTx().TxOut {
		s := fmt.Sprintf("%s:%d", txb.Hash().String(), i)
		tmpId := fmt.Sprintf("%si%d", txb.Hash().String(), i)
		if tmpId != pinNode.Id {
			inputList = append(inputList, s)
		}
	}
	if len(inputList) <= 0 {
		err = errors.New(mrc20.ErrMintPopNull)
		return
	}
	// fmt.Println(inputList)
	pinsTmp, err := DbAdapter.GetPinListByOutPutList(inputList)
	//inputList := []string{content.Pin}
	//pins, err := DbAdapter.GetPinListByIdList(inputList)
	if err != nil {
		log.Println("GetPinListByOutPutList:", err, inputList)
		return
	}
	if len(pinsTmp) <= 0 {
		err = errors.New(mrc20.ErrMintPopNull)
		return
	}
	var pins []*pin.PinInscription
	for _, pinNode := range pinsTmp {
		if pinNode.Operation == "hide" {
			continue
		}
		pins = append(pins, pinNode)
	}
	// if pinNode.Address != pins[0].Address {
	// 	err = errors.New(mrc20.ErrMintPinOwner)
	// 	return
	// }
	var pinIds []string
	for _, pinNode := range pins {
		pinIds = append(pinIds, pinNode.Id)
	}
	if len(pinIds) <= 0 {
		err = errors.New(mrc20.ErrMintPopNull)
		return
	}
	usedShovels, err := DbAdapter.GetMrc20Shovel(pinIds, content.Id)

	shovelsCount, _ := strconv.Atoi(info.Qual.Count)
	shovelChcek := true
	var lvShovelList []string
	var creatorShovelList []string
	if info.Qual.Lv != "" {
		popLimit, _ := strconv.Atoi(info.Qual.Lv)
		shovelChcek, lvShovelList = lvCheck(usedShovels, pins, shovelsCount, popLimit)
		if !shovelChcek {
			err = errors.New(mrc20.ErrMintPopDiff)
			return
		}
	}
	//creator check
	if info.Qual.Creator != "" {
		shovelChcek, creatorShovelList = creatorCheck(usedShovels, pins, shovelsCount, info.Qual.Creator)
		if !shovelChcek {
			err = errors.New(mrc20.ErrMintCreator)
			return
		}
	}

	var pathShovelList []string
	if info.Qual.Path != "" {
		shovelChcek, pathShovelList = pathCheck(usedShovels, pins, shovelsCount, info.Qual.Path)
		if !shovelChcek {
			err = errors.New(mrc20.ErrMintPathCheck)
			return
		}
	}
	var countShovelList []string
	if info.Qual.Creator == "" && info.Qual.Lv == "" && info.Qual.Path == "" {
		shovelChcek, countShovelList = onlyCountCheck(usedShovels, pins, shovelsCount)
		if !shovelChcek {
			err = errors.New(mrc20.ErrMintCountCheck)
			return
		}
	}
	if len(lvShovelList) > 0 {
		shovelList = append(shovelList, lvShovelList...)
	}
	if len(creatorShovelList) > 0 {
		shovelList = append(shovelList, creatorShovelList...)
	}
	if len(pathShovelList) > 0 {
		shovelList = append(shovelList, pathShovelList...)
	}
	if len(countShovelList) > 0 {
		shovelList = append(shovelList, countShovelList...)
	}
	return
}
func lvCheck(usedShovels map[string]mrc20.Mrc20Shovel, pins []*pin.PinInscription, shovelsCount int, popLimit int) (verified bool, shovelList []string) {
	x := 0
	for _, pinNode := range pins {
		if _, ok := usedShovels[pinNode.Id]; ok {
			continue
		}
		find := countLeadingZeros(pinNode.Pop)
		if find >= popLimit {
			x += 1
			shovelList = append(shovelList, pinNode.Id)
		}
		if x == shovelsCount {
			break
		}
	}
	if x >= shovelsCount {
		verified = true
	}
	return
}
func creatorCheck(usedShovels map[string]mrc20.Mrc20Shovel, pins []*pin.PinInscription, shovelsCount int, creator string) (verified bool, shovelList []string) {
	x := 0
	for _, pinNode := range pins {
		if _, ok := usedShovels[pinNode.Id]; ok {
			continue
		}
		if common.GetMetaIdByAddress(pinNode.CreateAddress) == creator {
			x += 1
			shovelList = append(shovelList, pinNode.Id)
		}
		if x == shovelsCount {
			break
		}
	}
	if x >= shovelsCount {
		verified = true
	}
	return
}
func pathCheck(usedShovels map[string]mrc20.Mrc20Shovel, pins []*pin.PinInscription, shovelsCount int, pathStr string) (verified bool, shovelList []string) {
	path, query, key, operator, value := mrc20.PathParse(pathStr)
	if path == "" && query == "" {
		verified, shovelList = onlyPathCheck(usedShovels, pins, shovelsCount, pathStr)
		return
	}
	if path != "" && query != "" {
		if key == "" && operator == "" && value == "" {
			query = query[2 : len(query)-2]
			verified, shovelList = followPathCheck(usedShovels, pins, shovelsCount, path, query)
		} else if key != "" && operator != "" && value != "" {
			if operator == "=" {
				verified, shovelList = equalPathCheck(usedShovels, pins, shovelsCount, path, key, value)
			} else if operator == "#=" {
				verified, shovelList = contentPathCheck(usedShovels, pins, shovelsCount, path, key, value)
			}
		}
	}
	return
}
func onlyPathCheck(usedShovels map[string]mrc20.Mrc20Shovel, pins []*pin.PinInscription, shovelsCount int, pathStr string) (verified bool, shovelList []string) {
	pathArr := strings.Split(pathStr, "/")
	//Wildcard
	if pathArr[len(pathArr)-1] == "*" {
		pathStr = pathStr[0 : len(pathStr)-2]
	}
	x := 0
	for _, pinNode := range pins {
		if _, ok := usedShovels[pinNode.Id]; ok {
			continue
		}
		if len(pinNode.Path) < len(pathStr) {
			continue
		}
		//Wildcard
		if pinNode.Path[0:len(pathStr)] == pathStr {
			x += 1
			shovelList = append(shovelList, pinNode.Id)
		}
		if x == shovelsCount {
			break
		}
	}
	if x >= shovelsCount {
		verified = true
	}
	return
}
func followPathCheck(usedShovels map[string]mrc20.Mrc20Shovel, pins []*pin.PinInscription, shovelsCount int, pathStr string, queryStr string) (verified bool, shovelList []string) {
	x := 0
	if pathStr != "/follow" {
		return
	}
	for _, pinNode := range pins {
		if _, ok := usedShovels[pinNode.Id]; ok {
			continue
		}
		if string(pinNode.ContentBody) != queryStr {
			continue
		}
		x += 1
		shovelList = append(shovelList, pinNode.Id)
		if x == shovelsCount {
			break
		}
	}
	if x >= shovelsCount {
		verified = true
	}
	return
}

func equalPathCheck(usedShovels map[string]mrc20.Mrc20Shovel, pins []*pin.PinInscription, shovelsCount int, pathStr string, key string, value string) (verified bool, shovelList []string) {
	x := 0
	for _, pinNode := range pins {
		if _, ok := usedShovels[pinNode.Id]; ok {
			continue
		}
		if pinNode.Path != pathStr {
			continue
		}
		m := make(map[string]interface{})
		err := json.Unmarshal(pinNode.ContentBody, &m)
		if err != nil {
			continue
		}
		if _, ok := m[key]; !ok {
			continue
		}
		c := fmt.Sprintf("%s", m[key])
		if c != value {
			continue
		}
		x += 1
		shovelList = append(shovelList, pinNode.Id)
		if x == shovelsCount {
			break
		}
	}
	if x >= shovelsCount {
		verified = true
	}
	return
}
func contentPathCheck(usedShovels map[string]mrc20.Mrc20Shovel, pins []*pin.PinInscription, shovelsCount int, pathStr string, key string, value string) (verified bool, shovelList []string) {
	x := 0
	for _, pinNode := range pins {
		if _, ok := usedShovels[pinNode.Id]; ok {
			continue
		}
		if pinNode.Path != pathStr {
			continue
		}

		m := make(map[string]interface{})
		err := json.Unmarshal(pinNode.ContentBody, &m)
		if err != nil {
			continue
		}
		if _, ok := m[key]; !ok {
			continue
		}
		c := fmt.Sprintf("%s", m[key])
		if !strings.Contains(c, value) {
			continue
		}
		x += 1
		shovelList = append(shovelList, pinNode.Id)
		if x == shovelsCount {
			break
		}
	}
	if x >= shovelsCount {
		verified = true
	}
	return
}
func onlyCountCheck(usedShovels map[string]mrc20.Mrc20Shovel, pins []*pin.PinInscription, shovelsCount int) (verified bool, shovelList []string) {
	x := 0
	for _, pinNode := range pins {
		if _, ok := usedShovels[pinNode.Id]; ok {
			continue
		}
		x += 1
		shovelList = append(shovelList, pinNode.Id)
		if x == shovelsCount {
			break
		}
	}
	if x >= shovelsCount {
		verified = true
	}
	return
}
func countLeadingZeros(str string) int {
	count := 0
	for _, char := range str {
		if char == '0' {
			count++
		} else {
			break
		}
	}
	return count
}
func (validator *Mrc20Validator) Transfer(content []mrc20.Mrc20TranferData, pinNode *pin.PinInscription) (toAddress map[int]string, utxoList []*mrc20.Mrc20Utxo, outputValueList []int64, msg string, firstIdx int, err error) {
	if len(content) <= 0 {
		err = errors.New(mrc20.ErrTranferReqData)
		msg = mrc20.ErrTranferReqData
		return
	}
	outMap := make(map[string]decimal.Decimal)
	maxVout := 0
	for _, item := range content {
		if item.Id == "" || item.Amount == "" {
			err = errors.New(mrc20.ErrTranferReqData)
			msg = mrc20.ErrTranferReqData
			return
		}
		if maxVout < item.Vout {
			maxVout = item.Vout
		}
		//amt, _ := strconv.ParseInt(item.Amount, 10, 64)
		amt, _ := decimal.NewFromString(item.Amount)
		outMap[item.Id] = outMap[item.Id].Add(amt)

		tick, err1 := DbAdapter.GetMrc20TickInfo(item.Id, "")
		if err1 != nil {
			err = errors.New(mrc20.ErrMintTickIdNull)
			msg = mrc20.ErrMintTickIdNull
			return
		}
		decimals, _ := strconv.ParseInt(tick.Decimals, 10, 64)
		if getDecimalPlaces(item.Amount) > decimals {
			err = errors.New(mrc20.ErrMintDecimals)
			msg = mrc20.ErrMintDecimals
			return
		}

	}

	//get  mrc20 list in tx input
	tx, err := ChainAdapter[pinNode.ChainName].GetTransaction(pinNode.GenesisTransaction)
	if err != nil {
		log.Println("GetTransaction:", err)
		return
	}
	txb := tx.(*btcutil.Tx)
	//check output
	if maxVout > len(txb.MsgTx().TxOut) {
		msg = "Incorrect number of outputs in the transfer transaction"
		err = errors.New("valueErr")
		return
	}
	for _, item := range content {
		class, _, _, _ := txscript.ExtractPkScriptAddrs(txb.MsgTx().TxOut[item.Vout].PkScript, ChainParams)
		if class.String() == "nulldata" || class.String() == "nonstandard" {
			msg = "Incorrect vout target for the transfer"
			err = errors.New("valueErr")
			return
		}
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
	inMap := make(map[string]decimal.Decimal)
	for _, item := range list {
		inMap[item.Mrc20Id] = inMap[item.Mrc20Id].Add(item.AmtChange)
		utxoList = append(utxoList, item)
	}
	//if out list value error
	for k, v := range outMap {
		if in, ok := inMap[k]; ok {
			//in < v
			if in.Compare(v) == -1 {
				msg = "The total input amount is less than the output"
				err = errors.New("valueErr")
				return
			}
		} else {
			msg = "No available tick in the input"
			err = errors.New("valueErr")
			return
		}
	}
	toAddress = make(map[int]string)
	firstIdx = -1
	for i, out := range txb.MsgTx().TxOut {
		class, addresses, _, _ := txscript.ExtractPkScriptAddrs(out.PkScript, ChainParams)
		if class.String() != "nulldata" && class.String() != "nonstandard" && len(addresses) > 0 {
			toAddress[i] = addresses[0].String()
			if firstIdx < 0 {
				firstIdx = i
			}
		} else {
			toAddress[i] = "nonexistent"
		}
		outputValueList = append(outputValueList, out.Value)
	}
	return
}
func getDecimalPlaces(str string) int64 {
	if dotIndex := strings.IndexByte(str, '.'); dotIndex != -1 {
		return int64(len(str) - dotIndex - 1)
	}
	return int64(0)
}
func getBalanceByaddressAndTick(address string, tickId string) (blance decimal.Decimal, tick string, err error) {
	list, err := DbAdapter.GetMrc20ByAddressAndTick(address, tickId)
	if err != nil {
		return
	}
	for _, item := range list {
		//item.AmtChange == 0
		if item.AmtChange.Compare(decimal.Zero) == 0 {
			continue
		}
		//blance += item.AmtChange
		blance = blance.Add(item.AmtChange)
		if tick == "" {
			tick = item.Tick
		}
	}
	return
}
