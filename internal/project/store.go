package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/teamswyg/riido-daemon/pkg/util/fileutil"
)

const StateSchemaVersion = "riido-project-state.v1"

// StateFile is Riido's first local persisted project/task view.
//
// It is a deterministic projection file, not the final transactional DB. The
// next storage layer can migrate this shape into SQLite or an event store
// without re-reading macmini-workspace files directly.
type StateFile struct {
	SchemaVersion          string                 `json:"schema_version"`
	ProjectionVersion      string                 `json:"projection_version"`
	Root                   string                 `json:"root"`
	Domain                 string                 `json:"domain"`
	HarnessRunCount        int                    `json:"harness_run_count"`
	HarnessNextDirection   string                 `json:"harness_next_direction"`
	OrchestrationMode      string                 `json:"orchestration_mode"`
	DecisionGate           string                 `json:"decision_gate"`
	DecisionBy             []string               `json:"decision_by"`
	DecisionLLMs           []string               `json:"decision_llms"`
	ProviderCandidates     []ProviderCandidate    `json:"provider_candidates"`
	RecommendedProvider    string                 `json:"recommended_provider"`
	RecommendedDecisionLLM string                 `json:"recommended_decision_llm"`
	NextAction             NextAction             `json:"next_action"`
	Projects               []ProjectState         `json:"projects"`
	Tasks                  []TaskState            `json:"tasks"`
	Diagnostics            []ProjectionDiagnostic `json:"diagnostics"`
}

type ProjectState struct {
	ID            string           `json:"id"`
	Owner         string           `json:"owner"`
	Visibility    string           `json:"visibility"`
	SSOTScope     string           `json:"ssot_scope"`
	LocalPath     string           `json:"local_path"`
	Remote        string           `json:"remote"`
	Role          string           `json:"role"`
	Health        RepositoryHealth `json:"health"`
	LocalPresent  bool             `json:"local_present"`
	GitPresent    bool             `json:"git_present"`
	RemoteMatches bool             `json:"remote_matches"`
}

type TaskState struct {
	ID                     string `json:"id"`
	ProjectID              string `json:"project_id"`
	State                  string `json:"state"`
	SourceDocumentID       string `json:"source_document_id"`
	SourceDocumentPath     string `json:"source_document_path"`
	Title                  string `json:"title"`
	Owner                  string `json:"owner"`
	SourceStatus           string `json:"source_status"`
	RecommendedProvider    string `json:"recommended_provider"`
	RecommendedDecisionLLM string `json:"recommended_decision_llm"`
	RequiresHumanApproval  bool   `json:"requires_human_approval"`
	HarnessNextDirection   string `json:"harness_next_direction"`
}

func StateFromProjection(projection WorkspaceProjection) StateFile {
	state := StateFile{
		SchemaVersion:          StateSchemaVersion,
		ProjectionVersion:      projection.SchemaVersion,
		Root:                   projection.Root,
		Domain:                 projection.Domain,
		HarnessRunCount:        projection.HarnessRunCount,
		HarnessNextDirection:   projection.HarnessNextDirection,
		OrchestrationMode:      projection.OrchestrationMode,
		DecisionGate:           projection.DecisionGate,
		DecisionBy:             append([]string(nil), projection.DecisionBy...),
		DecisionLLMs:           append([]string(nil), projection.DecisionLLMs...),
		ProviderCandidates:     append([]ProviderCandidate(nil), projection.ProviderCandidates...),
		RecommendedProvider:    projection.RecommendedProvider,
		RecommendedDecisionLLM: projection.RecommendedDecisionLLM,
		NextAction:             projection.NextAction,
		Diagnostics:            append([]ProjectionDiagnostic(nil), projection.Diagnostics...),
	}
	for _, project := range projection.Projects {
		state.Projects = append(state.Projects, ProjectState{
			ID:            project.ID,
			Owner:         project.Owner,
			Visibility:    project.Visibility,
			SSOTScope:     project.SSOTScope,
			LocalPath:     project.LocalPath,
			Remote:        project.Remote,
			Role:          project.Role,
			Health:        project.Health,
			LocalPresent:  project.LocalPresent,
			GitPresent:    project.GitPresent,
			RemoteMatches: project.RemoteMatches,
		})
	}
	for _, link := range projection.DocumentTaskLinks {
		state.Tasks = append(state.Tasks, TaskState{
			ID:                     link.TaskID,
			ProjectID:              link.ProjectID,
			State:                  "Created",
			SourceDocumentID:       link.DocumentID,
			SourceDocumentPath:     link.DocumentPath,
			Title:                  link.Title,
			Owner:                  link.Owner,
			SourceStatus:           link.Status,
			RecommendedProvider:    link.RecommendedProvider,
			RecommendedDecisionLLM: link.RecommendedDecisionLLM,
			RequiresHumanApproval:  link.RequiresHumanApproval,
			HarnessNextDirection:   link.HarnessNextDirection,
		})
	}
	if state.Diagnostics == nil {
		state.Diagnostics = []ProjectionDiagnostic{}
	}
	if state.Projects == nil {
		state.Projects = []ProjectState{}
	}
	if state.Tasks == nil {
		state.Tasks = []TaskState{}
	}
	return state
}

func DefaultStatePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "Application Support", "riido", "workspace-state.json"), nil
}

func SaveState(path string, state StateFile) error {
	if path == "" {
		return fmt.Errorf("state path is empty")
	}
	if err := fileutil.WriteJSONAtomic(path, state); err != nil {
		return fmt.Errorf("save state file: %w", err)
	}
	return nil
}

func LoadState(path string) (StateFile, error) {
	var state StateFile
	data, err := os.ReadFile(path)
	if err != nil {
		return state, err
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return state, fmt.Errorf("decode state file: %w", err)
	}
	if state.SchemaVersion != StateSchemaVersion {
		return state, fmt.Errorf("state schema mismatch: got %q want %q", state.SchemaVersion, StateSchemaVersion)
	}
	return state, nil
}
