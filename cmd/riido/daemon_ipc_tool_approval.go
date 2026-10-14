package main

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/logging"
)

func writeToolApprovalResolution(
	ctx context.Context,
	conn net.Conn,
	req daemonRequest,
	resolver agentbridge.ToolApprovalResolver,
	authorizer agentbridge.ToolApprovalAuthorizer,
	log logging.Logger,
) {
	if resolver == nil {
		writeToolApprovalError(conn, "tool approval resolver unavailable")
		return
	}
	if authorizer == nil {
		writeToolApprovalError(conn, "tool approval authorizer unavailable")
		return
	}
	assignmentID := strings.TrimSpace(req.AssignmentID)
	if assignmentID == "" {
		writeToolApprovalError(conn, "assignment_id is required")
		return
	}
	ok, err := authorizer.AuthorizeToolApproval(ctx, assignmentID)
	if err != nil {
		log.Printf("tool approval authorizer error assignment=%s: %v", assignmentID, err)
		writeToolApprovalError(conn, err.Error())
		return
	}
	if !ok {
		log.Printf("tool approval rejected inactive assignment=%s", assignmentID)
		writeToolApprovalError(conn, "assignment is not active")
		return
	}
	tool := req.Tool
	if tool.Kind == "" {
		tool.Kind = tool.Name
	}
	started := time.Now()
	log.Printf("tool approval wait started assignment=%s tool=%s", assignmentID, tool.Kind)
	resolution, err := resolver.ResolveToolApproval(ctx, assignmentID, tool)
	if err != nil {
		log.Printf("tool approval resolver error assignment=%s tool=%s: %v", assignmentID, tool.Kind, err)
		writeToolApprovalError(conn, err.Error())
		return
	}
	log.Printf("tool approval wait resolved assignment=%s tool=%s approved=%t duration_ms=%d", assignmentID, tool.Kind, resolution.Approved, time.Since(started).Milliseconds())
	_ = writeDaemonJSON(conn, map[string]any{
		"schema_version": DaemonStatusSchemaVersion,
		"approved":       resolution.Approved,
		"reason":         resolution.Reason,
	})
}

func writeToolApprovalError(conn net.Conn, detail string) {
	_ = writeDaemonJSON(conn, map[string]any{
		"schema_version": DaemonStatusSchemaVersion,
		"approved":       false,
		"error":          "tool approval failed",
		"detail":         detail,
	})
}
