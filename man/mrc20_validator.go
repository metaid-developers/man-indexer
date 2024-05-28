package man

import (
	"encoding/json"
	"errors"
	"log"
	"manindexer/mrc20"
	"manindexer/pin"
	"strconv"
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
	if len(data.Tick) < 4 {
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

func (validator *Mrc20Validator) Mint(content mrc20.Mrc20MintData, pinNode *pin.PinInscription) (info mrc20.Mrc20DeployInfo, shovel string, err error) {
	if content.Id == "" {
		err = errors.New(mrc20.ErrMintTickIdNull)
		return
	}
	if content.Pin == "" {
		err = errors.New(mrc20.ErrMintPinIdNull)
		return
	}
	info, err = DbAdapter.GetMrc20TickInfo(content.Id)
	if err != nil {
		log.Println("GetMrc20TickInfo:", err)
		return
	}
	if info == (mrc20.Mrc20DeployInfo{}) {
		err = errors.New(mrc20.ErrMintTickNotExists)
		return
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
	if info.Qual.Lv == "0" {
		return
	}
	// tx, err := ChainAdapter.GetTransaction(pinNode.GenesisTransaction)
	// if err != nil {
	// 	log.Println("GetTransaction:", err)
	// 	return
	// }
	// txb := tx.(*btcutil.Tx)
	// var inputList []string
	// for _, in := range txb.MsgTx().TxIn {
	// 	s := fmt.Sprintf("%si%d", in.PreviousOutPoint.Hash.String(), in.PreviousOutPoint.Index)
	// 	inputList = append(inputList, s)
	// }
	// fmt.Println(inputList)
	// pins, err := DbAdapter.GetPinListByIdList(inputList)
	inputList := []string{content.Pin}
	pins, err := DbAdapter.GetPinListByIdList(inputList)
	if err != nil {
		log.Println("GetPinListByIdList:", err)
		return
	}
	if len(pins) <= 0 {
		err = errors.New(mrc20.ErrMintPopNull)
		return
	}
	if pinNode.Address != pins[0].Address {
		err = errors.New(mrc20.ErrMintPinOwner)
		return
	}
	var pinIds []string
	for _, pinNode := range pins {
		pinIds = append(pinIds, pinNode.Id)
	}
	usedShovels, err := DbAdapter.GetMrc20Shovel(pinIds)
	popChcek := false
	popLimit, _ := strconv.Atoi(info.Qual.Lv)
	for _, pinNode := range pins {
		if _, ok := usedShovels[pinNode.Id]; ok {
			continue
		}
		find := countLeadingZeros(pinNode.Pop)
		if find >= popLimit {
			popChcek = true
			shovel = pinNode.Id
			break
		}
	}
	if !popChcek {
		err = errors.New(mrc20.ErrMintPopDiff)
		return
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
func (validator *Mrc20Validator) Transfer(content []mrc20.Mrc20TranferData, pinNode *pin.PinInscription) (tickNameMap map[string]string, err error) {
	if len(content) <= 0 {
		err = errors.New(mrc20.ErrTranferReqData)
		return
	}
	sendMap := make(map[string]int64)
	for _, item := range content {
		if item.Addr == "" || item.Id == "" || item.Amount == "" {
			err = errors.New(mrc20.ErrTranferReqData)
			return
		}
		amt, err1 := strconv.ParseInt(item.Amount, 10, 64)
		if err1 != nil {
			err = errors.New(mrc20.ErrTranferReqData)
			return
		}
		if v, ok := sendMap[item.Id]; ok {
			sendMap[item.Id] = v + amt
		} else {
			sendMap[item.Id] = amt
		}
	}
	tickNameMap = make(map[string]string)
	for k, v := range sendMap {
		blance, tick, err1 := getBalanceByaddressAndTick(pinNode.Address, k)
		tickNameMap[k] = tick
		if err1 != nil {
			err = errors.New(mrc20.ErrTranferBalnceErr)
			return
		}
		if v > blance {
			err = errors.New(mrc20.ErrTranferBalnceLess)
			return
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
