package agentbridge

import (
	"slices"
	"strings"
	"testing"
)

func TestFilterBlockedArgs(t *testing.T) {
	blocked := []string{"-p", "--output-format", "--permission-mode"}
	custom := []string{"-p", "--output-format", "json", "--permission-mode=bypassPermissions", "--keep", "value"}
	kept, dropped := FilterBlockedArgs(custom, blocked)
	if strings.Join(kept, " ") != "--keep value" {
		t.Fatalf("kept wrong: %v", kept)
	}
	if len(dropped) == 0 {
		t.Fatalf("dropped should be non-empty: %v", dropped)
	}
	for _, badArg := range []string{"-p", "--output-format", "json"} {
		if !slices.Contains(dropped, badArg) {
			t.Fatalf("expected %q in dropped, got %v", badArg, dropped)
		}
	}
}
