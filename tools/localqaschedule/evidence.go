package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

type scheduleEvidence struct {
	SchemaVersion       string `json:"schema_version"`
	ID                  string `json:"id"`
	Status              string `json:"status"`
	PlistPath           string `json:"plist_path"`
	Hour                int    `json:"hour"`
	Minute              int    `json:"minute"`
	S3PrefixConfigured  bool   `json:"s3_prefix_configured"`
	TaskMutations       bool   `json:"task_mutations"`
	TaskIDConfigured    bool   `json:"task_id_configured"`
	CommandHasTokenText bool   `json:"command_has_token_text"`
}

func writeScheduleEvidence(cfg config, paths schedulePaths, command string) error {
	evidence := scheduleEvidence{
		SchemaVersion:       "riido-local-qa-schedule.v1",
		ID:                  "local-qa-schedule",
		Status:              "passed",
		PlistPath:           paths.plist,
		Hour:                *cfg.hour,
		Minute:              *cfg.minute,
		S3PrefixConfigured:  strings.TrimSpace(*cfg.s3Prefix) != "",
		TaskMutations:       *cfg.taskMutations,
		TaskIDConfigured:    strings.TrimSpace(*cfg.productTaskID) != "",
		CommandHasTokenText: commandMentionsToken(command),
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

func scheduleEvidencePath(cfg config, paths schedulePaths) string {
	if filepath.IsAbs(*cfg.evidenceOut) {
		return *cfg.evidenceOut
	}
	return filepath.Join(paths.repo, *cfg.evidenceOut)
}
