package main

type entrypointRouteMap struct {
	SchemaVersion string            `json:"schema_version"`
	ID            string            `json:"id"`
	Title         string            `json:"title"`
	GeneratedDoc  string            `json:"generated_doc"`
	Loop          evidenceLoop      `json:"loop"`
	Routes        []entrypointRoute `json:"routes"`
}

type entrypointRoute struct {
	ID          string   `json:"id"`
	Owner       string   `json:"owner"`
	Description string   `json:"description"`
	Includes    []string `json:"includes"`
}
