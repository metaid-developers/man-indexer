package common

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
)

func CalculateHash(pinid string, merkleRoot string) string {
	h := sha256.New()
	h.Write([]byte(pinid + merkleRoot))
	return hex.EncodeToString(h.Sum(nil))
}

func CalculateProductToHexStr(blockhash string, pinHash string) string {
	blockhashByte, _ := hex.DecodeString(blockhash)
	blockhashInt, _ := new(big.Int).SetString(blockhash, 16)
	pinHashByte, _ := hex.DecodeString(pinHash)
	pinHashInt, _ := new(big.Int).SetString(pinHash, 16)
	popByte := new(big.Int).Mul(blockhashInt, pinHashInt).Bytes()
	//calculate the total number of digits:32+32=64
	totalLen := len(blockhashByte) + len(pinHashByte)
	//number of leading zeros required
	remainingLen := totalLen - len(popByte)
	for i := 0; i < remainingLen; i++ {
		popByte = append([]byte{0}, popByte...)
	}
	return hex.EncodeToString(popByte)
}

func ConvertToOctalHex(productHex string) (string, int64) {
	productByte, _ := hex.DecodeString(productHex)
	//convert to binary
	bList := make([]string, 0)
	for _, b := range productByte {
		binaryB := fmt.Sprintf("%b", b)
		bList = append(bList, fmt.Sprintf("%08s", binaryB))
	}
	productBinaryStr := ""
	for _, b := range bList {
		productBinaryStr += b
	}
	productBinaryStr = productBinaryStr[:510]

	bCount := int64(0)
	for _, b := range productBinaryStr {
		if b == '0' {
			bCount++
		} else {
			break
		}
	}
	//convert binary string to octal string
	octal := ""
	for i := 0; i < len(productBinaryStr); i += 3 {
		binaryStr := productBinaryStr[i : i+3]
		num, err := strconv.ParseInt(binaryStr, 2, 64)
		if err != nil {
			fmt.Println("ParseInt error:", err)
			return "", 0
		}
		octal += strconv.FormatInt(num, 8)
	}
	return octal, bCount
}

func GenPop(pinid, merkleRoot, blockHash string) (string, int64) {
	//calculate pinHash
	pinHash := CalculateHash(pinid, merkleRoot)
	//blockhash * pinHash
	productHexStr := CalculateProductToHexStr(blockHash, pinHash)
	//convert to octal
	octal, bCount := ConvertToOctalHex(productHexStr)
	return octal, bCount
}
