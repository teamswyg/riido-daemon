package detectutil

import "time"

const versionProbeTimeout = 10 * time.Second

// loginShellPATHTimeout bounds the one-shot login-shell PATH probe so a slow
// or misbehaving shell profile can never hang Detect.
const loginShellPATHTimeout = 3 * time.Second

const (
	loginPATHMarkerStart = "__RIIDO_PATH_START__"
	loginPATHMarkerEnd   = "__RIIDO_PATH_END__"
)
