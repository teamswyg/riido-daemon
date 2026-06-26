package main

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/logging"
)

type daemonApprovalResolverFunc func(context.Context, string, agentbridge.ToolRef) (agentbridge.ToolApprovalResolution, error)

func (f daemonApprovalResolverFunc) ResolveToolApproval(ctx context.Context, executionID string, tool agentbridge.ToolRef) (agentbridge.ToolApprovalResolution, error) {
	return f(ctx, executionID, tool)
}

type daemonApprovalAuthorizerFunc func(context.Context, string) (bool, error)

func (f daemonApprovalAuthorizerFunc) AuthorizeToolApproval(ctx context.Context, executionID string) (bool, error) {
	return f(ctx, executionID)
}

func TestDaemonToolApprovalRequestUsesInProcessResolver(t *testing.T) {
	server, client := net.Pipe()
	defer client.Close()
	seen := make(chan agentbridge.ToolRef, 1)
	authorizer := daemonApprovalAuthorizerFunc(func(_ context.Context, executionID string) (bool, error) {
		return executionID == "asn-1", nil
	})
	resolver := daemonApprovalResolverFunc(func(_ context.Context, executionID string, tool agentbridge.ToolRef) (agentbridge.ToolApprovalResolution, error) {
		if executionID != "asn-1" {
			t.Fatalf("executionID = %q", executionID)
		}
		seen <- tool
		return agentbridge.ToolApprovalResolution{Approved: true, Reason: "approved in web"}, nil
	})
	go handleDaemonConn(server, startFlags{}, daemonSettings{}, time.Now(), nil, resolver, authorizer, nil, logging.NewWriterLogger(io.Discard))

	req := daemonRequest{
		Method:       daemonMethodToolApproval,
		AssignmentID: "asn-1",
		Tool:         agentbridge.ToolRef{ID: "toolu-1", Name: "Write", Kind: "Write"},
	}
	if err := json.NewEncoder(client).Encode(req); err != nil {
		t.Fatalf("encode request: %v", err)
	}
	var resp daemonToolApprovalResponse
	if err := json.NewDecoder(client).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !resp.Approved || resp.Reason != "approved in web" {
		t.Fatalf("response = %+v", resp)
	}
	select {
	case tool := <-seen:
		if tool.ID != "toolu-1" || tool.Kind != "Write" {
			t.Fatalf("tool = %+v", tool)
		}
	case <-time.After(time.Second):
		t.Fatal("resolver was not called")
	}
}

func TestDaemonToolApprovalRequestRejectsInactiveAssignmentBeforeResolver(t *testing.T) {
	server, client := net.Pipe()
	defer client.Close()
	resolverCalled := make(chan struct{}, 1)
	authorizer := daemonApprovalAuthorizerFunc(func(_ context.Context, executionID string) (bool, error) {
		if executionID != "asn-forged" {
			t.Fatalf("executionID = %q", executionID)
		}
		return false, nil
	})
	resolver := daemonApprovalResolverFunc(func(_ context.Context, _ string, _ agentbridge.ToolRef) (agentbridge.ToolApprovalResolution, error) {
		resolverCalled <- struct{}{}
		return agentbridge.ToolApprovalResolution{Approved: true}, nil
	})
	go handleDaemonConn(server, startFlags{}, daemonSettings{}, time.Now(), nil, resolver, authorizer, nil, logging.NewWriterLogger(io.Discard))

	req := daemonRequest{
		Method:       daemonMethodToolApproval,
		AssignmentID: "asn-forged",
		Tool:         agentbridge.ToolRef{ID: "toolu-forged", Name: "Bash", Kind: "Bash"},
	}
	if err := json.NewEncoder(client).Encode(req); err != nil {
		t.Fatalf("encode request: %v", err)
	}
	var resp daemonToolApprovalResponse
	if err := json.NewDecoder(client).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Approved || resp.Detail != "assignment is not active" {
		t.Fatalf("response = %+v", resp)
	}
	select {
	case <-resolverCalled:
		t.Fatal("resolver was called for inactive assignment")
	case <-time.After(100 * time.Millisecond):
	}
}
