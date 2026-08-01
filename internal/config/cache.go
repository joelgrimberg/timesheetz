package config

import (
	"encoding/json"
	"os"
	"sync/atomic"
)

// configCache holds the parsed config file so callers don't re-read and
// re-JSON-parse on every access. Many callers hit config getters several
// times per handler; without a cache each one does a disk read + JSON
// unmarshal.
//
// The cache is invalidated whenever the file is written (SaveConfig),
// whenever the config path override changes (SetConfigPathOverride), or
// whenever a caller explicitly asks (invalidateConfigCache).
var configCache atomic.Pointer[Config]

// readCachedConfig returns the current config. On cache hit it is a
// pointer copy. On cache miss it reads the file from disk, parses it,
// and stores the result in the cache before returning.
//
// The returned Config is a defensive copy — callers can mutate fields
// without affecting the cache.
func readCachedConfig() (Config, error) {
	if cached := configCache.Load(); cached != nil {
		return *cached, nil
	}

	path := GetConfigPath()
	raw, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var parsed Config
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return Config{}, err
	}

	stored := parsed
	configCache.Store(&stored)
	return parsed, nil
}

// invalidateConfigCache drops the cached config. The next getter call
// will re-read from disk. Safe to call from any goroutine.
func invalidateConfigCache() {
	configCache.Store(nil)
}
