package riidoapi

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func TestServerExposesTaskDB(t *testing.T) {
	socketPath, taskDBPath, stop := serveTestAPI(t)
	defer stop()

	client := NewClient(socketPath)
	var status Status
	if err := client.Request(context.Background(), "status", nil, &status); err != nil {
		t.Fatalf("status request failed: %v", err)
	}
	assertStatus(t, status, taskDBPath)

	var db taskdb.TaskDB
	if err := client.Request(context.Background(), "tasks", nil, &db); err != nil {
		t.Fatalf("tasks request failed: %v", err)
	}
	assertTaskDB(t, db)
}

func assertStatus(t *testing.T, status Status, taskDBPath string) {
	t.Helper()
	if status.SchemaVersion != StatusSchemaVersion {
		t.Fatalf("unexpected status schema: %s", status.SchemaVersion)
	}
	if status.Transport != string(LocalTransportUnixSocket) {
		t.Fatalf("unexpected API transport: %s", status.Transport)
	}
	if status.AppVersion != "riido-daemon test.v1" {
		t.Fatalf("unexpected app version: %s", status.AppVersion)
	}
	if status.TaskCount != 1 || status.EvidenceCount != 0 || status.CommandReceiptCount != 0 {
		t.Fatalf("unexpected status counts: %#v", status)
	}
	if status.TaskDBPath != taskDBPath {
		t.Fatalf("unexpected task DB path: %s", status.TaskDBPath)
	}
}

func assertTaskDB(t *testing.T, db taskdb.TaskDB) {
	t.Helper()
	if db.SchemaVersion != taskdb.TaskDBSchemaVersion {
		t.Fatalf("unexpected task DB schema: %s", db.SchemaVersion)
	}
	if db.Tasks[0].ID != "task:test" {
		t.Fatalf("unexpected task: %#v", db.Tasks[0])
	}
	if db.Tasks[0].RecommendedProvider != "codex" || !db.Tasks[0].RequiresHumanApproval {
		t.Fatalf("unexpected task gate fields: %#v", db.Tasks[0])
	}
}
