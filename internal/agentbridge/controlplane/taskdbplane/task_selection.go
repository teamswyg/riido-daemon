package taskdbplane

import (
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/provider/capability"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/scheduling"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
	"github.com/teamswyg/riido-daemon/pkg/util/textutil"
)

func applyTransition(db taskdb.TaskDB, record taskdb.TaskRecord, to task.TaskState, event ir.EventType, reason, commandSuffix string, now time.Time) (taskdb.TaskDB, error) {
	approvalID := approvalIDForTask(db, record.ID)
	if requiresApproval(db, record) && approvalID == "" {
		return taskdb.TaskDB{}, planeErrorf(ErrTaskDBPlaneTaskState, "apply-transition", "task %s requires approval_id before %s", record.ID, event)
	}
	updated, _, _, err := taskdb.ApplyGuardedTaskTransition(db, taskdb.TaskTransitionInput{
		TaskID:  record.ID,
		ToState: to,
		Event:   event,
		Actor:   defaultActor,
		Source:  sourceName,
		Reason:  reason,
		Guard:   guardFor(db, record, commandSuffix, approvalID),
	}, now)
	if err != nil {
		return taskdb.TaskDB{}, err
	}
	return updated, nil
}

func guardFor(db taskdb.TaskDB, record taskdb.TaskRecord, suffix, approvalID string) taskdb.TaskMutationGuardInput {
	return taskdb.TaskMutationGuardInput{
		CommandID:   commandIDPrefix + record.ID + ":" + suffix,
		Provider:    providerFor(db, record),
		DecisionLLM: decisionLLMFor(db, record),
		ApprovalID:  approvalID,
	}
}

func findTask(db taskdb.TaskDB, taskID string) (taskdb.TaskRecord, bool) {
	for _, record := range db.Tasks {
		if record.ID == taskID {
			return record, true
		}
	}
	return taskdb.TaskRecord{}, false
}

func providerFor(db taskdb.TaskDB, record taskdb.TaskRecord) string {
	return textutil.FirstNonEmptyTrimmed(record.RecommendedProvider, db.RecommendedProvider)
}

func decisionLLMFor(db taskdb.TaskDB, record taskdb.TaskRecord) string {
	return textutil.FirstNonEmptyTrimmed(record.RecommendedDecisionLLM, db.RecommendedDecisionLLM)
}

func promptFor(record taskdb.TaskRecord) string {
	return textutil.FirstNonEmptyTrimmed(record.HarnessNextDirection, record.Title)
}

func requiresApproval(db taskdb.TaskDB, record taskdb.TaskRecord) bool {
	return record.RequiresHumanApproval || db.DecisionGate == "human-approval-required"
}

func approvalIDForTask(db taskdb.TaskDB, taskID string) string {
	for _, receipt := range slices.Backward(db.CommandReceipts) {
		if receipt.TaskID == taskID && strings.TrimSpace(receipt.ApprovalID) != "" {
			return strings.TrimSpace(receipt.ApprovalID)
		}
	}
	return ""
}

func providerAvailable(db taskdb.TaskDB, provider string) bool {
	if len(db.ProviderCandidates) == 0 {
		return true
	}
	for _, candidate := range db.ProviderCandidates {
		if candidate.ID == provider {
			return candidate.Available
		}
	}
	return false
}

func (p *Plane) runtimeSelectionForTask(provider, runtimeID string) (scheduling.RuntimeSelection, bool) {
	if len(p.runtimes) == 0 {
		return scheduling.RuntimeSelection{Runtime: scheduling.RuntimeCapability{
			RuntimeID: capability.RuntimeID(runtimeID),
			Provider:  capability.ProviderKind(provider),
			Available: true,
		}}, true
	}
	selection, ok := scheduling.SelectRuntime(scheduling.TaskRequirements{
		Provider: capability.ProviderKind(provider),
	}, p.runtimeCandidatesForProvider(provider))
	if !ok {
		return scheduling.RuntimeSelection{}, false
	}
	if string(selection.Runtime.RuntimeID) != runtimeID {
		return scheduling.RuntimeSelection{}, false
	}
	return selection, true
}

func (p *Plane) runtimeCandidatesForProvider(provider string) []scheduling.RuntimeCapability {
	ids := make([]string, 0, len(p.runtimes))
	for id := range p.runtimes {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	out := make([]scheduling.RuntimeCapability, 0, len(ids))
	for _, id := range ids {
		candidate, ok := runtimeCapabilityForProvider(p.runtimes[id], provider)
		if ok {
			out = append(out, candidate)
		}
	}
	return out
}

func runtimeCapabilityForProvider(rec controlplane.RegisteredRuntime, provider string) (scheduling.RuntimeCapability, bool) {
	provider = strings.TrimSpace(provider)
	if rec.RuntimeID == "" || provider == "" {
		return scheduling.RuntimeCapability{}, false
	}
	prefix := "provider." + provider + "."
	available, ok := rec.Capabilities[prefix+"available"]
	if !ok {
		return scheduling.RuntimeCapability{}, false
	}
	return scheduling.RuntimeCapability{
		RuntimeID:                 capability.RuntimeID(rec.RuntimeID),
		Provider:                  capability.ProviderKind(provider),
		CapabilityFingerprint:     capability.CapabilityFingerprint(rec.CapabilityAttributes[prefix+"capability_fingerprint"]),
		SlotLimit:                 rec.SlotLimit,
		SlotsInUse:                rec.SlotsInUse,
		Available:                 available,
		CompatibilityStatus:       capability.CompatibilityStatus(rec.CapabilityAttributes[prefix+"compatibility_status"]),
		RequiresExperimentalOptIn: rec.Capabilities[prefix+"requires_experimental_opt_in"],
		SupportsStreaming:         rec.Capabilities[prefix+"supports_streaming"],
		SupportsResume:            rec.Capabilities[prefix+"supports_resume"],
		SupportsSystem:            rec.Capabilities[prefix+"supports_system"],
		SupportsMaxTurns:          rec.Capabilities[prefix+"supports_max_turns"],
		SupportsMCP:               rec.Capabilities[prefix+"supports_mcp"],
		SupportsToolHooks:         rec.Capabilities[prefix+"supports_tool_hooks"],
		SupportsUsage:             rec.Capabilities[prefix+"supports_usage"],
		SupportsWorktree:          rec.Capabilities[prefix+"supports_worktree"],
	}, true
}

func timeoutCanOriginate(state task.TaskState) bool {
	switch state.Code() {
	case task.TaskStateCodeRunning, task.TaskStateCodeNeedsInput, task.TaskStateCodeBlocked, task.TaskStateCodeValidating, task.TaskStateCodeHumanReview:
		return true
	default:
		return false
	}
}

func resultReason(res agentbridge.Result, fallback string) string {
	return textutil.FirstNonEmptyTrimmed(res.Error, res.Output, fallback)
}
