package main

import (
	"errors"
	"testing"
)

func TestFigmaFailureScenariosDescribeRepairClass(t *testing.T) {
	rows := failedFigmaIntentScenarios("figma.json", errors.New("missing"))
	if len(rows) != 4 {
		t.Fatalf("rows=%d", len(rows))
	}
	for _, row := range rows {
		if row.Status != statusFailed || row.Repair.Class != "figma_intent_manifest_unavailable" {
			t.Fatalf("unexpected row: %#v", row)
		}
	}
	if rows[0].Endpoint != "figma.json" || rows[0].FailureSummary != "missing" {
		t.Fatalf("missing path or failure summary: %#v", rows[0])
	}
}

func TestFigmaRepairScenariosKeepObservedMatches(t *testing.T) {
	matches := []figmaIntentEntry{{NodeID: "1:2", Name: "runtime"}}
	missing := figmaIntentMissingScenario("figma.runtime.settings", matches)
	required := figmaGoldenRequiredScenario("figma.runtime.settings", matches, errors.New("no golden"))
	stale := figmaGoldenStaleScenario("figma.runtime.settings", matches, errors.New("drift"))
	if missing.Status != statusFailed || missing.Repair.Class != "figma_intent_missing" {
		t.Fatalf("unexpected missing scenario: %#v", missing)
	}
	if required.Status != statusSkipped || required.Repair.Class != "figma_visual_golden_required" {
		t.Fatalf("unexpected required scenario: %#v", required)
	}
	if stale.Status != statusFailed || stale.Repair.Class != "figma_visual_golden_stale" {
		t.Fatalf("unexpected stale scenario: %#v", stale)
	}
	if missing.Observed["matches_count"] != 1 {
		t.Fatalf("observed matches missing: %#v", missing.Observed)
	}
}

func TestFailedDSLScenariosPreserveContext(t *testing.T) {
	closed := failedClosedLoopMaturityScenario("parse", errors.New("bad json"))
	if closed.Status != statusFailed || closed.ID != "local.qa.closed_loop_maturity" {
		t.Fatalf("unexpected closed-loop failure: %#v", closed)
	}
	if closed.FailureSummary != "parse: bad json" {
		t.Fatalf("failure summary=%q", closed.FailureSummary)
	}
	qa := failedQASystemScenario("read", errors.New("missing"))
	if qa.Status != statusFailed || qa.ID != "local.qa.dsl_system_audit" {
		t.Fatalf("unexpected QA failure: %#v", qa)
	}
}

func TestTaskFixtureSummariesAndGeneratedSkip(t *testing.T) {
	teams := summarizeTaskFixtureTeams(map[string]any{"teams": []any{"a"}, "data": []any{"b"}, "items": []any{"c"}})
	if teams["teams_count"] != 3 {
		t.Fatalf("teams summary=%#v", teams)
	}
	created := summarizeTaskFixtureCreate(map[string]any{"component_id": "task-a"})
	if created["task_id_present"] != true {
		t.Fatalf("create summary=%#v", created)
	}
	cleanup := summarizeTaskFixtureCleanup(nil)
	if cleanup["cleanup_requested"] != true {
		t.Fatalf("cleanup summary=%#v", cleanup)
	}
	if !shouldSkipGeneratedTaskFlow(scenario{Status: statusFailed}, "generated") {
		t.Fatal("generated failed assignable scenario should skip mutation tail")
	}
}
