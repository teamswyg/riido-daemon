package claude

import (
	"slices"
	"testing"
)

func TestBlockedArgsCoverProtocolCritical(t *testing.T) {
	want := []string{"-p", "--output-format", "--input-format", "--permission-mode", "--mcp-config", "--strict-mcp-config", "--permission-prompt-tool"}
	got := BlockedArgs()
	for _, w := range want {
		if !slices.Contains(got, w) {
			t.Fatalf("BlockedArgs missing %q (got %v)", w, got)
		}
	}
}
