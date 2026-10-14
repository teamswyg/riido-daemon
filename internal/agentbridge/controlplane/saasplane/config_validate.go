package saasplane

import (
	"errors"
	"fmt"
	"net/url"
)

func validateConfig(cfg Config) error {
	if cfg.BaseURL == "" {
		return errors.New("saasplane: BaseURL is required")
	}
	if _, err := url.ParseRequestURI(cfg.BaseURL); err != nil {
		return fmt.Errorf("saasplane: invalid BaseURL: %w", err)
	}
	if cfg.DaemonID == "" {
		return errors.New("saasplane: DaemonID is required")
	}
	if cfg.DeviceSecret != "" && cfg.DeviceID == "" {
		return errors.New("saasplane: DeviceID is required when DeviceSecret is set")
	}
	if len(cfg.Agents) == 0 && cfg.DeviceSecret == "" {
		return errors.New("saasplane: at least one static agent binding or a device credential is required")
	}
	return nil
}
