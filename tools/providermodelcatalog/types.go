package main

type catalog struct {
	SchemaVersion string             `json:"schema_version"`
	Providers     map[string][]model `json:"providers"`
}

type model struct {
	ModelID string `json:"model_id"`
	Label   string `json:"label"`
}
