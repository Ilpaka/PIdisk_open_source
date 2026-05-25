package domain

type AppSettings struct {
	Theme               string   `toml:"theme" json:"theme"`
	Language            string   `toml:"language" json:"language"`
	SyncIntervalSeconds uint64   `toml:"sync_interval_seconds" json:"syncIntervalSeconds"`
	DefaultIgnoredPaths []string `toml:"default_ignored_paths" json:"defaultIgnoredPaths"`
	PrometheusEnabled   bool     `toml:"prometheus_enabled" json:"prometheusEnabled"`
	PrometheusAddr      string   `toml:"prometheus_addr" json:"prometheusAddr"`
	LogLevel            string   `toml:"log_level" json:"logLevel"`
	ShowHidden          bool     `toml:"show_hidden" json:"showHidden"`
}

func DefaultAppSettings() AppSettings {
	return AppSettings{
		Theme:               "system",
		Language:            "en",
		SyncIntervalSeconds: 30,
		DefaultIgnoredPaths: []string{".DS_Store", "Thumbs.db", "*.tmp", "*.swp", ".git/", "node_modules/"},
		PrometheusEnabled:   false,
		PrometheusAddr:      "127.0.0.1:9101",
		LogLevel:            "info",
		ShowHidden:          false,
	}
}
