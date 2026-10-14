package main

import (
	"context"
	"time"

	"github.com/teamswyg/riido-daemon/internal/mwsdbridge"
	"github.com/teamswyg/riido-daemon/internal/project"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

type mwsdSyncResult struct {
	OK          bool                           `json:"ok"`
	StatePath   string                         `json:"state_path"`
	TaskDBPath  string                         `json:"task_db_path"`
	Schema      string                         `json:"schema_version"`
	TaskDB      string                         `json:"task_db_schema_version"`
	ProjectNum  int                            `json:"project_count"`
	TaskNum     int                            `json:"task_count"`
	Transition  int                            `json:"transition_count"`
	Evidence    int                            `json:"evidence_count"`
	Next        string                         `json:"harness_next_direction"`
	DecisionLLM string                         `json:"recommended_decision_llm"`
	Provider    string                         `json:"recommended_provider"`
	Diagnostics []project.ProjectionDiagnostic `json:"diagnostics"`
}

func runMwsdSync(ctx context.Context, client mwsdbridge.Client, options mwsdOptions) error {
	projection, err := fetchMwsdProjection(ctx, client)
	if err != nil {
		return err
	}
	state := project.StateFromProjection(projection)
	if err := project.SaveState(options.statePath, state); err != nil {
		return err
	}
	db, err := taskdb.LoadTaskDBOrEmpty(options.taskDBPath)
	if err != nil {
		return err
	}
	db = project.SyncTaskDBFromState(db, state, time.Now())
	if err := taskdb.SaveTaskDB(options.taskDBPath, db); err != nil {
		return err
	}
	return printJSON(newMwsdSyncResult(options, state, db))
}

func newMwsdSyncResult(options mwsdOptions, state project.StateFile, db taskdb.TaskDB) mwsdSyncResult {
	return mwsdSyncResult{
		OK:          len(state.Diagnostics) == 0,
		StatePath:   options.statePath,
		TaskDBPath:  options.taskDBPath,
		Schema:      state.SchemaVersion,
		TaskDB:      db.SchemaVersion,
		ProjectNum:  len(state.Projects),
		TaskNum:     len(state.Tasks),
		Transition:  len(db.Transitions),
		Evidence:    len(db.Evidence),
		Next:        state.HarnessNextDirection,
		DecisionLLM: state.RecommendedDecisionLLM,
		Provider:    state.RecommendedProvider,
		Diagnostics: state.Diagnostics,
	}
}
