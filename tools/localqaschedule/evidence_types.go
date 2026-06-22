package main

type scheduleEvidence struct {
	SchemaVersion       string `json:"schema_version"`
	ID                  string `json:"id"`
	Status              string `json:"status"`
	Label               string `json:"label"`
	Installed           bool   `json:"installed"`
	PlistPath           string `json:"plist_path"`
	StdoutPath          string `json:"stdout_path"`
	StderrPath          string `json:"stderr_path"`
	Hour                int    `json:"hour"`
	Minute              int    `json:"minute"`
	RunAtLoad           bool   `json:"run_at_load"`
	S3PrefixConfigured  bool   `json:"s3_prefix_configured"`
	TaskMutations       bool   `json:"task_mutations"`
	TaskIDConfigured    bool   `json:"task_id_configured"`
	CommandHasTokenText bool   `json:"command_has_token_text"`
	CommandPreview      string `json:"command_preview"`
}
