package main

import "time"

const (
	statusPassed  = "passed"
	statusFailed  = "failed"
	statusPartial = "partial"
	statusSkipped = "skipped"
)

type config struct {
	clientRoot    *string
	baseURL       *string
	apiToken      *string
	workspaceID   *string
	taskID        *string
	firstAgentID  *string
	secondAgentID *string
	evidenceOut   *string
	labOut        *string
	screenshots   *string
	storageState  *string
	figmaManifest *string
	figmaGolden   *string
	validFor      *time.Duration
	probeRoutes   *bool
	browserE2E    *bool
	startClient   *bool
	agentHost     *string
	riidoAPIHost  *string
	teamID        *string
	taskFixture   *bool
	runMutations  *bool
	commentBody   *string
	prepareDaemon *bool
	daemonBinary  *string
	daemonRunDir  *string
	daemonSlots   *int
}

type evidenceFile struct {
	SchemaVersion string     `json:"schema_version"`
	ID            string     `json:"id"`
	ObservedAt    string     `json:"observed_at"`
	ExpiresAt     string     `json:"expires_at"`
	Status        string     `json:"status"`
	Scenarios     []scenario `json:"scenarios"`
}

type scenario struct {
	ID             string         `json:"id"`
	Status         string         `json:"status"`
	Method         string         `json:"method,omitempty"`
	Endpoint       string         `json:"endpoint,omitempty"`
	FailureSummary string         `json:"failure_summary,omitempty"`
	Screenshot     string         `json:"screenshot,omitempty"`
	Observed       map[string]any `json:"observed,omitempty"`
	Repair         *repair        `json:"repair,omitempty"`
}

type repair struct {
	Class            string `json:"class"`
	Owner            string `json:"owner"`
	Mode             string `json:"mode"`
	Summary          string `json:"summary"`
	SuggestedCommand string `json:"suggested_command,omitempty"`
}
