package main

import (
	"bytes"
	"encoding/json"
	"net"
	"strings"
	"testing"
)

func TestClaudePermissionMCPDelegatesToDaemonSocket(t *testing.T) {
	socketPath := daemonSocketPath(t)
	ln, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	seen := make(chan daemonRequest, 1)
	go serveOneToolApprovalRequest(t, ln, seen)

	input := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`,
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"approval","arguments":{"tool_name":"Write","input":{"file_path":"/tmp/main.go","password":"super-secret-value"},"tool_use_id":"toolu-1"}}}`,
		"",
	}, "\n")
	var out bytes.Buffer
	err = serveClaudePermissionMCP(strings.NewReader(input), &out, claudePermissionMCPOptions{
		socket:       socketPath,
		assignmentID: "asn-1",
		taskID:       "task-1",
		runtimeID:    "rt-1",
	})
	if err != nil {
		t.Fatalf("serve MCP: %v", err)
	}
	assertMCPAllowed(t, out.String())
	req := <-seen
	if req.AssignmentID != "asn-1" || req.TaskID != "task-1" || req.RuntimeID != "rt-1" {
		t.Fatalf("daemon request identity = %+v", req)
	}
	if req.Tool.ID != "toolu-1" || req.Tool.Kind != "Write" {
		t.Fatalf("daemon request tool = %+v", req.Tool)
	}
	if got := req.Tool.Args["password"]; got != "[redacted]" {
		t.Fatalf("tool args must be redacted before daemon IPC, got %q", got)
	}
}

func TestClaudePermissionMCPSupportsLargeToolInput(t *testing.T) {
	socketPath := daemonSocketPath(t)
	ln, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	seen := make(chan daemonRequest, 1)
	go serveOneToolApprovalRequest(t, ln, seen)

	input := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"approval","arguments":{"tool_name":"Write","input":{"file_path":"/tmp/large.txt","content":"` + strings.Repeat("x", 96*1024) + `"},"tool_use_id":"toolu-large"}}}`,
		"",
	}, "\n")
	var out bytes.Buffer
	err = serveClaudePermissionMCP(strings.NewReader(input), &out, claudePermissionMCPOptions{
		socket:       socketPath,
		assignmentID: "asn-large",
		taskID:       "task-large",
		runtimeID:    "rt-large",
	})
	if err != nil {
		t.Fatalf("serve MCP: %v", err)
	}
	assertMCPAllowed(t, out.String())
	req := <-seen
	if req.Tool.ID != "toolu-large" || req.Tool.Kind != "Write" {
		t.Fatalf("daemon request tool = %+v", req.Tool)
	}
}

func serveOneToolApprovalRequest(t *testing.T, ln net.Listener, seen chan<- daemonRequest) {
	t.Helper()
	conn, err := ln.Accept()
	if err != nil {
		return
	}
	defer conn.Close()
	var req daemonRequest
	if err := json.NewDecoder(conn).Decode(&req); err != nil {
		t.Errorf("decode daemon request: %v", err)
		return
	}
	seen <- req
	_ = json.NewEncoder(conn).Encode(daemonToolApprovalResponse{Approved: true, Reason: "ok"})
}

func assertMCPAllowed(t *testing.T, output string) {
	t.Helper()
	if !strings.Contains(output, `\"behavior\":\"allow\"`) {
		t.Fatalf("MCP output missing allow decision:\n%s", output)
	}
	if strings.Contains(output, "RIIDO_DEVICE_SECRET") || strings.Contains(output, "RIIDO_SAAS_URL") {
		t.Fatalf("MCP output leaked daemon credential surface:\n%s", output)
	}
}
