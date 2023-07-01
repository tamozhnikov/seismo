package crt

import (
	"fmt"
	"seismo/provider"
	"seismo/provider/pseudo"
	"seismo/provider/seishub"
	"time"
)

// NewWatcher localizes creating watchers depending on a specified type of provider
func NewWatcher(conf provider.WatcherConfig) (provider.Watcher, error) {
	switch conf.T {
	case provider.Pseudo:
		return pseudo.NewHub(conf.Id), nil
	case provider.Seishub:
		return seishub.NewHub(conf.Id, conf.ConnStr, time.Duration(conf.Timeout)*time.Second), nil
	default:
		return nil, fmt.Errorf("Unknown watcher type: %q", conf.T)
	}
}
