package taskvalidation

import (
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func TestProviderForTaskFallbacksAndAvailability(t *testing.T) {
	db := validationTaskDB(task.StateValidating)
	got, err := ProviderForTask(db, "task:test", " ")
	if err != nil || got != "codex" {
		t.Fatalf("fallback provider=%q err=%v", got, err)
	}
	got, err = ProviderForTask(db, "task:test", " codex ")
	if err != nil || got != "codex" {
		t.Fatalf("requested provider=%q err=%v", got, err)
	}
	db.ProviderCandidates[0].Available = false
	if _, err := ProviderForTask(db, "task:test", "codex"); err == nil {
		t.Fatal("unavailable provider should fail")
	}
	db.ProviderCandidates = nil
	db.Tasks[0].RecommendedProvider = ""
	db.RecommendedProvider = ""
	if _, err := ProviderForTask(db, "task:test", ""); err == nil {
		t.Fatal("missing provider should fail")
	}
	if _, err := ProviderForTask(db, "missing", "codex"); err == nil {
		t.Fatal("missing task should fail")
	}
}

func TestDecisionLLMCommandIDAndContextRules(t *testing.T) {
	db := validationTaskDB(task.StateValidating)
	if err := ValidateDecisionLLMForTask(db, "task:test", ""); err != nil {
		t.Fatal(err)
	}
	err := ValidateDecisionLLMForTask(db, "task:test", "other")
	if err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("expected mismatch, got %v", err)
	}
	if err := ValidateDecisionLLMForTask(db, "missing", "codex"); err == nil {
		t.Fatal("missing task should fail")
	}
	now := time.Date(2026, 5, 20, 8, 0, 0, 123, time.FixedZone("KST", 9*3600))
	id := CommandID("task:test", now)
	if !strings.Contains(id, "20260519T230000.000000123Z") {
		t.Fatalf("command id=%q", id)
	}
	ctx, normalized := normalizeRunContext(nil, time.Time{})
	if ctx == nil || normalized.IsZero() {
		t.Fatalf("context/time not normalized: %v %v", ctx, normalized)
	}
}

func TestValidationFailureTransitionAndRequestValidation(t *testing.T) {
	state, event := validationTransitionForResult("failed")
	if state != task.StateFailed || event != ir.EventValidationFailed {
		t.Fatalf("failure transition=%s %s", state, event)
	}
	if !providerAvailable(nil, "any") {
		t.Fatal("empty candidate list should allow provider")
	}
	if err := (Request{TaskID: "task:test", Command: "ok", ApprovalID: "approval", Timeout: -1}).validate(); err == nil {
		t.Fatal("negative timeout should fail")
	}
	if _, ok := FindTask(taskdb.EmptyTaskDB(), "missing"); ok {
		t.Fatal("empty db should not find task")
	}
}
