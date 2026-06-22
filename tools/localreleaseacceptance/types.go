package main

import "time"

const (
	statusPassed = "passed"
	statusFailed = "failed"
)

type options struct {
	Repo          string
	EvidenceOut   string
	ValidFor      time.Duration
	ReleaseAPIURL string
}

type evidenceFile struct {
	SchemaVersion string     `json:"schema_version"`
	ID            string     `json:"id"`
	ObservedAt    string     `json:"observed_at"`
	ExpiresAt     string     `json:"expires_at"`
	Status        string     `json:"status"`
	Artifacts     artifacts  `json:"artifacts"`
	Scenarios     []scenario `json:"scenarios"`
}

type artifacts struct {
	InstallDir       string   `json:"install_dir"`
	InstalledBinary  string   `json:"installed_binary"`
	VersionOutput    string   `json:"version_output,omitempty"`
	LatestReleaseTag string   `json:"latest_release_tag,omitempty"`
	ExpectedAsset    string   `json:"expected_asset,omitempty"`
	ReleaseAssets    []string `json:"release_assets,omitempty"`
}

type scenario struct {
	ID             string `json:"id"`
	Status         string `json:"status"`
	FailureSummary string `json:"failure_summary,omitempty"`
}

type installFixture struct {
	assetDir   string
	binDir     string
	installDir string
	marker     string
}
