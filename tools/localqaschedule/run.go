package main

import "fmt"

func run(cfg config) (string, error) {
	if err := validateTime(*cfg.hour, *cfg.minute); err != nil {
		return "", err
	}
	paths, err := resolvePaths(cfg)
	if err != nil {
		return "", err
	}
	body := renderPlist(cfg, paths)
	if err := writeText(paths.plist, body); err != nil {
		return "", fmt.Errorf("write plist: %w", err)
	}
	if *cfg.install {
		if err := installLaunchAgent(paths); err != nil {
			return "", err
		}
	}
	if err := writeScheduleEvidence(cfg, paths, body); err != nil {
		return "", err
	}
	return paths.plist, nil
}

func validateTime(hour, minute int) error {
	if hour < 0 || hour > 23 {
		return fmt.Errorf("hour must be 0..23")
	}
	if minute < 0 || minute > 59 {
		return fmt.Errorf("minute must be 0..59")
	}
	return nil
}
