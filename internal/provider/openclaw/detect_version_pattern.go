package openclaw

import "regexp"

// openClawVersionRE matches OpenClaw's 20XX calendar version at line start,
// optionally prefixed by "openclaw", "openclaw version", or "v".
var openClawVersionRE = regexp.MustCompile(`(?im)^\s*(?:openclaw(?:\s+version)?\s+|v)?(20\d{2})\.(\d{1,2})\.(\d{1,2})(?:\s|$|[^.\d])`)
