package provider

import (
	"context"
	"time"
)

// WatcherStateInfo contains infomation about the state of a watcher.
type WatcherStateInfo string

const (
	//watcher states

	Stopped WatcherStateInfo = "Stopped"
	Run     WatcherStateInfo = "Run"
)

// Watcher represents a type which can start watching seismic activity (StartWatch method),
// i.e. waiting and getting seismic event messages, report its state information (StateInfo method),
// and return its configuration (GetConfig methods).
//
// As a rule one instance of an interface implementation is created for a separate message source.
type Watcher interface {
	// StartWatch starts watching (waiting and getting) seismic event messages
	// beginning from the time, specified by the "from" argument.
	// The method returns a new channel, from which messages are received, and
	// an error. If watching is already started (the watcher is running),
	// the method returns an AlreadyRunErr error and nil channel.
	//
	// If the returned error is not nil, the returned channel is nil.
	//
	// Watching can be stopped (cancelled) through context (the "ctx" argument).
	// As a result of cancellation the message channel will be closed.
	StartWatch(ctx context.Context, from time.Time) (<-chan Message, error)

	//StateInfo returns information about the current state of the watcher.
	StateInfo() WatcherStateInfo

	//GetConfig returns the current configuration of the watcher.
	GetConfig() WatcherConfig
}

// AlreadyRunErr indicates that a watcher is already running (watching)
type AlreadyRunErr struct {
}

func (e AlreadyRunErr) Error() string {
	return "Already running"
}
