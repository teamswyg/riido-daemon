// Package mwsdbridge is Riido's anti-corruption layer for the
// macmini-workspace daemon.
//
// mwsd remains the local control-plane SSOT for document graph, domain DSL,
// harness history, and private repo registry. Riido consumes those contracts
// through this package instead of parsing macmini-workspace files directly.
package mwsdbridge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"
)

const (
	GraphSchemaVersion         = "mws-doc-graph.v1"
	DomainSchemaVersion        = "mws-cl-domain.v1"
	HarnessSchemaVersion       = "mws-harness-run.v1"
	ProjectsSchemaVersion      = "mws-project-registry.v1"
	OrchestrationSchemaVersion = "mws-orchestration-snapshot.v1"
)

// Client reads the local mwsd Unix socket.
type Client struct {
	SocketPath string
	Timeout    time.Duration
}

// NewClient returns a Client with a conservative local timeout.
func NewClient(socketPath string) Client {
	return Client{
		SocketPath: socketPath,
		Timeout:    3 * time.Second,
	}
}

// DefaultSocketPath returns the launchd-backed mwsd socket path.
func DefaultSocketPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "Application Support", "macmini-workspace", "mwsd.sock"), nil
}

// FetchSnapshot reads every mwsd contract Riido needs for its first workspace
// state projection.
func (c Client) FetchSnapshot(ctx context.Context) (Snapshot, error) {
	var snapshot Snapshot
	if err := c.Request(ctx, "status", &snapshot.Status); err != nil {
		return snapshot, err
	}
	if err := c.Request(ctx, "graph", &snapshot.Graph); err != nil {
		return snapshot, err
	}
	if err := c.Request(ctx, "domain", &snapshot.Domain); err != nil {
		return snapshot, err
	}
	if err := c.Request(ctx, "harness", &snapshot.Harness); err != nil {
		return snapshot, err
	}
	if err := c.Request(ctx, "orchestration", &snapshot.Orchestration); err != nil {
		return snapshot, err
	}
	if err := c.Request(ctx, "projects", &snapshot.Projects); err != nil {
		return snapshot, err
	}
	return snapshot, snapshot.Validate()
}

