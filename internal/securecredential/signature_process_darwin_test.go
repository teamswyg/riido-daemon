//go:build darwin

package securecredential

import (
	"slices"
	"testing"
)

// The pid target skips only resource re-reading; the path check keeps it.
func TestProcessVerificationSkipsOnlyResources(t *testing.T) {
	if !slices.Equal(processVerifyArgs, []string{"--ignore-resources"}) {
		t.Fatalf("process verify args = %v", processVerifyArgs)
	}
}
