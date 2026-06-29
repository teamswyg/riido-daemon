package main

type closedLoopMaturitySpec struct {
	SchemaVersion string             `json:"schema_version"`
	ID            string             `json:"id"`
	Generated     string             `json:"generated"`
	StaleAfter    int                `json:"stale_after_days"`
	Metrics       []closedLoopMetric `json:"metrics"`
	PartialWhen   []string           `json:"partial_when"`
}

type closedLoopMetric struct {
	ID       string `json:"id"`
	Class    string `json:"class"`
	Mode     string `json:"mode"`
	Evidence string `json:"evidence"`
}

type closedLoopCoverage struct {
	Scenarios []struct {
		ID       string `json:"id"`
		Surface  string `json:"surface"`
		Evidence string `json:"evidence"`
	} `json:"scenarios"`
}
