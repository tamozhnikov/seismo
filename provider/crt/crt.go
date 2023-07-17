// Package seismo/provider/crt localizes factory functions for creating
// new instances implementing abstractions of the seismo/provider package.
// The package provides an additional layer, that allows to avoid cyclic
// dependencies between the seismo/provider package and its sub-packages.
package crt

import (
	"fmt"
	"seismo/provider"
	"seismo/provider/pseudo"
	"seismo/provider/seishub"
)

// NewWatcher creats a new watcher implementation depending on a specified provider type.
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
