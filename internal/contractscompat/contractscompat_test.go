package contractscompat

import (
	"testing"

	"github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/provider/capability"
	"github.com/teamswyg/riido-contracts/task"
)

func TestContractsBaseline(t *testing.T) {
	if assignment.SchemaVersion != "riido-ai-server.v1" {
		t.Fatalf("assignment SchemaVersion = %q", assignment.SchemaVersion)
	}
	if !assignment.CanTransition(assignment.AssignmentQueued, assignment.AssignmentLeased) {
		t.Fatal("Queued -> Leased assignment transition must remain legal")
	}
	assignmentFSM := assignment.GeneratedAssignmentFSM()
	if !assignmentFSM.CanTransition(assignment.AssignmentStateCodeQueued, assignment.AssignmentStateCodeLeased) {
		t.Fatal("queued -> leased assignment FSM transition must remain legal")
	}
	if assignment.GeneratedAssignmentFSMServiceProvider().AssignmentFSM().Name() != "assignment" {
		t.Fatal("assignment FSM service provider must return the generated assignment FSM")
	}
	if !ir.EventTaskQueued.IsTransition() {
		t.Fatal("TaskQueued must remain a transition event")
	}
	if task.FSMSchemaVersion != 1 {
		t.Fatalf("FSMSchemaVersion = %d", task.FSMSchemaVersion)
	}
	if !task.ValidateTransition(task.StateCreated, task.StateQueued, ir.EventTaskQueued) {
		t.Fatal("Created -> Queued must remain legal")
	}
	taskFSM := task.GeneratedTaskFSM()
	if !taskFSM.CanTransition(task.TaskStateCodeRunning, task.TaskStateCodeValidating, ir.EventTypeCodeRunReportedDone) {
		t.Fatal("Running -> Validating task FSM transition must remain legal")
	}
	if len(taskFSM.NextStates(task.TaskStateCodeHumanReview, ir.EventTypeCodeHumanRejected)) != 2 {
		t.Fatal("HumanRejected must keep the documented rework/cancel split")
	}
	if task.GeneratedTaskFSMServiceProvider().TaskFSM().Name() != "task" {
		t.Fatal("task FSM service provider must return the generated task FSM")
	}

	fingerprint, err := capability.ComputeCapabilityFingerprint(capability.CapabilityFingerprintInput{
		ProviderKind:          capability.ProviderKind("codex"),
		ProtocolKind:          capability.ProtocolCodexAppServer,
		ProviderVersion:       "codex test",
		DetectedFingerprint:   capability.DetectedFingerprint("detected"),
		AdapterID:             "codex",
		AdapterVersion:        "riido-agentbridge-adapter.v1",
		ProtocolVersion:       "v1",
		DefaultSandboxMode:    "workspace-write",
		DefaultApprovalPolicy: "on-request",
		PolicyBundleVersion:   "policy-bundle.test.v1",
		ImportantSurfaceFlags: map[string]any{"structuredEventStream": true},
	})
	if err != nil {
		t.Fatalf("ComputeCapabilityFingerprint: %v", err)
	}
	if fingerprint == "" {
		t.Fatal("CapabilityFingerprint is empty")
	}
}
