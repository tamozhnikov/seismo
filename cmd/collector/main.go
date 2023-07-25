// The Collector application is designed to collect messages about seismic
// activity from various sources and save them into a database.
package main

import (
	"context"
	"flag"
	"log"
	"seismo/collector"
	"seismo/collector/db"
	"seismo/provider"
	"time"
)

func main() {
	log.SetPrefix("Collector: ")
	log.Println("main: starting")

	confFileName := flag.String("confFile", "", "config file full name")
	flag.Parse()

	var err error
	if *confFileName == "" {
		log.Println("Config file name is not specified, trying to get it from an environment variable...")
		*confFileName, err = collector.ConfigFileNameFromEnv()
		if err != nil {
			log.Printf("main: cannot get config file name: error: %v\n", err)
			return
		}
	}

	//cancel can be used in case of extending Collector
	//with a special canceling gorouting
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf, err := collector.ConfigFromFile(*confFileName) //collector.ConfigFromFile("collector/testdata/double_mongo_conf.json") //DefaultConfig()
	if err != nil {
		log.Printf("main: cannot read config file: %v\n", err)
		return
	}

	watchers, err := collector.CreateWatchers(conf)
	if err != nil {
		log.Printf("main: cannot create watchers %v\n", err)
		return
	}

	dbAdapter, err := db.NewAdapter(conf.Db)
	if err != nil {
		log.Printf("main: cannot create database adaper %v\n", err)
		return
	}

	err = dbAdapter.Connect(ctx, conf.Db.ConnStr)
	if err != nil {
		log.Printf("main: cannot connect to database %v\n", err)
		return
	}
	defer dbAdapter.Close(ctx)

	watchPipes := make(chan (<-chan provider.Message))

	msgChan := collector.MergeWatchPipes(watchPipes)

	//maintaining watchers (start and restart)
	go func() {
		t := time.NewTicker(time.Duration(conf.MaintainPeriod) * time.Second)
		for {
			select {
			case <-t.C:
				collector.RestartWatchers(ctx, watchers, dbAdapter, watchPipes)
			case <-ctx.Done():
				return
			}
		}
	}()

	//main loop: getting messages from the merged channel
	//and saving in database
	for {
		select {
		case m := <-msgChan:
			err = dbAdapter.SaveMsg(ctx, []provider.Message{m})
			if err != nil {
				log.Printf("main: cannot save message in database: error: %v\n", err)
				return
			}
		case <-ctx.Done():
			log.Print("main: ended with context")
			return
		}
	}
}
