package man

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"manindexer/mrc20"
	"manindexer/pin"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/txscript"
)

type Mrc20Validator struct {
}

func (validator *Mrc20Validator) Check(pinNode *pin.PinInscription) {

}
func (validator *Mrc20Validator) Deploy(content []byte) error {
	var data mrc20.Mrc20Deploy
	err := json.Unmarshal(content, &data)
	if err != nil {
		return errors.New(mrc20.ErrDeployContent)
	}
	if len(data.Tick) < 2 || len(data.Tick) > 24 {
		return errors.New(mrc20.ErrDeployTickLength)
	}
	/*
		info, err := DbAdapter.GetMrc20TickInfo(strings.ToLower(data.Tick))
		if err != nil {
			log.Println("GetMrc20TickInfo:", err)
		}
		if info != (mrc20.Mrc20DeployInfo{}) {
			return errors.New(mrc20.ErrDeployTickExists)
		}
	*/
	return nil
}

func (validator *Mrc20Validator) Mint(content mrc20.Mrc20MintData, pinNode *pin.PinInscription) (info mrc20.Mrc20DeployInfo, shovelList []string, err error) {
	if content.Id == "" {
		err = errors.New(mrc20.ErrMintTickIdNull)
		return
	}
	// if content.Pin == "" {
	// 	err = errors.New(mrc20.ErrMintPinIdNull)
	// 	return
	// }
	info, err = DbAdapter.GetMrc20TickInfo(content.Id)
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
	if info.Qual.Count == "" {
		info.Qual.Count = "1"
	}
	count, _ := strconv.ParseInt(info.MintCount, 10, 64)
	height, _ := strconv.ParseInt(info.Blockheight, 10, 64)
	if info.TotalMinted >= count {
		err = errors.New(mrc20.ErrMintLimit)
		return
	}
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
	pins, err := DbAdapter.GetPinListByOutPutList(inputList)
	//inputList := []string{content.Pin}
	//pins, err := DbAdapter.GetPinListByIdList(inputList)
	if err != nil {
		log.Println("GetPinListByOutPutList:", err, inputList)
		return
	}
	if len(pins) <= 0 {
		err = errors.New(mrc20.ErrMintPopNull)
		return
	}
	// if pinNode.Address != pins[0].Address {
	// 	err = errors.New(mrc20.ErrMintPinOwner)
	// 	return
	// }
	var pinIds []string
	for _, pinNode := range pins {
		pinIds = append(pinIds, pinNode.Id)
	}
	usedShovels, err := DbAdapter.GetMrc20Shovel(pinIds)
	shovelsCount, _ := strconv.Atoi(info.Qual.Count)
	shovelChcek := true
	var lvShovelList []string
	if info.Qual.Lv != "" {
		popLimit, _ := strconv.Atoi(info.Qual.Lv)
		shovelChcek, lvShovelList = lvCheck(usedShovels, pins, shovelsCount, popLimit)
		if !shovelChcek {
			err = errors.New(mrc20.ErrMintPopDiff)
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
	if len(lvShovelList) > 0 {
		shovelList = append(shovelList, lvShovelList...)
	}
	if len(pathShovelList) > 0 {
		shovelList = append(shovelList, pathShovelList...)
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
func (validator *Mrc20Validator) Transfer(content []mrc20.Mrc20TranferData, pinNode *pin.PinInscription) (toAddress map[int]string, utxoList []*mrc20.Mrc20Utxo, err error) {
	if len(content) <= 0 {
		err = errors.New(mrc20.ErrTranferReqData)
		return
	}
	outMap := make(map[string]int64)
	maxVout := 0
	for _, item := range content {
		if item.Id == "" || item.Amount == 0 {
			err = errors.New(mrc20.ErrTranferReqData)
			return
		}
		if maxVout < item.Vout {
			maxVout = item.Vout
		}
		outMap[item.Id] += item.Amount
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
		err = errors.New("valueErr")
		return
	}
	for _, item := range content {
		class, _, _, _ := txscript.ExtractPkScriptAddrs(txb.MsgTx().TxOut[item.Vout].PkScript, ChainParams)
		if class.String() == "nulldata" || class.String() == "nonstandard" {
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
		log.Println("GetMrc20UtxoByOutPutList:", err)
		return
	}
	inMap := make(map[string]int64)
	for _, item := range list {
		inMap[item.Mrc20Id] += item.AmtChange
		utxoList = append(utxoList, item)
	}
	//if out list value error
	for k, v := range outMap {
		if in, ok := inMap[k]; ok {
			if in != v {
				err = errors.New("valueErr")
				return
			}
		} else {
			err = errors.New("valueErr")
			return
		}
	}
	toAddress = make(map[int]string)
	for i, out := range txb.MsgTx().TxOut {
		class, addresses, _, _ := txscript.ExtractPkScriptAddrs(out.PkScript, ChainParams)
		if class.String() != "nulldata" && class.String() != "nonstandard" && len(addresses) > 0 {
			toAddress[i] = addresses[0].String()
		}
	}
	return
}
func getBalanceByaddressAndTick(address string, tickId string) (blance int64, tick string, err error) {
	list, err := DbAdapter.GetMrc20ByAddressAndTick(address, tickId)
	if err != nil {
		return
	}
	for _, item := range list {
		if item.AmtChange == 0 {
			continue
		}
		blance += item.AmtChange
		if tick == "" {
			tick = item.Tick
		}
	}
	return
}
