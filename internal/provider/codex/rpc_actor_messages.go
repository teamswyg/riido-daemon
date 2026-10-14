package codex

type registerMsg struct {
	id    int64
	reply chan RPCResult
}

type resolveMsg struct {
	id     int64
	result map[string]any
	err    error
}
