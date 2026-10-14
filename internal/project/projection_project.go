package project

type Project struct {
	ID            string           `json:"id"`
	Owner         string           `json:"owner"`
	Visibility    string           `json:"visibility"`
	SSOTScope     string           `json:"ssot_scope"`
	LocalPath     string           `json:"local_path"`
	Remote        string           `json:"remote"`
	Role          string           `json:"role"`
	Consumes      []string         `json:"consumes"`
	Health        RepositoryHealth `json:"health"`
	LocalPresent  bool             `json:"local_present"`
	GitPresent    bool             `json:"git_present"`
	RemoteMatches bool             `json:"remote_matches"`
}
