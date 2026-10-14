package main

type figmaGoldenCatalog struct {
	SchemaVersion string              `json:"schema_version"`
	FileKey       string              `json:"file_key"`
	CapturedAt    string              `json:"captured_at"`
	Source        string              `json:"source"`
	Screens       []figmaGoldenScreen `json:"screens"`
}

type figmaGoldenScreen struct {
	ScenarioID     string `json:"scenario_id"`
	NodeID         string `json:"node_id"`
	Name           string `json:"name"`
	GoldenPath     string `json:"golden_path"`
	Width          int    `json:"width"`
	Height         int    `json:"height"`
	OriginalWidth  int    `json:"original_width"`
	OriginalHeight int    `json:"original_height"`
	SHA256         string `json:"sha256"`
	ResolvedPath   string `json:"-"`
}
