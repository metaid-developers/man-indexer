package main

import (
	"embed"
	"fmt"
	"log"
	"manindexer/api"
	"manindexer/common"
	"manindexer/man"
	"time"
)

var (
	//go:embed web/static/* web/template/*
	f embed.FS
)

func main() {
	banner := `
    __  ___  ___     _   __
   /  |/  / /   |   / | / / v0.0.2.1
  / /|_/ / / /| |  /  |/ / 
 / /  / / / ___ | / /|  /  
/_/  /_/ /_/  |_|/_/ |_/                   
 `
	fmt.Println(banner)
	common.InitConfig()
	man.InitAdapter(common.Chain, common.Db, common.TestNet, common.Server)
	log.Printf("ManIndex,chain=%s,test=%s,db=%s,server=%s", common.Chain, common.TestNet, common.Db, common.Server)
	if common.Server == "1" {
		go api.Start(f)
	}
	go man.ZmqRun()
	mm := man.ManMempool{}
	go mm.CheckMempool("btc")
	// if *reindex != "" {
	// 	arr := strings.Split(*reindex, ":")
	// 	from, err1 := strconv.ParseInt(arr[0], 10, 64)
	// 	to, err2 := strconv.ParseInt(arr[1], 10, 64)
	// 	if err1 == nil && err2 == nil {
	// 		man.IndexerRun(from, to)
	// 	}
	// }
	for {
		man.IndexerRun()
		man.CheckNewBlock()
		time.Sleep(time.Second * 10)
	}
}
