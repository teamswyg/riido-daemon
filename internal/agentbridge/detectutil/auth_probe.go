package detectutil

import (
	"context"
	"strings"
)

type AuthProbeStatus string

const (
	AuthProbeAuthenticated   AuthProbeStatus = "authenticated"
	AuthProbeUnauthenticated AuthProbeStatus = "unauthenticated"
	AuthProbeUnknown         AuthProbeStatus = "unknown"
)

func AuthStatusProbe(ctx context.Context, executable string, args ...string) AuthProbeStatus {
	probe := VersionProbeStrict(ctx, executable, args...)
	if !probe.OK {
		return AuthProbeUnknown
	}
	if probe.ExitCode == 0 {
		return AuthProbeAuthenticated
	}
	output := strings.ToLower(probe.Output)
	for _, marker := range []string{"not logged in", "login required", "authentication required", "not authenticated"} {
		if strings.Contains(output, marker) {
			return AuthProbeUnauthenticated
		}
	}
	return AuthProbeUnknown
}
