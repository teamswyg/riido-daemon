package main

import "time"

const (
	statusPassed  = "passed"
	statusPartial = "partial"
	statusFailed  = "failed"
)

type config struct {
	repo                 *string
	providerEvidence     *string
	productEvidence      *string
	releaseEvidence      *string
	coverageEvidence     *string
	manualEvidence       *string
	domainCache          *string
	productLab           *string
	scheduleEvidence     *string
	infraEvidence        *string
	runEvidence          *string
	dashboardHTML        *string
	coverageManifest     *string
	s3Prefix             *string
	validFor             *time.Duration
	providerTool         *string
	productTool          *string
	releaseTool          *string
	scheduleTool         *string
	dashboardTool        *string
	clientRoot           *string
	productAgentHost     *string
	productRiidoHost     *string
	productBaseURL       *string
	productWorkspace     *string
	productTeamID        *string
	productScreenshots   *string
	productStorage       *string
	productTaskID        *string
	productAgentID1      *string
	productAgentID2      *string
	productCommentBody   *string
	runIntegration       *bool
	runRelease           *bool
	runProduct           *bool
	productMutations     *bool
	productBrowserE2E    *bool
	productStartClient   *bool
	productTaskFixture   *bool
	productPrepareDaemon *bool
	continueOnFailure    *bool
	strictCoverage       *bool
}

type runEvidence struct {
	SchemaVersion    string                `json:"schema_version"`
	ID               string                `json:"id"`
	ObservedAt       string                `json:"observed_at"`
	ExpiresAt        string                `json:"expires_at"`
	Status           string                `json:"status"`
	CoverageStatus   string                `json:"coverage_status"`
	ProviderStatus   string                `json:"provider_status,omitempty"`
	ProviderSummary  []runProviderSummary  `json:"provider_summary,omitempty"`
	StrictCoverage   bool                  `json:"strict_coverage,omitempty"`
	Coverage         *runCoverage          `json:"coverage,omitempty"`
	DeploymentGate   runDeploymentGate     `json:"deployment_gate"`
	Artifacts        runArtifacts          `json:"artifacts"`
	OpenRepairs      []runRepair           `json:"open_repairs,omitempty"`
	CandidateSummary closedLoopSummary     `json:"closed_loop_candidate_summary"`
	Candidates       []closedLoopCandidate `json:"closed_loop_candidates"`
	Steps            []stepEvidence        `json:"steps"`
}
