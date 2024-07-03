package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"manindexer/api"
	"manindexer/man"
	"os"
	"time"
)

var (
	//go:embed web/static/* web/template/*
	f embed.FS
)

func main() {
	log.Println("hello")
	chain := flag.String("chain", "btc", "Which chain to perform indexing")
	db := flag.String("database", "mongo", "Which database to use")
	test := flag.String("test", "0", "Connect to testnet")
	server := flag.String("server", "1", "Run the explorer service")
	//manCli := flag.String("cli", "0", "Run the man cmd cli")
	flag.Parse()
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "args:\n")
		flag.PrintDefaults()
	}
	man.InitAdapter(*chain, *db, *test, *server)
	log.Printf("ManIndex,chain=%s,test=%s,db=%s,server=%s", *chain, *test, *db, *server)
	if *server == "1" {
		go api.Start(f)
	}
	//if *manCli == "1" {
	//	go cli.Execute()
	//}
	go man.ZmqRun()
	for {
		man.CheckNewBlock()
		err := man.IndexerRun()
		if err != nil {
			log.Println(err)
		}
		time.Sleep(time.Second * 10)
	}
}
