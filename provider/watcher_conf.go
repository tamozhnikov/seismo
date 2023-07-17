package provider

// ProviderType represents types (implementations) of message sources.
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

// WatcherConfig contains configurations of a message watcher.
type WatcherConfig struct {
	// Id specifies the identifier of the message source.
	Id string `json:"id"`

	// T specifies the type of the message source (source provider).
	T ProviderType `json:"t"`

	// ConnStr specifies a connection string for the message source.
	// E.g., the connection string can be an address of the source.
	ConnStr string `json:"conn_str"`

	//Timeout specifies a timeout value for connection to the message source.
	Timeout uint `json:"timeout"`

	// CheckPeriod specifies a period of checking the appearance of new messages.
	CheckPeriod uint `json:"check_period"`
}

// DefaultWatcherConfig returns a watcher configuration with default values.
// Use this function for test purpose only.
func DefaultWatcherConfig() WatcherConfig {
	return WatcherConfig{Id: DefId, T: DefT, ConnStr: DefConnStr, Timeout: DefTimeout, CheckPeriod: DefCheckPeriod}
}
