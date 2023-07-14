package provider

type ProviderType string

const (
	Pseudo  ProviderType = "pseudo"
	Seishub ProviderType = "seishub"

	//default values

	DefId          string       = "pseudo_1"
	DefT           ProviderType = Pseudo
	DefConnStr     string       = ""
	DefTimeout     uint         = 120
	DefCheckPeriod uint         = 2
)

// WatcherConfig represents
type WatcherConfig struct {
	Id          string       `json:"id"`
	T           ProviderType `json:"t"`
	ConnStr     string       `json:"conn_str"`
	Timeout     uint         `json:"timeout"`
	CheckPeriod uint         `json:"check_period"`
}

// DefaultWatcherConfig returns a watcher configuration with default values.
// Use this function for test purpose only.
func DefaultWatcherConfig() WatcherConfig {
	return WatcherConfig{Id: DefId, T: DefT, ConnStr: DefConnStr, Timeout: DefTimeout, CheckPeriod: DefCheckPeriod}
}
