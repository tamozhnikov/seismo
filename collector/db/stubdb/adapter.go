package stubdb

import (
	"context"
	"fmt"
	"seismo/provider"
	"time"
)

// Adapter is an empty struct with fake methods.
type Adapter struct{}

// SaveMsg writes messages in stdout and always returns nil as error.
func (a *Adapter) SaveMsg(ctx context.Context, msgs []provider.Message) error {
	fmt.Println(msgs)
	return nil
}

// GetLastTime always returns current UTC time as focus time in the last message
// and nil as error.
func (a *Adapter) GetLastTime(ctx context.Context, sorceId string) (time.Time, error) {
	return time.Now().UTC(), nil
}

// Close always returns nil as error.
func (a *Adapter) Close(ctx context.Context) error {
	return nil
}

// Connect always returns nil as error.
func (a *Adapter) Connect(ctx context.Context, connStr string) error {
	return nil
}
