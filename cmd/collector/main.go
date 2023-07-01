package main

import (
	"context"
	"fmt"
	"log"
	"seismo/collector"
	"seismo/collector/db"
	"seismo/provider"
	"seismo/provider/crt"
)

func main() {

	// conf := collector.DefaultConfig()

	// watchers, err := createWatchers(conf)
	// if err != nil {
	// 	log.Fatalf("Collector: main: %v", err)
	// }

	// dbAdapter, err := db.NewAdapter(conf.Db)
	// if err != nil {
	// 	log.Fatalf("Collector: main: %v", err)
	// }

	// go func() {

	// }
	// var w seismo.Watcher = pseudo.NewHub("pseudo") //seishub.NewHub("seishub", "", 0)
	// ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	// //ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	// ch, err := w.StartWatch(ctx, time.Date(2023, 6, 24, 12, 0, 0, 0, time.UTC), time.Second*2)
	// if err != nil {
	// 	log.Printf("Cannot start watching: %v", err)
	// 	return
	// }

	// for m := range ch {
	// 	fmt.Println(m)
	// }
}

func createWatchers(conf collector.Config) (map[string]provider.Watcher, error) {
	watchers := make(map[string]provider.Watcher, len(conf.Watchers))

	for _, c := range conf.Watchers {
		w, err := crt.NewWatcher(c)
		if err != nil {
			return nil, fmt.Errorf("createWatchers: %w", err)
		}

		id := w.GetConfig().Id
		if _, ok := watchers[id]; ok {
			return nil, fmt.Errorf("createWatchers: duplicated watcher id in config")
		}

		watchers[id] = w
	}

	return watchers, nil
}

// maintainWatchers checks a current state of every watcher and tries to restart if it is stopped
func maintainWatchers(ctx context.Context, watchers map[string]provider.Watcher, watchConfs map[string]provider.WatcherConfig,
	dbAdapter db.Adapter, watchPipes chan<- <-chan provider.Message) {

	for id, w := range watchers {
		if _, ok := <-ctx.Done(); ok {
			return
		}

		if w.StateInfo() != provider.Stopped {
			continue
		}

		t, err := dbAdapter.GetLastTime(ctx, id)
		if err != nil {
			log.Printf("mainWatcher: error: %v", err)
			continue
		}

		ch, err := w.StartWatch(ctx, t)
		if err != nil {
			log.Printf("mainWatcher: error: %v", err)
			continue
		}

		watchPipes <- ch
	}
}
