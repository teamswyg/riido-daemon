package agentbridge

import "github.com/teamswyg/riido-contracts/metadatakeys"

const (
	ProgressMessageMetadataCode      = string(metadatakeys.ProgressMessageCode)
	ProgressMessageMetadataKey       = string(metadatakeys.ProgressMessageKey)
	ProgressMessageMetadataArgPrefix = string(metadatakeys.ProgressMessageArgPrefix)
)

// ProgressCode is the canonical numeric progress-message code. The underlying
// integer is preserved for telemetry metadata and DynamoDB compatibility, while
// daemon code branches on the named constants below.
type ProgressCode int

const (
	ProgressCodeUnknown ProgressCode = 0

	ProgressCodeAgentThinking                 ProgressCode = 1001
	ProgressCodeAssignmentQueuedAgentBusy     ProgressCode = 1002
	ProgressCodeAssignmentStoppedAgentDeleted ProgressCode = 1003
	ProgressCodeAssignmentStoppedByUser       ProgressCode = 1004

	ProgressCodeToolCollecting               ProgressCode = 1101
	ProgressCodeToolCollectionCompletedCount ProgressCode = 1102
	ProgressCodeToolRunning                  ProgressCode = 1103
	ProgressCodeToolCompleted                ProgressCode = 1104

	ProgressCodeAssignmentStarted   ProgressCode = 1201
	ProgressCodeAssignmentCompleted ProgressCode = 1202
	ProgressCodeAssignmentFailed    ProgressCode = 1203
)

type telemetryProgressPayload struct {
	Code    ProgressCode   `json:"code"`
	Key     string         `json:"key,omitempty"`
	Args    map[string]any `json:"args,omitempty"`
	Message string         `json:"message,omitempty"`
}
