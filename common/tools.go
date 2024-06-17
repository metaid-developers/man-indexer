package common

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"manindexer/common/btc_util"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/txscript"
)

func DetectContentType(content *[]byte) (contentType string) {
	c := *content
	var buffer []byte
	if len(c) > 512 {
		buffer = c[0:512]
	} else {
		buffer = c
	}
	contentType = http.DetectContentType(buffer)
	return
}
func GetMetaIdByAddress(address string) (metaId string) {
	hash := sha256.New()
	hash.Write([]byte(address))
	hashBytes := hash.Sum(nil)
	metaId = hex.EncodeToString(hashBytes)
	return
}
func InitHeightFile(filePath string, height int64) {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		file, err := os.Create(filePath)
		if err != nil {
			return
		}
		defer file.Close()
		_, err = file.WriteString(strconv.FormatInt(height, 10))
		if err != nil {
			return
		}
	}
}
func GetLocalLastHeight(filePath string) (last int64, err error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Println(err)
		return
	}
	s := strings.Replace(string(content), "\n", "", -1)
	last, err = strconv.ParseInt(s, 10, 64)
	if err != nil {
		//fmt.Println("s:", s)
		last = 0
		err = nil
	}
	return
}

func UpdateLocalLastHeight(filePath string, newHeight int64) (err error) {
	if filePath == "" {
		return
	}
	return os.WriteFile(filePath, []byte(strconv.FormatInt(newHeight, 10)), 0644)
}

// isBase64 checks if a given string is a valid base64 encoded string
func isBase64(s string) bool {
	if len(s)%4 != 0 {
		return false
	}
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

// isBase64Image checks if a given string is a valid base64 encoded image
func IsBase64Image(base64Str string) (baseStr string, isImage bool) {
	if len(base64Str) < 23 {
		return
	}
	if base64Str[0:22] == "data:image/gif;base64," || base64Str[0:22] == "data:image/png;base64," || base64Str[0:23] == "data:image/jpeg;base64," {
		return
	}

	// Check if the string is a valid base64 encoded string
	if !isBase64(base64Str) {
		return
	}

	// Decode the base64 string
	data, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return
	}
	// Check if the decoded data is a valid image
	_, imgType, err := image.DecodeConfig(strings.NewReader(string(data)))
	if err != nil {
		return
	}
	isImage = true
	baseStr = fmt.Sprintf("data:image/%s;base64,", imgType)
	return
}
func BtcParseWitnessScript(witness [][]byte) (data [][]string, err error) {
	if len(witness) == 0 {
		err = errors.New("no witness data found")
		return
	}
	var witnessScript []byte
	if witness[len(witness)-1][0] == txscript.TaprootAnnexTag {
		witnessScript = witness[len(witness)-1]
	} else {
		if len(witness) < 2 {
			return
		}
		witnessScript = witness[len(witness)-2]
	}

	tokenizer := txscript.MakeScriptTokenizer(0, witnessScript)
	i := 0
	for tokenizer.Next() {
		codeName := "Unkonw"
		if v, ok := btc_util.OpcodeMap[tokenizer.Opcode()]; ok {
			codeName = v
		}
		codeName = fmt.Sprintf("%s(%d)", codeName, tokenizer.Opcode())
		if i == 0 {
			data = append(data, []string{codeName, hex.EncodeToString(tokenizer.Data())})
		} else {
			data = append(data, []string{codeName, string(tokenizer.Data())})
		}
		i += 1
	}
	return
}
