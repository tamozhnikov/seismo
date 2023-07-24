//go:build outtest
// +build outtest

package seishub

import (
	"context"
	"seismo/provider"
	"testing"
)

func Test_ExtractMessages(t *testing.T) {
	c := provider.WatcherConfig{Id: "seishub", Timeout: 120, CheckPeriod: 2}
	ext, err := NewHub(c)
	if err != nil {
		t.Fatalf("ExtractMessages: error: %v", err)
	}
	msgs, err := ext.Extract(context.Background(),
		provider.MonthYear{Month: 6, Year: 2023}, provider.MonthYear{Month: 7, Year: 2023}, 7)
	if err != nil {
		t.Fatalf("ExtractMessages: error: %v", err)
	}
	t.Log(len(msgs))
}
