package common

import (
	"encoding/json"
	"manindexer/basicprotocols/metaaccess"
	"os"
)

type EcdhKey struct {
	Pubkey string `json:"pubkey"`
	Prikey string `json:"prikey"`
}

func InitMetasoKey() (config EcdhKey) {
	fileName := "/metaso/key.json"
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		prikey, pubkey, err := metaaccess.GenKeyPair()
		if err != nil {
			return
		}
		defaultConfig := EcdhKey{
			Pubkey: pubkey,
			Prikey: prikey,
		}
		file, err := os.Create(fileName)
		if err != nil {
			return
		}
		defer file.Close()
		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(defaultConfig); err != nil {
			return
		}
		config = defaultConfig
	} else {
		content, err := os.ReadFile(fileName)
		if err != nil {
			return
		}
		if err := json.Unmarshal(content, &config); err != nil {
			return
		}
	}
	return
}
