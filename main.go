package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"runtime"

	"github.com/cmertz/repo-sync/sync"
)

func main() {
	var configFilePath string

	var concurrent int

	flag.IntVar(&concurrent, "concurrent", runtime.GOMAXPROCS(0), "")
	flag.StringVar(&configFilePath, "config", "", "")
	flag.Parse()

	s, err := Config(configFilePath).syncs(context.TODO())
	if err != nil {
		log.Fatalln(err)
	}

	sync.Run(context.TODO(), s, concurrent, func(e error) {
		// TODO: add a better way for handling errors
		fmt.Println(e)
	})
}
