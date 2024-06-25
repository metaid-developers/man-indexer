package pin

import (
	"manindexer/common"
	"math"
	"math/big"
	"strings"
)

func PopLevelCount(chainName, pop string) (lv int, lastStr string) {
	PopCutNum := 1000
	switch chainName {
	case "btc":
		PopCutNum = common.Config.Btc.PopCutNum
	case "mvc":
		PopCutNum = common.Config.Mvc.PopCutNum
	}
	if len(pop) < PopCutNum {
		lv = -1
		lastStr = pop
		return
	}
	cnt := len(pop) - len(strings.TrimLeft(pop, "0"))
	if cnt <= PopCutNum {
		lv = -1
		lastStr = pop[PopCutNum:]
		return
	} else {
		lv = cnt - PopCutNum
		lastStr = pop[PopCutNum:]
		return
	}
}
func RarityScoreBinary(chainName, binaryStr string) int {
	popCutNum := 0
	switch chainName {
	case "btc":
		popCutNum = common.Config.Btc.PopCutNum
	case "mvc":
		popCutNum = common.Config.Mvc.PopCutNum
	}
	if len(binaryStr) < popCutNum {
		return 0
	}
	binaryStr = binaryStr[popCutNum:]
	// Step 1: Count the number of leading zeros
	n := len(binaryStr) - len(strings.TrimLeft(binaryStr, "0"))

	// Step 2: Remove leading zeros and calculate the decimal value of the rest part
	restPart := strings.TrimLeft(binaryStr, "0")
	if restPart == "" {
		// In case the binary string is all zeros
		return int(math.Pow(2, float64(n)))
	}

	//fmt.Println("rest:", restPart)
	// restValue, err := strconv.ParseInt(restPart, 2, 64)
	// if err != nil {
	// 	fmt.Printf("Error parsing binary string: %v\n", err)
	// 	return 0
	// }
	// k := len(restPart)
	// // Step 3: Normalize the rest value and invert it
	// normalizedValue := (1 - (float64(restValue)+1)/math.Pow(2, float64(k))) * 2
	bigInt := new(big.Int)
	bigInt.SetString(restPart, 10)
	base := new(big.Int)
	max := int64(170 - popCutNum)
	base.Exp(big.NewInt(10), big.NewInt(max), nil)
	bigFloat := new(big.Float).SetInt(bigInt)
	baseFloat := new(big.Float).SetInt(base)
	normalizedFloat := new(big.Float).Quo(bigFloat, baseFloat)
	normalizedValue, _ := normalizedFloat.Float64()
	// Step 4: Calculate the final score
	score := math.Pow(2, float64(n)) + normalizedValue*math.Pow(2, float64(n))
	// Step 5: Round the score to the nearest integer
	return int(math.Round(score))
}
