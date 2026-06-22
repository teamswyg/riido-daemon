package main

import "time"

const (
	statusPassed = "passed"
	statusFailed = "failed"
)

type config struct {
	repo               *string
	providerEvidence   *string
	productEvidence    *string
	productLab         *string
	scheduleEvidence   *string
	runEvidence        *string
	dashboardHTML      *string
	coverageManifest   *string
	s3Prefix           *string
	validFor           *time.Duration
	providerTool       *string
	productTool        *string
	dashboardTool      *string
	clientRoot         *string
	productAgentHost   *string
	productBaseURL     *string
	productWorkspace   *string
	productScreenshots *string
	productStorage     *string
	productTaskID      *string
	productAgentID1    *string
	productAgentID2    *string
	productCommentBody *string
	runIntegration     *bool
	runProduct         *bool
	productMutations   *bool
	productBrowserE2E  *bool
	productStartClient *bool
	continueOnFailure  *bool
}

type runEvidence struct {
	SchemaVersion string         `json:"schema_version"`
	ID            string         `json:"id"`
	ObservedAt    string         `json:"observed_at"`
	ExpiresAt     string         `json:"expires_at"`
	Status        string         `json:"status"`
	Artifacts     runArtifacts   `json:"artifacts"`
	Steps         []stepEvidence `json:"steps"`
}

type runArtifacts struct {
	ProviderEvidence string `json:"provider_evidence"`
	ProductEvidence  string `json:"product_evidence,omitempty"`
	ProductLab       string `json:"product_lab,omitempty"`
	ScheduleEvidence string `json:"schedule_evidence,omitempty"`
	DashboardHTML    string `json:"dashboard_html"`
	S3Prefix         string `json:"s3_prefix,omitempty"`
}

type stepEvidence struct {
	ID         string `json:"id"`
	Status     string `json:"status"`
	Command    string `json:"command"`
	ExitCode   int    `json:"exit_code"`
	OutputTail string `json:"output_tail,omitempty"`
}

type uploadSpec struct {
	id        string
	source    string
	target    string
	recursive bool
}
