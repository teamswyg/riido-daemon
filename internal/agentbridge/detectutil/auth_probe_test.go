package detectutil

import (
	"context"
	"testing"
)

func TestAuthStatusProbeClassifiesBoundedStates(t *testing.T) {
	tests := []struct {
		command string
		want    AuthProbeStatus
	}{
		{"exit 0", AuthProbeAuthenticated},
		{"printf 'Not logged in'; exit 0", AuthProbeUnauthenticated},
		{"printf 'Login required' >&2; exit 0", AuthProbeUnauthenticated},
		{"printf 'Authentication required'; exit 0", AuthProbeUnauthenticated},
		{"printf 'Not authenticated'; exit 0", AuthProbeUnauthenticated},
		{"printf 'Not logged in'; exit 1", AuthProbeUnauthenticated},
		{"printf 'unexpected provider output'; exit 1", AuthProbeUnknown},
	}
	for _, tt := range tests {
		if got := AuthStatusProbe(context.Background(), "/bin/sh", "-c", tt.command); got != tt.want {
			t.Fatalf("AuthStatusProbe(%q) = %q, want %q", tt.command, got, tt.want)
		}
	}
}
