package openclaw

import (
	"context"
	"regexp"
	"strconv"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/detectutil"
)

const EnvOverride = "RIIDO_OPENCLAW_PATH"

// MinSupportedVersion is the lowest openclaw version this adapter
// promises to work with. Older versions are reported unavailable so the daemon
// can surface a capability gap instead of starting a run against an unknown CLI
// protocol.
//
// The version string is calendar-versioned (YYYY.MM.DD). The strict
// parser below only accepts year prefixes of `20XX` so Node-style
// semver (e.g. 20.10.0, 22.12.0) embedded in dependency errors is
// NEVER mistaken for an OpenClaw version.
const MinSupportedVersion = "2026.5.5"

// openClawVersionRE matches OpenClaw's calendar-style version anchored
// either at line start or after `openclaw `, `openclaw version `, or a
// bare `v`. The 4-digit year MUST start with `20`. Node semver such as
// `22.12.0` (2-digit year) is rejected by construction.
//
// Examples accepted:
//
//	2026.5.5
//	v2026.5.5
//	openclaw 2026.5.5
//	openclaw version 2026.05.05
//	OpenClaw 2026.12.31
//
// Examples rejected:
//
//	22.12.0
//	v20.10.0
//	requires Node >=22.12.0
//	Node.js v20.10.0
//	    at /path/22.12.0/file.js
var openClawVersionRE = regexp.MustCompile(`(?im)^\s*(?:openclaw(?:\s+version)?\s+|v)?(20\d{2})\.(\d{1,2})\.(\d{1,2})(?:\s|$|[^.\d])`)

// parseVersion extracts a date-style version tuple (year, month, day)
// from s. Returns (tuple, true) when s matches OpenClaw's version
// shape; ([3]int{}, false) otherwise. Node-style semver is rejected.
func parseVersion(s string) ([3]int, bool) {
	m := openClawVersionRE.FindStringSubmatch(s)
	if m == nil {
		return [3]int{}, false
	}
	year, err := strconv.Atoi(m[1])
	if err != nil || year < 2020 || year > 2099 {
		return [3]int{}, false
	}
	month, err := strconv.Atoi(m[2])
	if err != nil {
		return [3]int{}, false
	}
	day, err := strconv.Atoi(m[3])
	if err != nil {
		return [3]int{}, false
	}
	return [3]int{year, month, day}, true
}

// compareVersions returns -1, 0, +1.
func compareVersions(a, b [3]int) int {
	for i := range 3 {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	return 0
}

// sanitizeReason renders raw command output into a short, single-line,
// length-capped Reason string fit for capability.Reason. Multi-line
// diagnostics survive (newlines normalized to spaces, runs collapsed)
// so operators still see actionable text, but the capability JSON
// never carries a multi-line blob.
func sanitizeReason(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return "openclaw --version failed"
	}
	s := strings.ReplaceAll(raw, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	s = strings.TrimSpace(s)
	const maxLen = 300
	if len(s) > maxLen {
		s = s[:maxLen] + "..."
	}
	return s
}

// Detect resolves the openclaw executable and inspects `openclaw --version`.
//
// Fail-closed semantics (audit M-8):
//   - Binary missing → Available=false with PATH/env explanation.
//   - --version exits non-zero → Version="", Available=false, Reason
//     sanitized from the command's combined stdout+stderr. We do NOT
//     attempt to extract a version from failure output even if it
//     happens to contain digits that look like a semver (e.g. embedded
//     Node version "20.10.0").
//   - --version exits zero but output doesn't parse → Available=false
//     with parse-error Reason.
//   - --version exits zero and parses but version < MinSupportedVersion
//     → Available=false with upgrade Reason. Version field still
//     carries the observed value for diagnostics.
//   - --version exits zero and parses and version ≥ MinSupportedVersion
//     → Available=true with Version populated.
func Detect(ctx context.Context, env agentbridge.DetectEnv) (agentbridge.DetectResult, error) {
	candidates := detectutil.ResolveExecutableCandidates(DefaultExecutable, envValue(env, EnvOverride))
	if len(candidates) == 0 {
		return agentbridge.DetectResult{
			Available: false,
			Reason:    "openclaw executable not found on PATH and " + EnvOverride + " is not set",
		}, nil
	}

	var first agentbridge.DetectResult
	for i, exe := range candidates {
		res := detectExecutable(ctx, exe)
		if len(candidates) > 1 {
			res.Metadata["path_candidate_count"] = strconv.Itoa(len(candidates))
			res.Metadata["path_candidate_index"] = strconv.Itoa(i + 1)
		}
		if i == 0 {
			first = res
		}
		if res.Available {
			return res, nil
		}
	}

	return first, nil
}
