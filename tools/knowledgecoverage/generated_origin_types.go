package main

type generatedOrigin struct {
	Generator string   `json:"generator"`
	Count     int      `json:"count"`
	Samples   []string `json:"samples"`
}

type generatedOriginWorkflowCoverage struct {
	CoveredCount int                           `json:"covered_count"`
	MissingCount int                           `json:"missing_count"`
	Missing      []generatedOriginWorkflowMiss `json:"missing"`
}

type generatedOriginWorkflowMiss struct {
	Generator string `json:"generator"`
	Tool      string `json:"tool"`
	Count     int    `json:"count"`
}
