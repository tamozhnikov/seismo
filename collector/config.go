package collector

import (
	"encoding/json"
	"fmt"
	"os"
	"seismo/collector/db"
	"seismo/provider"
)

// Config represents configurations of the Collector application.
type Config struct {
	// Watchers specifies types ans settings of watchers that should
	// be started by Collector.
	Watchers map[string]provider.WatcherConfig `json:"watchers"`

	//Db specifies database configurations of collector.
	Db db.DbConfig `json:"db"`

	//MaintainPeriod specifies the period to check and restart watchers.
	MaintainPeriod uint `json:"maintain_period"`
}

const (
	// configFileVar is the name of an environment variable that contains
	// the full path of the Collector configuration file.
	configFileVar = "SEISMO_COLLECTOR_CONFIG"

	defMaintainPeriod uint = 2
)

// DefaultConfig returns a Config instance with default values.
func DefaultConfig() Config {
	var c = Config{}
	c.Watchers = make(map[string]provider.WatcherConfig, 1)
	wc := provider.DefaultWatcherConfig()
	c.Watchers[wc.Id] = wc
	c.Db = db.DefaultDbConfig()
	c.MaintainPeriod = defMaintainPeriod
	return c
}

// GetConfig gets the name of the config file from an environment variable
// and reads collector configurations from this file.
//
// The function returns a Config value and an error. If the error is not nil,
// the returned Config value is nil.
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

// ConfigFileNameFromEnv reads the full name of a config file
// from an environment variable.
//
// The function returns a name and an error. If the error is not nil,
// the returned name is nil.
func ConfigFileNameFromEnv() (string, error) {
	fn := os.Getenv(configFileVar)
	if fn == "" {
		return fn, fmt.Errorf("ConfigFileNameFromEnv: Cannot read the %q environment variable", configFileVar)
	}

	return fn, nil
}

// ConfigFromFile reads configurations from a specified json file.
// The "name" parameter specifies the name of a file.
//
// The function returns a Config value and an error. If the error is not nil,
// the returned Config value is nil.
func ConfigFromFile(name string) (Config, error) {
	var c = Config{}

	buf, err := os.ReadFile(name)
	if err != nil {
		return c, fmt.Errorf("ConfigFromFile: name: %s, error: %w", name, err)
	}

	err = json.Unmarshal(buf, &c)
	if err != nil {
		return c, fmt.Errorf("ConfigFromFile: name: %s, error: %w", name, err)
	}

	return c, nil
}
