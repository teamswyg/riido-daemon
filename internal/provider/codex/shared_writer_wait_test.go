package codex

import (
	"context"
	"testing"
	"testing/synctest"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestSharedWriterWaitsAndReconnectsWithoutUserRetry(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		d, io := sharedPendingFixture(t)
		clientID := d.shared.clientID
		d.startSharedOwner = func(context.Context, agentbridge.StartRequest, string, string) error { return errSharedOwnerConnect }
		_, _, _ = d.OnRaw(t.Context(), makeResponse(5, map[string]any{"data": []any{map[string]any{"id": "mine", "clientUserMessageId": clientID}}}), io)
		if !d.ConsumeIdleGrace() {
			t.Fatal("connect-failure cleanup lost coordination grace")
		}
		deletion := assertSharedMethod(t, io, "thread/queue/delete")
		_, _, _ = d.OnRaw(t.Context(), makeResponse(int64(deletion["id"].(float64)), map[string]any{"deleted": true}), io)
		resume := assertSharedMethod(t, io, "thread/resume")
		events, _, err := d.OnRaw(t.Context(), agentbridge.RawEvent{Type: "error", Payload: map[string]any{
			"id": int64(resume["id"].(float64)), "error": map[string]any{"message": "thread native already has an active writer"},
		}}, io)
		for _, event := range events {
			if event.Kind == agentbridge.EventResult {
				t.Fatalf("a live owner must keep the request waiting: %+v", event.Result)
			}
		}
		if err != nil || d.shared == nil || d.shared.clientID != clientID || !d.ConsumeIdleGrace() {
			t.Fatal("waiting lost request identity or its coordination grace", err)
		}
		_, _, _ = d.OnPoll(t.Context(), io)
		if len(io.writes) != 0 {
			t.Fatal("reconnect ignored backoff")
		}
		time.Sleep(10 * time.Second)
		_, _, _ = d.OnPoll(t.Context(), io)
		history := assertSharedMethod(t, io, "thread/turns/list")
		_, _, _ = d.OnRaw(t.Context(), makeResponse(int64(history["id"].(float64)), map[string]any{"data": []any{}}), io)
		queue := assertSharedMethod(t, io, "thread/queue/list")
		_, _, _ = d.OnRaw(t.Context(), makeResponse(int64(queue["id"].(float64)), map[string]any{"data": []any{}}), io)
		add := assertSharedMethod(t, io, "thread/queue/add")
		if add["params"].(map[string]any)["clientUserMessageId"] != clientID {
			t.Fatal("recovery changed request identity")
		}
		starts := 0
		d.startSharedOwner = func(context.Context, agentbridge.StartRequest, string, string) error { starts++; return nil }
		_, _, _ = d.OnRaw(t.Context(), makeResponse(int64(add["id"].(float64)), map[string]any{"queuedSubmission": map[string]any{"id": "replacement", "clientUserMessageId": clientID}}), io)
		if starts != 1 || !d.shared.startRequested || d.ConsumeIdleGrace() {
			t.Fatal("owner recovery did not leave coordination wait")
		}
		events, _, _ = d.finishSharedHistory(t.Context(), []any{sharedTurn(clientID, "reconnected", "completed", true)}, io)
		if events[len(events)-1].Result.Output != "reconnected" || len(io.writes) != 0 {
			t.Fatal("result lost or request duplicated")
		}
	})
}

func TestSharedWriterObservedExecutionEndsReconnectGrace(t *testing.T) {
	d, io := sharedPendingFixture(t)
	d.waitSharedReconnect()
	_, _, _ = d.finishSharedHistory(t.Context(), []any{sharedTurn(d.shared.clientID, "working", "inProgress", false)}, io)
	if d.ConsumeIdleGrace() {
		t.Fatal("an observed turn retained connection-wait grace")
	}
}

func TestSharedWriterReconnectCancellation(t *testing.T) {
	for _, stage := range []string{"cleanup", "resume", "backoff", "queue_read"} {
		t.Run(stage, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				d, io := sharedPendingFixture(t)
				d.startSharedOwner = func(context.Context, agentbridge.StartRequest, string, string) error { return errSharedOwnerConnect }
				_, _, _ = d.OnRaw(t.Context(), makeResponse(5, map[string]any{"data": []any{map[string]any{"id": "mine", "clientUserMessageId": d.shared.clientID}}}), io)
				pending := assertSharedMethod(t, io, "thread/queue/delete")
				response := map[string]any{"deleted": true}
				if stage != "cleanup" {
					_, _, _ = d.OnRaw(t.Context(), makeResponse(int64(pending["id"].(float64)), response), io)
					pending = assertSharedMethod(t, io, "thread/resume")
					response = map[string]any{"thread": map[string]any{"id": "native"}}
					if stage != "resume" {
						_, _, _ = d.OnRaw(t.Context(), agentbridge.RawEvent{Type: "error", Payload: map[string]any{
							"id": int64(pending["id"].(float64)), "error": map[string]any{"message": "thread native already has an active writer"},
						}}, io)
						pending = nil
						if stage == "queue_read" {
							time.Sleep(10 * time.Second)
							_, _, _ = d.OnPoll(t.Context(), io)
							pending = assertSharedMethod(t, io, "thread/turns/list")
							_, _, _ = d.OnRaw(t.Context(), makeResponse(int64(pending["id"].(float64)), map[string]any{"data": []any{}}), io)
							pending = assertSharedMethod(t, io, "thread/queue/list")
							response = map[string]any{"data": []any{}}
						}
					}
				}
				wait, err := d.OnTerminate(t.Context(), io)
				if err != nil || wait != (stage == "cleanup") {
					t.Fatalf("cleanup wait=%v err=%v", wait, err)
				}
				if pending != nil {
					_, _, err = d.OnRaw(t.Context(), makeResponse(int64(pending["id"].(float64)), response), io)
					if err != nil {
						t.Fatal(err)
					}
				}
				time.Sleep(20 * time.Second)
				_, _, _ = d.OnPoll(t.Context(), io)
				if len(io.writes) != 0 || d.ConsumeIdleGrace() {
					t.Fatal("cancelled request was restarted or retained idle grace")
				}
			})
		})
	}
}

func TestSharedWriterReconnectBackoffIsCappedAndAnnouncedOnce(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		d, _ := sharedPendingFixture(t)
		for i, delay := range []time.Duration{2 * time.Second, 4 * time.Second, 8 * time.Second, 10 * time.Second, 10 * time.Second, 10 * time.Second} {
			events := d.waitSharedReconnect()
			if (i == 0 && len(events) != 1) || (i > 0 && len(events) != 0) {
				t.Fatal("wait notice missing or repeated")
			}
			if got := time.Until(d.shared.retryAt); got != delay {
				t.Fatalf("attempt %d delay=%s", i, got)
			}
			time.Sleep(delay)
		}
	})
}
