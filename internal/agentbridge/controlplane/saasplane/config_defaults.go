package saasplane

import (
	"net/http"
	"strings"
)

func defaultConfig(cfg Config) (Config, *http.Client) {
	cfg.BaseURL = strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	cfg.DaemonID = strings.TrimSpace(cfg.DaemonID)
	cfg.DeviceID = strings.TrimSpace(cfg.DeviceID)
	cfg.DeviceSecret = strings.TrimSpace(cfg.DeviceSecret)
	cfg.Profile = strings.TrimSpace(cfg.Profile)
	cfg.AppVersion = strings.TrimSpace(cfg.AppVersion)
	cfg.BearerToken = strings.TrimSpace(cfg.BearerToken)
	if cfg.DeviceID == "" {
		cfg.DeviceID = cfg.DaemonID
	}
	cfg.Agents = normalizeAgents(cfg.Agents)
	if cfg.LongPollWait <= 0 {
		cfg.LongPollWait = defaultLongPollWait
	}
	minRequestTimeout := cfg.LongPollWait + longPollRequestTimeoutPad
	if cfg.RequestTimeout <= 0 || cfg.RequestTimeout < minRequestTimeout {
		cfg.RequestTimeout = minRequestTimeout
	}
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: cfg.RequestTimeout}
	}
	return cfg, client
}
