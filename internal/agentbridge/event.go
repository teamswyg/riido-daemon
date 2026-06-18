package agentbridge

import "time"

// Event is the run-scope envelope passed to the reducer. Fields not
// applicable to a given Kind are simply zero-valued.
type Event struct {
	Kind         EventKind
	At           time.Time
	SessionID    string
	Phase        RunState
	Text         string
	ProgressCode ProgressCode
	ProgressKey  string
	ProgressArgs map[string]string
	Tool         ToolRef
	Usage        Usage
	Result       Result
	ExitCode     int
	Err          string
}
