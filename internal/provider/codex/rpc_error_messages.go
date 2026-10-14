package codex

func codexRPCErrorMessage(payload map[string]any) string {
	if msg := errMessage(payload); msg != "" {
		return msg
	}
	return "codex rpc error"
}

func codexNotificationErrorMessage(p map[string]any) string {
	if msg := stringField(p, "message"); msg != "" {
		return msg
	}
	if msg := stringField(p, "detail"); msg != "" {
		return msg
	}
	if errText := stringField(p, "error"); errText != "" {
		return errText
	}
	if msg := stringField(mapField(p, "error"), "message"); msg != "" {
		return msg
	}
	return "codex runtime error"
}
