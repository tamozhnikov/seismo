package seismo

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
	StartWatch(ctx context.Context, from time.Time, checkPeriod time.Duration) (<-chan Message, error)
	StateInfo() WatcherStateInfo
	SetId(id string)
	GetId() string
}

// AlreadyRunErr indicates that a watcher is already running (watching)
type AlreadyRunErr struct {
}

func (e AlreadyRunErr) Error() string {
	return "Already running"
}
