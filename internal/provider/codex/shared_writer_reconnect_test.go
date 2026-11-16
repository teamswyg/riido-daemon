package codex

import (
	"context"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestSharedWriterReleasedOwnerResumesOnlyAfterConfirmedDeletion(t *testing.T) {
	d, io := sharedPendingFixture(t)
	d.startSharedOwner = func(context.Context, agentbridge.StartRequest, string, string) error { return errSharedOwnerConnect }
	_, _, _ = d.OnRaw(t.Context(), makeResponse(5, map[string]any{"data": []any{map[string]any{"id": "mine", "clientUserMessageId": d.shared.clientID}}}), io)
	deletion := assertSharedMethod(t, io, "thread/queue/delete")
	id := int64(deletion["id"].(float64))
	events, _, err := d.OnRaw(t.Context(), makeResponse(id, map[string]any{"deleted": true}), io)
	if err != nil || len(events) == 0 || events[len(events)-1].Kind == agentbridge.EventResult {
		t.Fatalf("confirmed unstarted request must reconnect, got %+v %v", events, err)
	}
	resume := assertSharedMethod(t, io, "thread/resume")
	if resume["params"].(map[string]any)["threadId"] != "native" {
		t.Fatal("reconnection changed native identity")
	}
	resumeID := int64(resume["id"].(float64))
	_, _, err = d.OnRaw(t.Context(), makeResponse(resumeID, map[string]any{"thread": map[string]any{"id": "native"}}), io)
	if err != nil {
		t.Fatal(err)
	}
	turn := assertSharedMethod(t, io, "turn/start")
	if turn["params"].(map[string]any)["threadId"] != "native" || len(io.writes) != 0 {
		t.Fatal("must send exactly one turn on the same native thread")
	}
}

func TestSharedWriterReconnectNeverRequeuesAnActiveOwner(t *testing.T) {
	d, io := sharedPendingFixture(t)
	d.startSharedOwner = func(context.Context, agentbridge.StartRequest, string, string) error { return errSharedOwnerConnect }
	_, _, _ = d.OnRaw(t.Context(), makeResponse(5, map[string]any{"data": []any{map[string]any{"id": "mine", "clientUserMessageId": d.shared.clientID}}}), io)
	deletion := assertSharedMethod(t, io, "thread/queue/delete")
	_, _, _ = d.OnRaw(t.Context(), makeResponse(int64(deletion["id"].(float64)), map[string]any{"deleted": true}), io)
	resume := assertSharedMethod(t, io, "thread/resume")
	events, _, err := d.OnRaw(t.Context(), agentbridge.RawEvent{Type: "error", Payload: map[string]any{
		"id": int64(resume["id"].(float64)), "error": map[string]any{"message": "thread native already has an active writer"},
	}}, io)
	if err != nil || len(events) == 0 || !strings.Contains(events[len(events)-1].Result.Error, "codex_owner_connect_failed") {
		t.Fatalf("persistent owner did not fail actionably: %+v %v", events, err)
	}
	_, _, _ = d.OnPoll(t.Context(), io)
	if len(io.writes) != 0 {
		t.Fatal("active owner was retried or requeued")
	}
}

func TestSharedWriterConnectFailureWithUncertainDeletionNeverResumes(t *testing.T) {
	for _, reply := range []map[string]any{{"deleted": false}, {}, {"deleted": "true"}} {
		d, io := sharedPendingFixture(t)
		d.startSharedOwner = func(context.Context, agentbridge.StartRequest, string, string) error { return errSharedOwnerConnect }
		_, _, _ = d.OnRaw(t.Context(), makeResponse(5, map[string]any{"data": []any{map[string]any{"id": "mine", "clientUserMessageId": d.shared.clientID}}}), io)
		deletion := assertSharedMethod(t, io, "thread/queue/delete")
		_, _, _ = d.OnRaw(t.Context(), makeResponse(int64(deletion["id"].(float64)), reply), io)
		for len(io.writes) > 0 {
			frame := assertSharedMethod(t, io, "thread/turns/list")
			if frame["method"] == "thread/resume" {
				t.Fatal("uncertain request replayed")
			}
		}
	}
}

func TestSharedWriterCancelledConnectionCleanupNeverResumes(t *testing.T) {
	d, io := sharedPendingFixture(t)
	d.startSharedOwner = func(context.Context, agentbridge.StartRequest, string, string) error { return errSharedOwnerConnect }
	_, _, _ = d.OnRaw(t.Context(), makeResponse(5, map[string]any{"data": []any{map[string]any{"id": "mine", "clientUserMessageId": d.shared.clientID}}}), io)
	deletion := assertSharedMethod(t, io, "thread/queue/delete")
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, _, _ = d.OnRaw(ctx, makeResponse(int64(deletion["id"].(float64)), map[string]any{"deleted": true}), io)
	if len(io.writes) != 0 {
		t.Fatal("cancelled request resumed")
	}
}
