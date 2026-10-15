package workdir

// NativeSessions is an optional durable metadata port. Native history stays in
// the provider's original working directory; the server chooses the resume ID.
type NativeSessions interface {
	LookupNativeSession(provider, sessionID string) (NativeSession, bool, error)
	PinNativeSession(NativeSession) error
}

// NativeSession contains no prompt, output, environment or credentials.
type NativeSession struct {
	Schema      int    `json:"schema"`
	Provider    string `json:"provider"`
	SessionID   string `json:"session_id"`
	WorkspaceID string `json:"workspace_id"`
	TaskID      string `json:"task_id"`
	AgentID     string `json:"agent_id"`
	Fingerprint string `json:"runtime_fingerprint"`
	GrantHash   string `json:"grant_hash"`
	OriginRunID string `json:"origin_run_id"`
	Workdir     string `json:"-"`
}
