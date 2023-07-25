// Package seismo/collector provides basic types and functions
// for the Collector application. The Collector application is
// designed to collect messages about seismic activity from various
// sources and save them into a database.
package collector

import (
	"context"
	"fmt"
	"log"
	"seismo/collector/db"
	"seismo/provider"
	"seismo/provider/crt"
	"time"
)

// CreateWatchers creates instances that implement the provider.Watcher interface.
// Provider types and quantity of the instances are defined in the "conf" parameter.
//
// The first return value is a map, keys of which are identifiers of watchers (i.e.
// their event sources), and values are watchers themselves. The second returned
// value is an error. If the returned error is not nil, the returned map value is nil.
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

// RestartWatchers permanently checks a current state of every watcher in a passed
// "watchers" map. If a watcher is stopped, the function tries to start it from the
// focus time of the last message saved in the database represented by "dbAdapter".
// If starting th watcher is successful, the function put the returned message channel
// into the "watchPipes" channel (channel of channels).
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

		//if a returned last time has the 0 value, watching start time is now
		var t0 time.Time
		if t == t0 {
			log.Println("RestartWatchers: the returned last time has the 0 value, watching start time is now")
			t = time.Now().UTC()
		}

		ch, err := w.StartWatch(ctx, t)
		if err != nil {
			log.Printf("RestartWatchers: error: %v", err)
			continue
		}

		watchPipes <- ch
	}
}

// MergeWatchPipes provides permanent merging message channels coming from the "watchPipes" channel
// into a common message channel, returned by the function.
//
// All go-routines started inside the function will end, when all channels (the watchPipes and all message
// channels transmitted through it) are closed and exhausted.
func MergeWatchPipes(watchPipes <-chan (<-chan provider.Message)) <-chan provider.Message {
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
