package collector

import (
	"encoding/json"
	"fmt"
	"os"
	"seismo/collector/db"
	"seismo/provider"
)

type Config struct {
	Watchers       map[string]provider.WatcherConfig `json:"watchers"`
	Db             db.DbConfig                       `json:"db"`
	MaintainPeriod uint                              `json:"maintain_period"`
}

const (
	configFileVar = "SEISMO_COLLECTOR_CONFIG"

	defMaintainPeriod uint = 2
)

func DefaultConfig() Config {
	var c = Config{}
	c.Watchers = make(map[string]provider.WatcherConfig, 1)
	wc := provider.DefaultWatcherConfig()
	c.Watchers[wc.Id] = wc
	c.Db = db.DefaultDbConfig()
	c.MaintainPeriod = defMaintainPeriod
	return c
}

func GetConfig() (Config, error) {
	fn, err := ConfigFileNameFromEnv()
	if err != nil {
		return Config{}, fmt.Errorf("GetConfig: %w", err)
	}

	c, err := ConfigFromFile(fn)
	if err != nil {
		return Config{}, fmt.Errorf("GetConfig: %w", err)
	}

	return c, nil
}

// configFileNameFromEnv reads the full name of a config file from an environment variable
func ConfigFileNameFromEnv() (string, error) {
	fn := os.Getenv(configFileVar)
	if fn == "" {
		return fn, fmt.Errorf("ConfigFileNameFromEnv: Cannot read the %q environment variable", configFileVar)
	}

	return fn, nil
}

// configFromFile reads configurations from a specified json file.
func ConfigFromFile(name string) (Config, error) {
	var c = Config{}

	buf, err := os.ReadFile(name)
	if err != nil {
		return c, fmt.Errorf("ConfigFileNameFromEnv: name: %s, error: %w", name, err)
	}

	err = json.Unmarshal(buf, &c)
	if err != nil {
		return c, fmt.Errorf("ConfigFileNameFromEnv: name: %s, error: %w", name, err)
	}

	return c, nil
}
