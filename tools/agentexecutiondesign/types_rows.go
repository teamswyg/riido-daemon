package main

type linkRef struct {
	Title string `json:"title"`
	Path  string `json:"path"`
}

type problemRow struct {
	ID        string `json:"id"`
	Symptom   string `json:"symptom"`
	Cause     string `json:"cause"`
	Direction string `json:"direction"`
}

type observationRow struct {
	Observation string `json:"observation"`
	SSOT        string `json:"ssot"`
	Meaning     string `json:"meaning"`
}

type fieldMeaning struct {
	Field   string `json:"field"`
	Meaning string `json:"meaning"`
}

type phaseRule struct {
	Field string `json:"field"`
	Phase string `json:"phase"`
	Rule  string `json:"rule"`
}
