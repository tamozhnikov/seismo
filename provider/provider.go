package provider

import (
	"fmt"
	"seismo"
	"seismo/provider/pseudo"
	"seismo/provider/seishub"
	"time"
)

type ProviderType string

const (
	Pseudo  ProviderType = "pseudo"
	Seishub ProviderType = "seishub"

	//default values
	defId      string       = "pseudo_1"
	defT       ProviderType = Pseudo
	defConnStr string       = ""
	defTimeout uint         = 120
)

type WatcherConfig struct {
	Id      string       `json:"id"`
	T       ProviderType `json:"t"`
	ConnStr string       `json:"conn_str"`
	Timeout uint         `json:"timeout"`
	//StartFrom   time.Time     `json:"start_from"`
	CheckPeriod uint `json:"check_period"`
}

// NewWatcher localizes creating watchers depending on a specified type of provider
func NewWatcher(conf WatcherConfig) (seismo.Watcher, error) {
	switch conf.T {
	case Pseudo:
		return pseudo.NewHub(conf.Id), nil
	case Seishub:
		return seishub.NewHub(conf.Id, conf.ConnStr, time.Duration(conf.Timeout)*time.Second), nil
	default:
		return nil, fmt.Errorf("Unknown watcher type: %q", conf.T)
	}
}

func DefaultWatcherConfig() WatcherConfig {
	return WatcherConfig{Id: defId, T: defT, ConnStr: defConnStr, Timeout: defTimeout}
}
