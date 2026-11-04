package codex

import (
	"regexp"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// A server identifier can appear in a copyable command. Never interpolate
// provider error text, endpoints or arbitrary shell arguments.
var commandSafeServer = regexp.MustCompile("^[a-zA-Z0-9_][a-zA-Z0-9_.-]{0,127}$")

func recoveryAdvice(name, reason, detail string) string {
	if !commandSafeServer.MatchString(name) {
		return ""
	}
	if reason == "reauthenticationRequired" {
		return name + " MCP에 다시 로그인해야 해요. 에이전트를 실행하는 기기의 터미널에서 `codex mcp login " + name + "`을 실행하고 브라우저 인증을 마친 뒤 이 요청을 다시 시도해 주세요."
	}
	detail = strings.ToLower(detail)
	for _, signal := range []string{"auth required", "401 unauthorized", "401 (unauthorized)", "authentication required", "invalid_token", "token expired"} {
		if strings.Contains(detail, signal) {
			return name + " MCP 인증을 갱신해 주세요. OAuth 연결이면 에이전트 기기에서 `codex mcp login " + name + "`을 실행하세요. API 키 연결이면 해당 연결의 키를 갱신한 뒤 이 요청을 다시 시도해 주세요."
		}
	}
	for _, signal := range []string{"connection refused", "econnrefused", "error sending request for url", "transport channel closed", "connection closed", "connection reset", "timed out handshaking", "os error 61", "os error 10061"} {
		if strings.Contains(detail, signal) {
			return name + " MCP에 연결할 수 없어요. 서버의 실행 상태와 에이전트 기기에 설정된 연결 주소를 확인한 뒤 이 요청을 다시 시도해 주세요."
		}
	}
	return ""
}

func recoveryReply(message string) []agentbridge.Event {
	return []agentbridge.Event{{Kind: agentbridge.EventResult, Result: agentbridge.Result{Status: agentbridge.ResultBlocked, RequiresUserAction: true, Output: message}}}
}

func (d *protocolDriver) otherThread(p map[string]any) bool {
	thread := stringField(p, "threadId")
	return thread != "" && d.threadID != "" && thread != d.threadID
}

// Codex rejects thread/start or thread/resume before a thread exists when
// its effective configuration marks the unavailable server as required.
func requiredMCPRecovery(message string) string {
	_, failure, ok := strings.Cut(message, "required MCP servers failed to initialize: ")
	if !ok {
		return ""
	}
	var advice []string
	for serverFailure := range strings.SplitSeq(failure, "; ") {
		name, detail, ok := strings.Cut(serverFailure, ": ")
		if ok {
			if message := recoveryAdvice(name, "", detail); message != "" {
				advice = append(advice, message)
			}
		}
	}
	return strings.Join(advice, "\n")
}

func (d *protocolDriver) onMCPStartup(p map[string]any) []agentbridge.Event {
	if d.otherThread(p) {
		return nil
	}
	name := stringField(p, "name")
	if stringField(p, "status") != "failed" {
		delete(d.mcpFailures, name)
		return codexStructuralLifecycleEvent(codexMethodMCPStartupStatusUpdated)
	}
	advice := recoveryAdvice(name, stringField(p, "failureReason"), stringField(p, "error"))
	if advice == "" {
		return codexStructuralLifecycleEvent(codexMethodMCPStartupStatusUpdated)
	}
	if d.mcpFailures[name] == advice {
		return nil
	}
	if d.mcpFailures == nil {
		d.mcpFailures = make(map[string]string)
	}
	d.mcpFailures[name] = advice
	// Optional MCP startup failures do not own this turn.
	return []agentbridge.Event{{Kind: agentbridge.EventProgress, Text: advice}}
}

func (d *protocolDriver) mcpToolRecovery(p map[string]any) []agentbridge.Event {
	item := mapField(p, "item")
	if d.otherThread(p) || stringField(item, "type") != "mcpToolCall" || stringField(item, "status") != "failed" {
		return nil
	}
	if advice := recoveryAdvice(stringField(item, "server"), "", stringField(mapField(item, "error"), "message")); advice != "" {
		return recoveryReply(advice)
	}
	return nil
}
