package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//go:embed qa_system.generated.json
var qaSystemFS embed.FS

type qaSystemSpec struct {
	SchemaVersion         string                `json:"schema_version"`
	ID                    string                `json:"id"`
	Scope                 string                `json:"scope"`
	Generated             string                `json:"generated"`
	Search                []qaSystemSearch      `json:"search"`
	ChangeDetection       []qaSystemChangeCheck `json:"change_detection"`
	ExecutionInventory    []qaSystemExecution   `json:"execution_inventory"`
	OptimizationQuestions []map[string]string   `json:"optimization_questions"`
	Precommit             map[string]string     `json:"precommit"`
}

type qaSystemSearch struct {
	ID        string   `json:"id"`
	Meaning   string   `json:"meaning"`
	Aliases   []string `json:"aliases"`
	Source    string   `json:"source"`
	Generated string   `json:"generated,omitempty"`
	Evidence  string   `json:"evidence"`
}

type qaSystemChangeCheck struct {
	ID       string   `json:"id"`
	File     string   `json:"file"`
	Contains []string `json:"contains"`
}

type qaSystemExecution struct {
	ID     string `json:"id"`
	Mode   string `json:"mode"`
	Owner  string `json:"owner"`
	Action string `json:"action"`
}

func qaSystemScenario() scenario {
	body, err := qaSystemFS.ReadFile("qa_system.generated.json")
	if err != nil {
		return failedQASystemScenario("read generated QA system DSL", err)
	}
	var spec qaSystemSpec
	if err := json.Unmarshal(body, &spec); err != nil {
		return failedQASystemScenario("parse generated QA system DSL", err)
	}
	root, err := repoRoot()
	if err != nil {
		return failedQASystemScenario("find repo root", err)
	}
	observed := map[string]any{
		"schema_version":         spec.SchemaVersion,
		"id":                     spec.ID,
		"scope":                  spec.Scope,
		"generated":              spec.Generated,
		"search":                 spec.Search,
		"execution_inventory":    spec.ExecutionInventory,
		"optimization_questions": spec.OptimizationQuestions,
		"precommit":              spec.Precommit,
	}
	generatedChecks, generatedOK, sourceOnly := qaSystemGeneratedChecks(root, spec.Search)
	changeChecks, changeOK := qaSystemChangeChecks(root, spec.ChangeDetection)
	executionCounts, executionOK := qaSystemExecutionCounts(spec.ExecutionInventory)
	observed["generated_checks"] = generatedChecks
	observed["change_detection"] = changeChecks
	observed["execution_counts"] = executionCounts
	observed["search_entries"] = len(spec.Search)
	observed["search_aliases"] = qaSystemAliasCount(spec.Search)
	observed["developer_query_surface"] = "contract lab query matches scenario.observed JSON, including search aliases and meanings"
	observed["inference_removed"] = map[string]any{
		"generated_artifacts":      generatedOK,
		"change_detection":         changeOK,
		"fully_systematized":       generatedOK && changeOK && len(sourceOnly) == 0,
		"all_execution_automated":  executionCounts["inference_required_count"] == 0,
		"system_automated_count":   executionCounts["system_automated_count"],
		"inference_required_count": executionCounts["inference_required_count"],
		"remaining_source_only":    sourceOnly,
		"remaining_source_count":   len(sourceOnly),
		"system_reports_problems":  true,
	}
	status := statusPassed
	if !generatedOK || !changeOK || !executionOK {
		status = statusFailed
	} else if len(sourceOnly) > 0 || executionCounts["inference_required_count"] != 0 {
		status = statusPartial
	}
	return scenario{ID: "local.qa.dsl_system_audit", Status: status, Observed: observed}
}

func failedQASystemScenario(summary string, err error) scenario {
	return scenario{
		ID:             "local.qa.dsl_system_audit",
		Status:         statusFailed,
		FailureSummary: summary + ": " + err.Error(),
	}
}

