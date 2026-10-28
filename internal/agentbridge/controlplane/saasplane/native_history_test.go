package saasplane

import (
	"encoding/json"
	"testing"

	"github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestNativeHistoryMappingPreservesSourceAndLease(t *testing.T) {
	m := agentbridge.NativeHistoryMessage{SessionID: "native", MessageID: "turn:item", Role: "agent", Body: "answer", Ordinal: 4}
	req, ok := eventRequestFromAgentEvent(assignment.Assignment{ID: "run", TaskID: "task", LeaseToken: "lease"}, agentbridge.Event{Kind: agentbridge.EventProgress, NativeHistory: &m})
	var decoded agentbridge.NativeHistoryMessage
	if !ok || req.EventType != assignment.EventRiidoLog || req.Metadata["lease_token"] != "lease" || json.Unmarshal([]byte(req.Metadata[agentbridge.NativeHistoryMetadataKey]), &decoded) != nil || decoded != m {
		t.Fatalf("bad native event: %+v", req)
	}
}
