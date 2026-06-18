package session

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func approvalAdapter(t *testing.T, toolKind string) *recordingAdapter {
	t.Helper()
	return &recordingAdapter{
		name:   "fake",
		parser: &recordingParser{},
		translateFn: func(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
			switch string(raw.Bytes) {
			case "ASK":
				return []agentbridge.Event{{
					Kind: agentbridge.EventToolApprovalNeeded,
					Tool: agentbridge.ToolRef{
						ID: "tool-1", Kind: toolKind, ProviderRequestID: "req-1",
					},
				}}, nil, nil
			case "DONE":
				return []agentbridge.Event{completedResultEvent("")}, nil, nil
			default:
				return nil, nil, nil
			}
		},
		inputFn: func(cmd agentbridge.Command) ([]byte, error) {
			assertApproveToolCommand(t, cmd)
			return []byte("approve:req-1\n"), nil
		},
	}
}
