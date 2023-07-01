package stubdb

import (
	"context"
	"seismo/provider"
	"time"
)

type Adapter struct{}

// SaveMsg always returns nil
func (s *Adapter) SaveMsg(ctx context.Context, msgs []provider.Message) error {
	return nil
}

// GetLastTime always returns current UTC time as focus time in the last message
// and nil as error
func (s *Adapter) GetLastTime(ctx context.Context, sorceId string) (time.Time, error) {
	return time.Now().UTC(), nil
}
