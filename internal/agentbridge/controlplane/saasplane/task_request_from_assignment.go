package saasplane

import (
	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	providercatalog "github.com/teamswyg/riido-contracts/provider/catalog"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func taskRequestFromAssignment(assignment assignmentcontract.Assignment) *bridge.TaskRequest {
	prompt, systemPrompt, telemetryPlacement, instructionPlacement := agentbridge.ApplyRuntimeInstructionContract(assignment.RuntimeProvider, assignment.Prompt, "", assignment.AgentInstruction)
	return &bridge.TaskRequest{
		ID:                       assignmentExecutionID(assignment),
		Provider:                 bridge.Provider(assignment.RuntimeProvider),
		Model:                    providercatalog.ModelOverride(assignment.RuntimeProvider, assignment.ModelID),
		Prompt:                   prompt,
		SystemPrompt:             systemPrompt,
		AllowExperimentalRuntime: assignment.AllowExperimentalRuntime,
		ResumeSessionID:          assignmentResumeSessionID(assignment),
		Worktree:                 cloneAssignmentWorktree(assignment.Worktree),
		Metadata:                 taskRequestMetadata(assignment, telemetryPlacement, instructionPlacement),
	}
}

func taskRequestMetadata(assignment assignmentcontract.Assignment, telemetryPlacement, instructionPlacement string) map[string]string {
	metadata := assignmentBaseMetadata(assignment)
	metadata[agentbridge.MetadataTelemetryContract] = telemetryPlacement
	if instructionPlacement != "" {
		metadata[agentbridge.MetadataAgentInstruction] = instructionPlacement
	}
	return metadata
}

func assignmentBaseMetadata(assignment assignmentcontract.Assignment) map[string]string {
	executionID := assignmentExecutionID(assignment)
	return map[string]string{
		MetadataAssignmentID:        assignment.ID,
		MetadataAgentID:             assignment.AgentID,
		MetadataComponentID:         assignment.ComponentID,
		MetadataLeaseToken:          assignment.LeaseToken,
		MetadataModelID:             assignment.ModelID,
		MetadataRuntimeProvider:     assignment.RuntimeProvider,
		controlplane.MetadataTaskID: assignment.TaskID,
		"workspace_id":              assignmentWorkspaceID(assignment),
		"run_id":                    executionID,
	}
}
