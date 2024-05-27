package common

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
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
