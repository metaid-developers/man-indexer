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
	log.Printf("ManIndex,chain=%s,test=%s,db=%s,server=%s,config=%s", common.Chain, common.TestNet, common.Db, common.Server, common.ConfigFile)
	if common.Server == "1" {
		go api.Start(f)
	}
	go man.ZmqRun()
	// chainList := strings.Split(common.Chain, ",")
	// for _, chainName := range chainList {
	// 	mm := man.ManMempool{}
	// 	go mm.CheckMempool(chainName)
	// }
	for {
		// man.IndexerRun()
		// man.CheckNewBlock()
		time.Sleep(time.Second * 10)
	}
}
