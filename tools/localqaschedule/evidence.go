package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

func writeScheduleEvidence(
	cfg config,
	paths schedulePaths,
	command string,
	live launchdEvidence,
) error {
	evidence := scheduleEvidence{
		SchemaVersion:       "riido-local-qa-schedule.v1",
		ID:                  "local-qa-schedule",
		Status:              "passed",
		Label:               *cfg.label,
		Installed:           *cfg.install || live.Loaded,
		PlistPath:           paths.plist,
		StdoutPath:          paths.stdout,
		StderrPath:          paths.stderr,
		Hour:                *cfg.hour,
		Minute:              *cfg.minute,
		RunAtLoad:           *cfg.runAtLoad,
		S3PrefixConfigured:  strings.TrimSpace(*cfg.s3Prefix) != "",
		CoverageEvidence:    *cfg.coverageEvidence,
		TaskMutations:       *cfg.taskMutations,
		TaskIDConfigured:    strings.TrimSpace(*cfg.productTaskID) != "",
		CommandHasTokenText: commandMentionsToken(command),
		CommandPreview:      safeCommandPreview(command),
		Launchd:             live,
	}
	return writeJSON(scheduleEvidencePath(cfg, paths), evidence)
}

func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("encode schedule evidence: %w", err)
	}
	return writeText(path, string(data)+"\n")
}

func commandMentionsToken(command string) bool {
	return strings.Contains(strings.ToLower(command), "token")
}

func safeCommandPreview(command string) string {
	if commandMentionsToken(command) {
		return "[redacted: command contains token text]"
	}
	return command
}

func scheduleEvidencePath(cfg config, paths schedulePaths) string {
	if filepath.IsAbs(*cfg.evidenceOut) {
		return *cfg.evidenceOut
	}
	return filepath.Join(paths.repo, *cfg.evidenceOut)
}
