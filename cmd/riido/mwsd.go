package main

import (
	"context"
	"fmt"
	"time"

	"github.com/teamswyg/riido-daemon/internal/mwsdbridge"
	"github.com/teamswyg/riido-daemon/internal/project"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func runMwsd(args []string) error {
	if len(args) < 1 {
		printUsage()
		return fmt.Errorf("missing mwsd command")
	}
	if args[0] == "--help" || args[0] == "-h" {
		printUsage()
		return nil
	}
	command := mwsdCommand(args[0])
	var socketPath string
	var statePath string
	var taskDBPath string
	for index := 1; index < len(args); index++ {
		switch args[index] {
		case "--socket":
			index++
			if index >= len(args) {
				return fmt.Errorf("--socket requires a path")
			}
			socketPath = args[index]
		case "--state":
			index++
			if index >= len(args) {
				return fmt.Errorf("--state requires a path")
			}
			statePath = args[index]
		case "--task-db":
			index++
			if index >= len(args) {
				return fmt.Errorf("--task-db requires a path")
			}
			taskDBPath = args[index]
		case "--help", "-h":
			printUsage()
			return nil
		default:
			return fmt.Errorf("unknown argument: %s", args[index])
		}
	}
	if socketPath == "" {
		defaultSocketPath, err := mwsdbridge.DefaultSocketPath()
		if err != nil {
			return err
		}
		socketPath = defaultSocketPath
	}
	if command == mwsdCommandSync {
		if statePath == "" {
			defaultStatePath, err := project.DefaultStatePath()
			if err != nil {
				return err
			}
			statePath = defaultStatePath
		}
		if taskDBPath == "" {
			defaultTaskDBPath, err := taskdb.DefaultTaskDBPath()
			if err != nil {
				return err
			}
			taskDBPath = defaultTaskDBPath
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client := mwsdbridge.NewClient(socketPath)

	switch command {
	case mwsdCommandSnapshot:
		snapshot, err := client.FetchSnapshot(ctx)
		if err != nil {
			return err
		}
		return printJSON(snapshot)
	case mwsdCommandProjection:
		snapshot, err := client.FetchSnapshot(ctx)
		if err != nil {
			return err
		}
		projection, err := project.FromMwsdSnapshot(snapshot)
		if err != nil {
			return err
		}
		return printJSON(projection)
	case mwsdCommandSync:
		snapshot, err := client.FetchSnapshot(ctx)
		if err != nil {
			return err
		}
		projection, err := project.FromMwsdSnapshot(snapshot)
		if err != nil {
			return err
		}
		state := project.StateFromProjection(projection)
		if err := project.SaveState(statePath, state); err != nil {
			return err
		}
		db, err := taskdb.LoadTaskDBOrEmpty(taskDBPath)
		if err != nil {
			return err
		}
		db = project.SyncTaskDBFromState(db, state, time.Now())
		if err := taskdb.SaveTaskDB(taskDBPath, db); err != nil {
			return err
		}
		return printJSON(struct {
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
		}{
			OK:          len(state.Diagnostics) == 0,
			StatePath:   statePath,
			TaskDBPath:  taskDBPath,
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
		})
	case mwsdCommandOrchestration:
		var orchestration mwsdbridge.OrchestrationSnapshot
		if err := client.Request(ctx, string(mwsdbridge.MethodOrchestration), &orchestration); err != nil {
			return err
		}
		return printJSON(orchestration)
	case mwsdCommandProjects:
		var projects mwsdbridge.ProjectRegistry
		if err := client.Request(ctx, string(mwsdbridge.MethodProjects), &projects); err != nil {
			return err
		}
		return printJSON(projects)
	case mwsdCommandStatus:
		var status mwsdbridge.Status
		if err := client.Request(ctx, string(mwsdbridge.MethodStatus), &status); err != nil {
			return err
		}
		return printJSON(status)
	default:
		printUsage()
		return fmt.Errorf("unknown mwsd command: %s", command)
	}
}
