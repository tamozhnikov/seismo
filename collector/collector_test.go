package collector

import (
	"context"
	"seismo/collector/db"
	"seismo/provider"
	"testing"
	"time"
)

func Test_RestartWatchers(t *testing.T) {
	conf := DefaultConfig()
	watchers, _ := CreateWatchers(conf)
	dbAdapter, _ := db.NewAdapter(conf.Db)
	watchPipes := make(chan (<-chan provider.Message), 10)
	ctx1, cancel_1 := context.WithCancel(context.Background())
	defer cancel_1()

	RestartWatchers(ctx1, watchers, dbAdapter, watchPipes)
	l := len(watchPipes)
	if l != 1 {
		t.Errorf("Restart watcher: want len watch pipes: 1; res len: %d", l)
	}

	//stop watching and waiting
	cancel_1()
	time.Sleep(3 * time.Second)

	ctx2, cancel_2 := context.WithCancel(context.Background())
	defer cancel_2()
	RestartWatchers(ctx2, watchers, dbAdapter, watchPipes)
	l = len(watchPipes)
	if l != 2 {
		t.Errorf("Restart watcher: want len watch pipes: 2; res len: %d", l)
	}
}
