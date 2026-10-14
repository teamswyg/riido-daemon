package main

type PolicySnapshot struct {
	SchemaVersion string            `json:"schema_version"`
	Surfaces      []SurfaceSnapshot `json:"surfaces"`
}

type SurfaceSnapshot struct {
	ID                 string   `json:"id"`
	OwnerContext       string   `json:"owner_context"`
	AllowedJSONPaths   []string `json:"allowed_json_paths"`
	ForbiddenJSONPaths []string `json:"forbidden_json_paths"`
}
