package crt

import (
	"fmt"
	"seismo/provider"
	"seismo/provider/pseudo"
	"seismo/provider/seishub"
)

// NewWatcher localizes creating watchers depending on a specified type of provider
func NewWatcher(conf provider.WatcherConfig) (provider.Watcher, error) {
	switch conf.T {
	case provider.Pseudo:
		h, err := pseudo.NewHub(conf)
		if err != nil {
			return nil, fmt.Errorf("NewWatcher: %w", err)
		}
		return h, nil
	case provider.Seishub:
		h, err := seishub.NewHub(conf)
		if err != nil {
			return nil, fmt.Errorf("NewWatcher: %w", err)
		}
		return h, nil
	default:
		return nil, fmt.Errorf("unknown watcher type: %q", conf.T)
	}
}
