package codex

import (
	"encoding/json"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// Recognize native model-rejection diagnostics only. In particular, HTTP 400,
// authentication failure and transport errors are not safe model substitutes.
func codexFailureStatus(p map[string]any, message string) agentbridge.ResultStatus {
	for _, object := range []map[string]any{p, mapField(p, "error")} {
		switch stringField(object, "code") {
		case "model_not_found", "unsupported_model":
			return agentbridge.ResultBlocked
		}
	}
	if native := stringField(mapField(p, "error"), "message"); native != "" {
		message = native
	}
	message = strings.TrimSpace(message)
	// The app-server can wrap the API error as a status prefix and JSON body.
	// Decode that object rather than matching arbitrary prose around a quote.
	if start := strings.IndexByte(message, '{'); start >= 0 {
		var body map[string]any
		if json.Unmarshal([]byte(message[start:]), &body) == nil {
			message = stringField(mapField(body, "error"), "message")
		}
	}
	const prefix = "The '"
	const suffix = "' model is not supported when using Codex with a ChatGPT account."
	if strings.HasPrefix(message, prefix) && strings.HasSuffix(message, suffix) {
		model := strings.TrimSuffix(strings.TrimPrefix(message, prefix), suffix)
		if strings.TrimSpace(model) != "" && !strings.ContainsAny(model, "\r\n'") {
			return agentbridge.ResultBlocked
		}
	}
	return agentbridge.ResultFailed
}
