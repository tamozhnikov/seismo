package stubdb

import (
	"context"
	"fmt"
	"seismo/provider"
	"time"
)

type Adapter struct{}

// SaveMsg always returns nil
func (a *Adapter) SaveMsg(ctx context.Context, msgs []provider.Message) error {
	fmt.Println(msgs)
	return nil
}

// GetLastTime always returns current UTC time as focus time in the last message
// and nil as error
func (a *Adapter) GetLastTime(ctx context.Context, sorceId string) (time.Time, error) {
	return time.Now().UTC(), nil
}

func (a *Adapter) Close(ctx context.Context) error {
	return nil
}

func (a *Adapter) Connect(ctx context.Context, connStr string) error {
	return nil
}
