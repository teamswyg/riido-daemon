package mwsdbridge

type OrchestrationNextAction struct {
	Direction             string `json:"direction"`
	CommandSurface        string `json:"command_surface"`
	Reason                string `json:"reason"`
	RequiresHumanApproval bool   `json:"requires_human_approval"`
}

type OrchestrationWorkflow struct {
	Name        string   `json:"name"`
	TopDown     []string `json:"top_down"`
	BottomUp    []string `json:"bottom_up"`
	Balance     []string `json:"balance"`
	DecisionBy  []string `json:"decision_by"`
	DecisionLLM []string `json:"decision_llm"`
	Providers   []string `json:"providers"`
	LoopSteps   []string `json:"loop_steps"`
}

type OrchestrationRun struct {
	ID        string `json:"id"`
	Direction string `json:"direction"`
	Source    string `json:"source"`
	Provider  string `json:"provider"`
	Command   string `json:"command"`
	Result    string `json:"result"`
}

type ProjectRegistry struct {
	SchemaVersion   string              `json:"schema_version"`
	Root            string              `json:"root"`
	DomainPath      string              `json:"domain_path"`
	RepositoryCount int                 `json:"repository_count"`
	Repositories    []ProjectRepository `json:"repositories"`
	Diagnostics     []Diagnostic        `json:"diagnostics"`
}

type ProjectRepository struct {
	Name          string   `json:"name"`
	Owner         string   `json:"owner"`
	Visibility    string   `json:"visibility"`
	SSOTScope     string   `json:"ssot_scope"`
	LocalPath     string   `json:"local_path"`
	Remote        string   `json:"remote"`
	Role          string   `json:"role"`
	Consumes      []string `json:"consumes"`
	LocalPresent  bool     `json:"local_present"`
	GitPresent    bool     `json:"git_present"`
	RemoteMatches bool     `json:"remote_matches"`
}

type Diagnostic struct {
	Severity string `json:"severity"`
	Code     string `json:"code"`
	Message  string `json:"message"`
}