// Request sends one mwsd method request and decodes the response data.
func (c Client) Request(ctx context.Context, method string, out any) error {
	if c.SocketPath == "" {
		return errors.New("mwsd socket path is empty")
	}
	timeout := c.Timeout
	if timeout == 0 {
		timeout = 3 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	conn, err := (&net.Dialer{}).DialContext(ctx, "unix", c.SocketPath)
	if err != nil {
		return fmt.Errorf("connect mwsd socket: %w", err)
	}
	defer conn.Close()

	body, err := json.Marshal(request{Method: method})
	if err != nil {
		return fmt.Errorf("encode mwsd request: %w", err)
	}
	if _, err := conn.Write(body); err != nil {
		return fmt.Errorf("write mwsd request: %w", err)
	}
	if unix, ok := conn.(*net.UnixConn); ok {
		if err := unix.CloseWrite(); err != nil {
			return fmt.Errorf("close mwsd request stream: %w", err)
		}
	}

	responseBody, err := io.ReadAll(conn)
	if err != nil {
		return fmt.Errorf("read mwsd response: %w", err)
	}
	var env responseEnvelope
	if err := json.Unmarshal(responseBody, &env); err != nil {
		return fmt.Errorf("decode mwsd response: %w", err)
	}
	if !env.OK {
		if env.Error != "" {
			return fmt.Errorf("mwsd %s failed: %s", method, env.Error)
		}
		return fmt.Errorf("mwsd %s failed", method)
	}
	if env.Method != method {
		return fmt.Errorf("mwsd method mismatch: requested %s got %s", method, env.Method)
	}
	if err := json.Unmarshal(env.Data, out); err != nil {
		return fmt.Errorf("decode mwsd %s data: %w", method, err)
	}
	return nil
}

type request struct {
	Method string `json:"method"`
}

type responseEnvelope struct {
	OK     bool            `json:"ok"`
	Method string          `json:"method"`
	Data   json.RawMessage `json:"data"`
	Error  string          `json:"error"`
}

// Snapshot is Riido's initial projection from macmini-workspace.
type Snapshot struct {
	Status        Status                `json:"status"`
	Graph         GraphExport           `json:"graph"`
	Domain        DomainExport          `json:"domain"`
	Harness       HarnessIndex          `json:"harness"`
	Orchestration OrchestrationSnapshot `json:"orchestration"`
	Projects      ProjectRegistry       `json:"projects"`
}

// Validate checks the schema-level handshake between Riido and mwsd.
func (s Snapshot) Validate() error {
	checks := []struct {
		name string
		got  string
		want string
	}{
		{"graph", s.Graph.SchemaVersion, GraphSchemaVersion},
		{"domain", s.Domain.SchemaVersion, DomainSchemaVersion},
		{"harness", s.Harness.SchemaVersion, HarnessSchemaVersion},
		{"orchestration", s.Orchestration.SchemaVersion, OrchestrationSchemaVersion},
		{"projects", s.Projects.SchemaVersion, ProjectsSchemaVersion},
	}
	for _, check := range checks {
		if check.got != check.want {
			return fmt.Errorf("%s schema mismatch: got %q want %q", check.name, check.got, check.want)
		}
	}
	if s.Status.Root == "" {
		return errors.New("status root is empty")
	}
	if s.Graph.Root != "" && s.Graph.Root != s.Status.Root {
		return fmt.Errorf("graph root mismatch: %s != %s", s.Graph.Root, s.Status.Root)
	}
	if s.Status.OrchestrationSchemaVersion != "" && s.Status.OrchestrationSchemaVersion != OrchestrationSchemaVersion {
		return fmt.Errorf("status orchestration schema mismatch: got %q want %q", s.Status.OrchestrationSchemaVersion, OrchestrationSchemaVersion)
	}
	if s.Orchestration.Root != "" && s.Orchestration.Root != s.Status.Root {
		return fmt.Errorf("orchestration root mismatch: %s != %s", s.Orchestration.Root, s.Status.Root)
	}
	if s.Orchestration.DomainSchemaVersion != "" && s.Orchestration.DomainSchemaVersion != DomainSchemaVersion {
		return fmt.Errorf("orchestration domain schema mismatch: got %q want %q", s.Orchestration.DomainSchemaVersion, DomainSchemaVersion)
	}
	if s.Orchestration.HarnessSchemaVersion != "" && s.Orchestration.HarnessSchemaVersion != HarnessSchemaVersion {
		return fmt.Errorf("orchestration harness schema mismatch: got %q want %q", s.Orchestration.HarnessSchemaVersion, HarnessSchemaVersion)
	}
	if s.Orchestration.TopDownCount != s.Harness.TopDownCount {
		return fmt.Errorf("orchestration top-down count mismatch: %d != %d", s.Orchestration.TopDownCount, s.Harness.TopDownCount)
	}
	if s.Orchestration.BottomUpCount != s.Harness.BottomUpCount {
		return fmt.Errorf("orchestration bottom-up count mismatch: %d != %d", s.Orchestration.BottomUpCount, s.Harness.BottomUpCount)
	}
	if s.Orchestration.NextAction.Direction != "" && s.Orchestration.NextAction.Direction != s.Harness.NextDirection {
		return fmt.Errorf("orchestration next direction mismatch: %s != %s", s.Orchestration.NextAction.Direction, s.Harness.NextDirection)
	}
	if s.Projects.RepositoryCount != len(s.Projects.Repositories) {
		return fmt.Errorf("project registry count mismatch: %d != %d", s.Projects.RepositoryCount, len(s.Projects.Repositories))
	}
	return nil
}

type Status struct {
	Root                       string   `json:"root"`
	SocketPath                 string   `json:"socket_path"`
	GraphSchemaVersion         string   `json:"graph_schema_version"`
	DomainSchemaVersion        string   `json:"domain_schema_version"`
	HarnessSchemaVersion       string   `json:"harness_schema_version"`
	OrchestrationSchemaVersion string   `json:"orchestration_schema_version"`
	DocumentCount              int      `json:"document_count"`
	RepositoryCount            int      `json:"repository_count"`
	DomainName                 string   `json:"domain_name"`
	HarnessRunCount            int      `json:"harness_run_count"`
	HarnessNextDirection       string   `json:"harness_next_direction"`
	HarnessRecentDirections    []string `json:"harness_recent_directions"`
	SSOTConflictCount          int      `json:"ssot_conflict_count"`
	DomainDiagnosticCount      int      `json:"domain_diagnostic_count"`
	HarnessDiagnosticCount     int      `json:"harness_diagnostic_count"`
	DiagnosticCount            int      `json:"diagnostic_count"`
	ErrorCount                 int      `json:"error_count"`
	WarningCount               int      `json:"warning_count"`
	UnresolvedLinkCount        int      `json:"unresolved_link_count"`
}

type GraphExport struct {
	SchemaVersion string     `json:"schema_version"`
	Root          string     `json:"root"`
	Documents     []Document `json:"documents"`
	Stats         GraphStats `json:"stats"`
}

type Document struct {
	Path                string   `json:"path"`
	ID                  string   `json:"id"`
	Title               string   `json:"title"`
	Status              string   `json:"status"`
	Owner               string   `json:"owner"`
	Links               []string `json:"links"`
	Backlinks           []string `json:"backlinks"`
	MissingLinks        []string `json:"missing_links"`
	HasBacklinksSection bool     `json:"has_backlinks_section"`
}

type GraphStats struct {
	DocumentCount       int `json:"document_count"`
	NodeCount           int `json:"node_count"`
	EdgeCount           int `json:"edge_count"`
	DiagnosticCount     int `json:"diagnostic_count"`
	ErrorCount          int `json:"error_count"`
	WarningCount        int `json:"warning_count"`
	UnresolvedLinkCount int `json:"unresolved_link_count"`
}

type DomainExport struct {
	SchemaVersion string             `json:"schema_version"`
	Path          string             `json:"path"`
	Domain        string             `json:"domain"`
	Repositories  []DomainRepository `json:"repositories"`
	Diagnostics   []Diagnostic       `json:"diagnostics"`
}

type DomainRepository struct {
	Name       string   `json:"name"`
	Owner      string   `json:"owner"`
	Visibility string   `json:"visibility"`
	SSOTScope  string   `json:"ssot_scope"`
	LocalPath  string   `json:"local_path"`
	Remote     string   `json:"remote"`
	Role       string   `json:"role"`
	Consumes   []string `json:"consumes"`
}

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

type OrchestrationSnapshot struct {
	SchemaVersion          string                  `json:"schema_version"`
	Root                   string                  `json:"root"`
	DomainPath             string                  `json:"domain_path"`
	HarnessRunPath         string                  `json:"harness_run_path"`
	DomainSchemaVersion    string                  `json:"domain_schema_version"`
	HarnessSchemaVersion   string                  `json:"harness_schema_version"`
	Mode                   string                  `json:"mode"`
	DecisionGate           string                  `json:"decision_gate"`
	DecisionBy             []string                `json:"decision_by"`
	DecisionLLMs           []string                `json:"decision_llms"`
	ProviderCandidates     []ProviderCandidate     `json:"provider_candidates"`
	RecommendedProvider    string                  `json:"recommended_provider"`
	RecommendedDecisionLLM string                  `json:"recommended_decision_llm"`
	NextAction             OrchestrationNextAction `json:"next_action"`
	TopDownCount           int                     `json:"top_down_count"`
	BottomUpCount          int                     `json:"bottom_up_count"`
	LastDirection          string                  `json:"last_direction"`
	Balanced               bool                    `json:"balanced"`
	DirectionBias          bool                    `json:"direction_bias"`
	Workflows              []OrchestrationWorkflow `json:"workflows"`
	RecentRuns             []OrchestrationRun      `json:"recent_runs"`
	Diagnostics            []Diagnostic            `json:"diagnostics"`
}

type ProviderCandidate struct {
	ID               string `json:"id"`
	SourceWorkflow   string `json:"source_workflow"`
	Available        bool   `json:"available"`
	ApprovalRequired bool   `json:"approval_required"`
}

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