func qaSystemGeneratedChecks(root string, entries []qaSystemSearch) ([]map[string]any, bool, []string) {
	checks := []map[string]any{}
	sourceOnly := []string{}
	ok := true
	for _, entry := range entries {
		check := map[string]any{"id": entry.ID, "source": entry.Source, "generated": entry.Generated}
		if entry.Source == "" || entry.Generated == "" {
			check["status"] = "source-only"
			sourceOnly = append(sourceOnly, entry.ID)
			checks = append(checks, check)
			continue
		}
		fresh, err := sameCanonicalJSON(filepath.Join(root, entry.Source), filepath.Join(root, entry.Generated))
		switch {
		case err != nil:
			check["status"] = statusFailed
			check["error"] = err.Error()
			ok = false
		case fresh:
			check["status"] = statusPassed
		default:
			check["status"] = statusFailed
			check["error"] = "generated JSON drift"
			ok = false
		}
		checks = append(checks, check)
	}
	return checks, ok, sourceOnly
}

func qaSystemChangeChecks(root string, checks []qaSystemChangeCheck) ([]map[string]any, bool) {
	out := []map[string]any{}
	ok := true
	for _, check := range checks {
		body, err := os.ReadFile(filepath.Join(root, check.File))
		row := map[string]any{"id": check.ID, "file": check.File, "contains": check.Contains}
		if err != nil {
			row["status"] = statusFailed
			row["error"] = err.Error()
			out = append(out, row)
			ok = false
			continue
		}
		missing := []string{}
		text := string(body)
		for _, needle := range check.Contains {
			if !strings.Contains(text, needle) {
				missing = append(missing, needle)
			}
		}
		row["missing"] = missing
		row["status"] = statusPassed
		if len(missing) > 0 {
			row["status"] = statusFailed
			ok = false
		}
		out = append(out, row)
	}
	return out, ok
}

func qaSystemAliasCount(entries []qaSystemSearch) int {
	total := 0
	for _, entry := range entries {
		total += len(entry.Aliases)
	}
	return total
}

func qaSystemExecutionCounts(entries []qaSystemExecution) (map[string]any, bool) {
	byMode := map[string]int{"system": 0, "inferred": 0}
	unknown := []string{}
	for _, entry := range entries {
		switch entry.Mode {
		case "system", "inferred":
			byMode[entry.Mode]++
		default:
			unknown = append(unknown, entry.ID+":"+entry.Mode)
		}
	}
	return map[string]any{
		"total":                    len(entries),
		"by_mode":                  byMode,
		"system_automated_count":   byMode["system"],
		"inference_required_count": byMode["inferred"],
		"inference_required_ids":   qaSystemExecutionIDs(entries, "inferred"),
		"unknown_modes":            unknown,
	}, len(entries) > 0 && len(unknown) == 0
}

func qaSystemExecutionIDs(entries []qaSystemExecution, mode string) []string {
	ids := []string{}
	for _, entry := range entries {
		if entry.Mode == mode {
			ids = append(ids, entry.ID)
		}
	}
	return ids
}

func sameCanonicalJSON(left, right string) (bool, error) {
	leftBody, err := os.ReadFile(left)
	if err != nil {
		return false, fmt.Errorf("read %s: %w", left, err)
	}
	rightBody, err := os.ReadFile(right)
	if err != nil {
		return false, fmt.Errorf("read %s: %w", right, err)
	}
	var leftValue any
	if err := json.Unmarshal(leftBody, &leftValue); err != nil {
		return false, fmt.Errorf("parse %s: %w", left, err)
	}
	var rightValue any
	if err := json.Unmarshal(rightBody, &rightValue); err != nil {
		return false, fmt.Errorf("parse %s: %w", right, err)
	}
	leftCanon, err := json.Marshal(leftValue)
	if err != nil {
		return false, err
	}
	rightCanon, err := json.Marshal(rightValue)
	if err != nil {
		return false, err
	}
	return bytes.Equal(leftCanon, rightCanon), nil
}

func repoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found from %s", dir)
		}
		dir = parent
	}
}
