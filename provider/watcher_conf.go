package provider

type ProviderType string

const (
	Pseudo  ProviderType = "pseudo"
	Seishub ProviderType = "seishub"

	//default values
	defId          string       = "pseudo_1"
	defT           ProviderType = Pseudo
	defConnStr     string       = ""
	defTimeout     uint         = 120
	defCheckPeriod uint         = 2
)

type WatcherConfig struct {
	Id          string       `json:"id"`
	T           ProviderType `json:"t"`
	ConnStr     string       `json:"conn_str"`
	Timeout     uint         `json:"timeout"`
	CheckPeriod uint         `json:"check_period"`
}

func DefaultWatcherConfig() WatcherConfig {
	return WatcherConfig{Id: defId, T: defT, ConnStr: defConnStr, Timeout: defTimeout, CheckPeriod: defCheckPeriod}
}
