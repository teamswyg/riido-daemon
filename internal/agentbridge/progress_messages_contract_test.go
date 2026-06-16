package agentbridge

import (
	"testing"

	"github.com/teamswyg/riido-contracts/progressmessage"
)

func TestProgressCodesExistInContractsCatalog(t *testing.T) {
	catalog, err := progressmessage.Catalog()
	if err != nil {
		t.Fatal(err)
	}
	codes := map[ProgressCode]string{}
	for _, message := range catalog.Messages {
		codes[ProgressCode(message.Code)] = message.Key
	}
	for _, code := range []ProgressCode{
		ProgressCodeAgentThinking,
		ProgressCodeAssignmentQueuedAgentBusy,
		ProgressCodeAssignmentStoppedAgentDeleted,
		ProgressCodeAssignmentStoppedByUser,
		ProgressCodeToolCollecting,
		ProgressCodeToolCollectionCompletedCount,
		ProgressCodeToolRunning,
		ProgressCodeToolCompleted,
		ProgressCodeAssignmentStarted,
		ProgressCodeAssignmentCompleted,
		ProgressCodeAssignmentFailed,
	} {
		if codes[code] == "" {
			t.Fatalf("progress code %d is missing from contracts catalog", code)
		}
	}
}
