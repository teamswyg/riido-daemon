package mwsdbridge

type HarnessIndex struct {
	SchemaVersion             string       `json:"schema_version"`
	Path                      string       `json:"path"`
	RunCount                  int          `json:"run_count"`
	TopDownCount              int          `json:"top_down_count"`
	BottomUpCount             int          `json:"bottom_up_count"`
	LastDirection             string       `json:"last_direction"`
	NextDirection             string       `json:"next_direction"`
	ConsecutiveDirectionCount int          `json:"consecutive_direction_count"`
	RecentDirections          []string     `json:"recent_directions"`
	Diagnostics               []Diagnostic `json:"diagnostics"`
}
