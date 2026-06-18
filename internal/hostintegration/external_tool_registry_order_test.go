package hostintegration

import (
	"reflect"
	"testing"

	"github.com/teamswyg/riido-contracts/provider/capability"
)

func TestExternalToolRegistryRecordsAreDeterministic(t *testing.T) {
	codex := validExternalToolRecord()
	codex.Provider = "codex"

	claude := validExternalToolRecord()
	claude.Provider = "claude"

	registry, err := NewExternalToolRegistry(codex, claude)
	if err != nil {
		t.Fatalf("registry create failed: %v", err)
	}

	records := registry.Records()
	got := []capability.ProviderKind{records[0].Provider, records[1].Provider}
	want := []capability.ProviderKind{"claude", "codex"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("records order = %v, want %v", got, want)
	}
}
