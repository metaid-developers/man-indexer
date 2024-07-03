package common

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/BurntSushi/toml"
)

var (
	Config      *AllConfig
	configMutex sync.Mutex
	Chain       string
	Db          string
	Server      string
	TestNet     string
)

type AllConfig struct {
	ProtocolID string `toml:"protocolID"`
	Sync       syncConfig
	Protocols  map[string]protocols
	Btc        btcConfig
	Mvc        mvcConfig
	MongoDb    mongoConfig
	Pebble     pebble
	Web        webConfig
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
type mvcConfig struct {
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
	fmt.Println("1>>")
	configMutex.Lock()
	defer configMutex.Unlock()
	filePath := "./config.toml"
	if _, err := toml.DecodeFile(filePath, &Config); err != nil {
		panic(err)
	}
	flagConfig := GetFlagConfig()
	for k, v := range flagConfig {
		if *v == "" {
			continue
		}
		switch k {
		case "btc_height":
			Config.Btc.InitialHeight, _ = strconv.ParseInt(*v, 10, 64)
		case "btc_rpc_host":
			Config.Btc.RpcHost = *v
		case "btc_rpc_user":
			Config.Btc.RpcUser = *v
		case "btc_rpc_password":
			Config.Btc.RpcPass = *v
		case "btc_zmqpubrawtx":
			Config.Btc.ZmqHost = *v
		case "mvc_height":
			Config.Mvc.InitialHeight, _ = strconv.ParseInt(*v, 10, 64)
		case "mvc_rpc_host":
			Config.Mvc.RpcHost = *v
		case "mvc_rpc_user":
			Config.Mvc.RpcUser = *v
		case "mvc_rpc_password":
			Config.Mvc.RpcPass = *v
		case "mvc_zmqpubrawtx":
			Config.Mvc.ZmqHost = *v
		case "server_port":
			Config.Web.Port = *v
		case "https_pem_file":
			Config.Web.PemFile = *v
		case "https_key_file":
			Config.Web.KeyFile = *v
		case "domain_name":
			Config.Web.Host = *v
		case "mongo_uri":
			Config.MongoDb.MongoURI = *v
		case "mongo_db_name":
			Config.MongoDb.DbName = *v
		}
	}
	if TestNet == "1" {
		Config.Btc.PopCutNum = 18
		Config.Mvc.PopCutNum = 8
		Config.ProtocolID = "746573746964"
	} else if TestNet == "2" {
		Config.Btc.PopCutNum = 0
		Config.Mvc.PopCutNum = 0
		Config.ProtocolID = "746573746964"
	} else if TestNet == "0" {
		Config.Btc.PopCutNum = 22
		Config.Mvc.PopCutNum = 12
		Config.ProtocolID = "6d6574616964"
	}
}
func GetFlagConfig() (flagConfig map[string]*string) {
	chain := flag.String("chain", "btc", "Which chain to perform indexing")
	db := flag.String("database", "mongo", "Which database to use")
	testNet := flag.String("test", "", "Connect to testnet")
	server := flag.String("server", "1", "Run the explorer service")
	flagConfig = make(map[string]*string)
	flagConfig["btc_height"] = flag.String("btc_height", "", "btc starting block height")
	flagConfig["btc_rpc_host"] = flag.String("btc_rpc_host", "", "btc rpc host")
	flagConfig["btc_rpc_user"] = flag.String("btc_rpc_user", "", "btc rpcuser")
	flagConfig["btc_rpc_password"] = flag.String("btc_rpc_password", "", "btc rpc password")
	flagConfig["btc_zmqpubrawtx"] = flag.String("btc_zmqpubrawtx", "", "btc zmqpubrawtx")
	flagConfig["mvc_height"] = flag.String("mvc_height", "", "mvc starting block height")
	flagConfig["mvc_rpc_host"] = flag.String("mvc_rpc_host", "", "mvc rpc host")
	flagConfig["mvc_rpc_user"] = flag.String("mvc_rpc_user", "", "mvc rpcuser")
	flagConfig["mvc_rpc_password"] = flag.String("mvc_rpc_password", "", "mvc rpc password")
	flagConfig["mvc_zmqpubrawtx"] = flag.String("mvc_zmqpubrawtx", "", "mvc zmqpubrawtx")
	flagConfig["server_port"] = flag.String("server_port", "", "server port")
	flagConfig["https_pem_file"] = flag.String("https_pem_file", "", "http pem file")
	flagConfig["https_key_file"] = flag.String("https_key_file", "", "https key file")
	flagConfig["domain_name"] = flag.String("domain_name", "", "domain name")
	flagConfig["mongo_uri"] = flag.String("mongo_uri", "", "mongodb uri")
	flagConfig["mongo_db_name"] = flag.String("mongo_db_name", "", "mongodb database name")
	//reindex := flag.String("reindex", "", "reindex block height,from:to")
	flag.Parse()
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "args:\n")
		flag.PrintDefaults()
	}
	Chain = *chain
	Db = *db
	TestNet = *testNet
	Server = *server
	return
}
