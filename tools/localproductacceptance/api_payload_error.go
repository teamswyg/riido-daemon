package main

func observePayloadError(out *scenario, payload map[string]any) {
	if summary := payloadErrorSummary(payload); summary != "" {
		out.Observed["error"] = summary
	}
}

func payloadErrorSummary(payload map[string]any) string {
	for _, key := range []string{"error", "message", "failure_summary"} {
		if text := stringValue(payload[key]); text != "" {
			return text
		}
	}
	return ""
}
