package execution

type Binding struct {
	WorkspaceID       string     `json:"workspace_id"`
	TaskID            string     `json:"task_id"`
	Repository        string     `json:"repository"`
	ChangedFiles      []string   `json:"changed_files"`
	Checks            [][]string `json:"checks"`
	TimeoutSeconds    int        `json:"timeout_seconds"`
	ContinuationLimit int        `json:"continuation_limit,omitempty"`
}

// HasChecks distinguishes a repository selection from configured host verification.
func (b Binding) HasChecks() bool { return len(b.ChangedFiles) > 0 && len(b.Checks) > 0 }

type Prepared struct {
	Binding   Binding
	Directory string
	before    map[string][32]byte
	existed   map[string]bool
}

type Evidence struct {
	Directory    string        `json:"directory"`
	ChangedFiles []string      `json:"changed_files"`
	Checks       []CheckResult `json:"checks"`
	Error        string        `json:"error,omitempty"`
	ErrorKind    string        `json:"error_kind,omitempty"`
}
type CheckResult struct {
	Command []string `json:"command"`
	Output  string   `json:"output"`
	Passed  bool     `json:"passed"`
}
