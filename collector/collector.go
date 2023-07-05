package collector

import (
	"context"
	"fmt"
	"log"
	"seismo/collector/db"
	"seismo/provider"
	"seismo/provider/crt"
)

func CreateWatchers(conf Config) (map[string]provider.Watcher, error) {
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

// RestartWatchers checks a current state of every watcher and tries to restart if it is stopped
func RestartWatchers(ctx context.Context, watchers map[string]provider.Watcher,
	dbAdapter db.Adapter, watchPipes chan<- (<-chan provider.Message)) {

	for id, w := range watchers {

		if w.StateInfo() != provider.Stopped {
			continue
		}

		t, err := dbAdapter.GetLastTime(ctx, id)
		if err != nil {
			log.Printf("RestartWatchers: error: %v", err)
			continue
		}

		ch, err := w.StartWatch(ctx, t)
		if err != nil {
			log.Printf("RestartWatchers: error: %v", err)
			continue
		}

		watchPipes <- ch
	}
}

func MergeWatchPipes(watchPipes <-chan <-chan provider.Message) <-chan provider.Message {
	outPipe := make(chan provider.Message)

	redirect := func(p <-chan provider.Message) {
		for m := range p {
			outPipe <- m
		}
	}

	go func() {
		for p := range watchPipes {
			go redirect(p)
		}
	}()

	return outPipe
}
