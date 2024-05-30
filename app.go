package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"manindexer/api"
	"manindexer/man"
	"os"
	"strconv"
	"strings"
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
	reindex := flag.String("reindex", "", "reindex block height,from:to")
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
	go man.ZmqRun()
	if *reindex != "" {
		arr := strings.Split(*reindex, ":")
		from, err1 := strconv.ParseInt(arr[0], 10, 64)
		to, err2 := strconv.ParseInt(arr[1], 10, 64)
		if err1 == nil && err2 == nil {
			man.IndexerRun(from, to)
		}
	}
	for {
		man.CheckNewBlock()
		err := man.IndexerRun(0, 0)
		if err != nil {
			log.Println(err)
		}
		time.Sleep(time.Second * 10)
	}
}
