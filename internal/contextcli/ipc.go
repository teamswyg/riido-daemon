package contextcli

import (
	"encoding/json"
	"strings"
)

const (
	// IPCMethodContextRead is the daemon IPC method agent-mode reads use.
	IPCMethodContextRead = "context_read"
	// ContextIPCSchemaVersion is the request schema the CLI sends.
	ContextIPCSchemaVersion = "1"
)

// ContextReadRequest is sent to the daemon broker over local IPC only. The
// run capability lets the broker match the caller to an active run; it is
// never printed or logged by the CLI.
type ContextReadRequest struct {
	Method        string            `json:"method"`
	SchemaVersion string            `json:"context_schema_version"`
	Command       string            `json:"command"`
	Args          map[string]string `json:"args,omitempty"`
	RunCapability string            `json:"run_capability"`
	TaskID        string            `json:"task_id,omitempty"`
	WorkspaceID   string            `json:"workspace_id,omitempty"`
	RequestID     string            `json:"request_id"`
}

// NewContextReadRequest builds an agent-mode read request.
func NewContextReadRequest(env Environment, requestID, command string, args map[string]string) ContextReadRequest {
	return ContextReadRequest{
		Method:        IPCMethodContextRead,
		SchemaVersion: ContextIPCSchemaVersion,
		Command:       command,
		Args:          args,
		RunCapability: env.RunCapability(),
		TaskID:        env.TaskID,
		WorkspaceID:   env.WorkspaceID,
		RequestID:     requestID,
	}
}

type daemonReply struct {
	SchemaVersion string          `json:"schema_version"`
	Error         json.RawMessage `json:"error"`
	Method        string          `json:"method"`
}

// DecodeContextReadReply turns a daemon reply into an envelope. Daemons
// without the broker answer "unknown method"; that maps to
// CONTEXT_READ_UNSUPPORTED (exit 8) so agents fall back explicitly.
func DecodeContextReadReply(body []byte) (Envelope, *CLIError) {
	var reply daemonReply
	if err := json.Unmarshal(body, &reply); err != nil {
		return Envelope{}, Fail(CodeUnsupported, "daemon 응답을 해석할 수 없습니다", "riido doctor 로 daemon 버전을 확인하세요")
	}
	if isUnknownMethod(reply) {
		return Envelope{}, Fail(CodeContextReadUnsupported,
			"설치된 daemon이 실행 중 맥락 조회(context_read)를 아직 지원하지 않습니다",
			"Riido Desktop을 업데이트하세요. 그전까지는 프롬프트로 받은 맥락만 사용하세요")
	}
	if reply.SchemaVersion != EnvelopeSchemaVersion {
		return Envelope{}, Failf(CodeUnsupported, "지원하지 않는 daemon 응답 버전입니다: %q", reply.SchemaVersion)
	}
	var env Envelope
	if err := json.Unmarshal(body, &env); err != nil {
		return Envelope{}, Fail(CodeUnsupported, "daemon 응답 형식이 envelope v1과 다릅니다", "riido doctor 로 daemon 버전을 확인하세요")
	}
	if env.Sources == nil {
		env.Sources = []Source{}
	}
	if env.Warnings == nil {
		env.Warnings = []Warning{}
	}
	return env, nil
}

func isUnknownMethod(reply daemonReply) bool {
	var text string
	if err := json.Unmarshal(reply.Error, &text); err != nil {
		return false
	}
	return strings.EqualFold(text, "unknown method") && reply.Method == IPCMethodContextRead
}

// ReplyError converts an envelope error returned by the daemon into a
// CLIError so the CLI exits with the documented status.
func ReplyError(env Envelope) *CLIError {
	if env.Error == nil {
		return nil
	}
	return &CLIError{Code: env.Error.Code, Message: env.Error.Message, NextAction: env.Error.NextAction}
}
