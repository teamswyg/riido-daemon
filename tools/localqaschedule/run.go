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
	command := localQACommand(cfg, paths)
	body := renderPlist(cfg, paths)
	if err := writeText(paths.plist, body); err != nil {
		return "", fmt.Errorf("write plist: %w", err)
	}
	if *cfg.install {
		if err := installLaunchAgent(paths); err != nil {
			return "", err
		}
	}
	live, err := launchdEvidenceForRun(cfg, paths)
	if err != nil {
		return "", err
	}
	if err := writeScheduleEvidence(cfg, paths, command, live); err != nil {
		return "", err
	}
	return paths.plist, nil
}

func launchdEvidenceForRun(cfg config, paths schedulePaths) (launchdEvidence, error) {
	if !*cfg.install {
		return launchdEvidence{}, nil
	}
	live, err := inspectLaunchAgent(paths, *cfg.label)
	if err != nil {
		return launchdEvidence{}, err
	}
	if !live.CalendarTrigger {
		return launchdEvidence{}, fmt.Errorf("launchd calendar trigger missing for %s", *cfg.label)
	}
	return live, nil
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
