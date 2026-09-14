package main

// mangaBakaSourceName is mappingSource.name for the MangaBaka entry, shared
// with tests that need to pick that one descriptor out of the table.
const mangaBakaSourceName = "MangaBaka"

// mappingSource describes one ID-mapping source's CLI flag, env var prefix,
// and config defaults. cli.go and config.go are driven from this table
// instead of hand-wiring each source in multiple places — adding a source
// now means one entry here, plus a client and a strategy.
type mappingSource struct {
	name     string
	flag     string
	urlFlag  string // "" if the source has no base-URL flag/env var
	usage    string
	urlUsage string

	envPrefix          string
	defaultEnabled     bool
	defaultBaseURL     string
	defaultCacheDir    func() string // nil if the source has no cache
	defaultCacheMaxAge string

	// configField returns a pointer to the source's field on cfg.
	configField func(cfg *Config) *MappingSourceConfig
}

// setConfig writes the source's config into cfg.
func (s mappingSource) setConfig(cfg *Config, v MappingSourceConfig) {
	*s.configField(cfg) = v
}

// mappingSources lists the ARM, Hato, MangaBaka and Jikan sources. Offline
// database is not here: it has its own config shape (AutoUpdate,
// ForceRefresh) and no base-URL flag.
func mappingSources() []mappingSource {
	return []mappingSource{
		{
			name:           "ARM",
			flag:           flagARMAPI,
			urlFlag:        flagARMAPIURL,
			usage:          "enable ARM API for anime ID mapping (ignored for --manga, fallback after offline DB) (default: false)",
			urlUsage:       "ARM API base URL",
			envPrefix:      "ARM_API_",
			defaultEnabled: false,
			defaultBaseURL: defaultARMBaseURL,
			configField:    func(cfg *Config) *MappingSourceConfig { return &cfg.ARMAPI },
		},
		{
			name:               "Hato",
			flag:               flagHatoAPI,
			urlFlag:            flagHatoAPIURL,
			usage:              "enable Hato API for anime and manga ID mapping (default: true)",
			urlUsage:           "Hato API base URL",
			envPrefix:          "HATO_API_",
			defaultEnabled:     true,
			defaultBaseURL:     defaultHatoBaseURL,
			defaultCacheDir:    getDefaultHatoCacheDir,
			defaultCacheMaxAge: defaultLongCacheMaxAge,
			configField:        func(cfg *Config) *MappingSourceConfig { return &cfg.HatoAPI },
		},
		{
			name:               mangaBakaSourceName,
			flag:               flagMangaBakaAPI,
			urlFlag:            flagMangaBakaAPIURL,
			usage:              "enable MangaBaka API for manga ID mapping (fallback after Hato) (default: false)",
			urlUsage:           "MangaBaka API base URL",
			envPrefix:          "MANGABAKA_API_",
			defaultEnabled:     false,
			defaultBaseURL:     defaultMangaBakaBaseURL,
			defaultCacheDir:    getDefaultMangaBakaCacheDir,
			defaultCacheMaxAge: defaultLongCacheMaxAge,
			configField:        func(cfg *Config) *MappingSourceConfig { return &cfg.MangaBakaAPI },
		},
		{
			name:               "Jikan",
			flag:               flagJikanAPI,
			usage:              "enable Jikan API for manga ID mapping (default: false)",
			envPrefix:          "JIKAN_API_",
			defaultEnabled:     false,
			defaultCacheDir:    getDefaultJikanCacheDir,
			defaultCacheMaxAge: "168h",
			configField:        func(cfg *Config) *MappingSourceConfig { return &cfg.JikanAPI },
		},
	}
}
