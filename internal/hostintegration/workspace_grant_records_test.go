package hostintegration_test

import (
	"reflect"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/hostintegration"
)

func TestWorkspaceGrantStoreRecordsAreDeterministic(t *testing.T) {
	first := workspaceGrantRecord()
	first.WorkspaceID = "workspace-b"
	second := workspaceGrantRecord()
	second.WorkspaceID = "workspace-a"

	store, err := hostintegration.NewWorkspaceGrantStore(first, second)
	if err != nil {
		t.Fatal(err)
	}

	gotRecords := store.Records()
	got := []string{gotRecords[0].WorkspaceID, gotRecords[1].WorkspaceID}
	want := []string{"workspace-a", "workspace-b"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("records order = %v, want %v", got, want)
	}
}
