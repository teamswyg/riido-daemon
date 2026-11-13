package codex

import (
	"context"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestOwnedMCPToolApprovalSurfacesAsHumanApproval(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		command agentbridge.CommandKind
		action  string
	}{
		{agentbridge.CommandApproveTool, "accept"},
		{agentbridge.CommandRejectTool, "decline"},
	} {
		t.Run(tc.action, func(t *testing.T) {
			d, io := ownedTurnDriver(t)
			request := makeServerRequest(42, codexMCPApprovalMethod, mcpApprovalParams())
			events, _, err := d.OnRaw(ctx, request, io)
			if err != nil || len(events) != 1 || events[0].Kind != agentbridge.EventToolApprovalNeeded {
				t.Fatalf("expected human approval, got %+v, %v", events, err)
			}
			tool := events[0].Tool
			if !tool.RequireHuman || tool.ProviderRequestID != "42" || tool.Name != "Allow Riido update task body?" || len(tool.Args) != 0 {
				t.Fatalf("unexpected approval: %+v", tool)
			}
			_, commands := agentbridge.Reduce(agentbridge.State{Tools: map[string]agentbridge.ToolRef{}}, events[0], func(agentbridge.ToolRef) bool {
				t.Fatal("human confirmation reached automatic policy")
				return true
			})
			if len(commands) != 0 {
				t.Fatalf("issued a command before a decision: %+v", commands)
			}
			duplicates, _, err := d.OnRaw(ctx, request, io)
			if err != nil || len(duplicates) != 0 {
				t.Fatalf("duplicate request: %+v, %v", duplicates, err)
			}
			cmd := agentbridge.Command{Kind: tc.command, ToolKind: tool.Kind, ToolID: tool.ID, ProviderRequestID: tool.ProviderRequestID}
			frame, err := d.BuildProviderInput(cmd)
			assertMCPAction(t, frame, tc.action, err)
			if _, err := d.BuildProviderInput(cmd); err == nil {
				t.Fatal("confirmation was answered twice")
			}
		})
	}
}

func TestUnsupportedMCPRequestsDoNotReachHumanApproval(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(map[string]any)
	}{
		{"other turn", func(p map[string]any) { p["turnId"] = "child-turn" }},
		{"other thread", func(p map[string]any) { p["threadId"] = "child" }},
		{"url", func(p map[string]any) { p["mode"] = "url" }},
		{"input form", func(p map[string]any) {
			p["requestedSchema"] = map[string]any{"type": "object", "properties": map[string]any{"name": map[string]any{"type": "string"}}}
		}},
		{"other kind", func(p map[string]any) { p["_meta"] = map[string]any{"codex_approval_kind": "browser_auth"} }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d, io := ownedTurnDriver(t)
			discardHandshakeWrites(io)
			params := mcpApprovalParams()
			tc.mutate(params)
			events, _, err := d.OnRaw(context.Background(), makeServerRequest(42, codexMCPApprovalMethod, params), io)
			if err != nil || len(events) != 1 || events[0].Kind != agentbridge.EventWarning || len(d.mcpApprovals) != 0 {
				t.Fatalf("expected unsupported request warning, got %+v, %v", events, err)
			}
			assertCancelWrite(t, io)
		})
	}
}

func TestOwnedMCPApprovalCancellation(t *testing.T) {
	d, io := ownedTurnDriver(t)
	discardHandshakeWrites(io)
	ctx := context.Background()
	if _, _, err := d.OnRaw(ctx, makeServerRequest(42, codexMCPApprovalMethod, mcpApprovalParams()), io); err != nil {
		t.Fatal(err)
	}
	if _, err := d.OnTerminate(ctx, io); err != nil {
		t.Fatal(err)
	}
	assertCancelWrite(t, io)
}

func discardHandshakeWrites(io *recordingIO) {
	for len(io.writes) > 0 {
		<-io.writes
	}
}

func assertCancelWrite(t *testing.T, io *recordingIO) {
	t.Helper()
	select {
	case frame := <-io.writes:
		assertMCPAction(t, frame, "cancel", nil)
	default:
		t.Fatal("expected cancel response")
	}
}

func assertMCPAction(t *testing.T, frame []byte, action string, err error) {
	t.Helper()
	want := `{"id":42,"jsonrpc":"2.0","result":{"action":"` + action + `"}}`
	if err != nil || strings.TrimSpace(string(frame)) != want {
		t.Fatalf("decision frame: %s, %v", frame, err)
	}
}

func mcpApprovalParams() map[string]any {
	return map[string]any{
		"threadId": "parent", "turnId": "parent-turn", "serverName": "codex_apps",
		"mode": "form", "message": "Allow Riido update task body?",
		"requestedSchema": map[string]any{"type": "object", "properties": map[string]any{}},
		"_meta": map[string]any{
			"codex_approval_kind": "mcp_tool_call",
			"tool_title":          "Update document body",
			"tool_params":         map[string]any{"content": "private draft"},
		},
	}
}
