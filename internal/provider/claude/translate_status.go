package claude

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func translateStderrRaw(raw agentbridge.RawEvent) []agentbridge.Event {
	return []agentbridge.Event{{
		Kind: agentbridge.EventLog,
		Text: string(raw.Bytes),
	}}
}

func translateMalformed(raw agentbridge.RawEvent) []agentbridge.Event {
	return []agentbridge.Event{{
		Kind: agentbridge.EventWarning,
		Text: "malformed claude stream-json line",
		Err:  string(raw.Bytes),
	}}
}

func translateLog(raw agentbridge.RawEvent) []agentbridge.Event {
	return []agentbridge.Event{{
		Kind: agentbridge.EventLog,
		Text: stringField(raw.Payload, "message"),
	}}
}

func translateError(raw agentbridge.RawEvent) []agentbridge.Event {
	return []agentbridge.Event{{
		Kind: agentbridge.EventError,
		Err:  stringField(raw.Payload, "message"),
	}}
}

func translateRateLimit(raw agentbridge.RawEvent) []agentbridge.Event {
	return []agentbridge.Event{{
		Kind: agentbridge.EventWarning,
		Text: "claude rate limited",
		Err:  claudeRateLimitDetail(raw.Payload),
	}}
}

func translateUnknown(raw agentbridge.RawEvent) []agentbridge.Event {
	return []agentbridge.Event{{
		Kind: agentbridge.EventLog,
		Text: "claude unknown event: " + raw.Type,
	}}
}
