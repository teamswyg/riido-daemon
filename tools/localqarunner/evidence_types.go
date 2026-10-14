package main

type runArtifacts struct {
	ProviderEvidence  string `json:"provider_evidence"`
	ProductEvidence   string `json:"product_evidence,omitempty"`
	ReleaseEvidence   string `json:"release_evidence,omitempty"`
	CoverageEvidence  string `json:"coverage_evidence,omitempty"`
	PromotionRegistry string `json:"promotion_registry,omitempty"`
	ManualEvidence    string `json:"manual_evidence,omitempty"`
	DomainCache       string `json:"domain_cache,omitempty"`
	ProductLab        string `json:"product_lab,omitempty"`
	ScheduleEvidence  string `json:"schedule_evidence,omitempty"`
	InfraEvidence     string `json:"infra_evidence,omitempty"`
	DashboardHTML     string `json:"dashboard_html"`
	S3Prefix          string `json:"s3_prefix,omitempty"`
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
