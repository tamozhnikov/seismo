package provider

import (
	"context"
	"time"
)

type WatcherStateInfo int

const (
	//watcher states
	Stopped WatcherStateInfo = iota
	Run
)

type Watcher interface {
	StartWatch(ctx context.Context, from time.Time) (<-chan Message, error)
	StateInfo() WatcherStateInfo
	GetConfig() WatcherConfig
}

// AlreadyRunErr indicates that a watcher is already running (watching)
type AlreadyRunErr struct {
}

func (e AlreadyRunErr) Error() string {
	return "Already running"
}
