package common

import (
	"sync"

	"github.com/BurntSushi/toml"
)

var (
	Config      *AllConfig
	configMutex sync.Mutex
)

type AllConfig struct {
	Sync      syncConfig
	Protocols map[string]protocols
	Btc       btcConfig
	MongoDb   mongoConfig
	Pebble    pebble
	Web       webConfig
}
type syncConfig struct {
	SyncAllData   bool     `toml:"syncAllData"`
	SyncProtocols []string `toml:"syncProtocols"`
	SyncBeginTime string   `toml:"syncBeginTime"`
	SyncEndTime   string   `toml:"syncEndTime"`
}
type protocols struct {
	Key     string          `toml:"key"`
	Fields  []protocolFeld  `toml:"fields"`
	Indexes []protocolIndex `toml:"indexes"`
}
type protocolFeld struct {
	Name   string `toml:"name"`
	Class  string `toml:"class"`
	Length int    `toml:"length"`
}
type protocolIndex struct {
	Fields []string `toml:"fields"`
	Unique bool     `toml:"unique"`
}
type btcConfig struct {
	InitialHeight   int64  `toml:"initialHeight"`
	RpcHost         string `toml:"rpcHost"`
	RpcUser         string `toml:"rpcUser"`
	RpcPass         string `toml:"rpcPass"`
	RpcHTTPPostMode bool   `toml:"rpcHttpPostMode"`
	RpcDisableTLS   bool   `toml:"rpcDisableTLS"`
	ZmqHost         string `toml:"zmqHost"`
	PopCutNum       int    `toml:"popCutNum"`
}
type mongoConfig struct {
	MongoURI string `toml:"mongoURI"`
	PoolSize int64  `toml:"poolSize"`
	TimeOut  int64  `toml:"timeOut"`
	DbName   string `toml:"dbName"`
}
type webConfig struct {
	Port    string `toml:"port"`
	PemFile string `toml:"pemFile"`
	KeyFile string `toml:"keyFile"`
	Host    string `toml:"host"`
}
type pebble struct {
	Dir string `toml:"dir"`
}

func init() {
	configMutex.Lock()
	defer configMutex.Unlock()
	filePath := "./config.toml"
	if _, err := toml.DecodeFile(filePath, &Config); err != nil {
		panic(err)
	}
}
