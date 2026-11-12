package codex

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func sharedPendingFixture(t *testing.T) (*protocolDriver, *recordingIO) {
	t.Helper()
	d, io := sharedFixture(t)
	_, _, _ = d.OnRaw(t.Context(), makeResponse(4, map[string]any{"data": []any{}}), io)
	assertSharedMethod(t, io, "thread/queue/list")
	return d, io
}

func TestSharedWriterStartsNewAndRecoveredSubmissionsOnOwner(t *testing.T) {
	for _, recovered := range []bool{false, true} {
		t.Run(map[bool]string{true: "recovered", false: "new"}[recovered], func(t *testing.T) {
			d, io := sharedPendingFixture(t)
			calls := 0
			d.startSharedOwner = func(_ context.Context, req agentbridge.StartRequest, thread, entry string) error {
				calls++
				if thread != "native" || entry != "mine" || req.TaskID != "run" {
					t.Fatalf("wrong target: %s %s %+v", thread, entry, req)
				}
				return nil
			}
			if recovered {
				_, _, _ = d.OnRaw(t.Context(), makeResponse(5, map[string]any{"data": []any{map[string]any{"id": "mine", "clientUserMessageId": d.shared.clientID}}}), io)
			} else {
				_, _, _ = d.OnRaw(t.Context(), makeResponse(5, map[string]any{"data": []any{}}), io)
				assertSharedMethod(t, io, "thread/queue/add")
				_, _, _ = d.OnRaw(t.Context(), makeResponse(6, map[string]any{"queuedSubmission": map[string]any{"id": "mine", "clientUserMessageId": d.shared.clientID}}), io)
			}
			if calls != 1 || !d.shared.startRequested || len(io.writes) != 0 {
				t.Fatal("queue acknowledgement did not start exactly once")
			}
			_, _, _ = d.finishSharedHistory(t.Context(), nil, io)
			if calls != 1 || len(io.writes) != 0 {
				t.Fatal("start duplicated before history materialized")
			}
		})
	}
}

func TestSharedWriterBusyOwnerRechecksQueueBeforeRetry(t *testing.T) {
	d, io := sharedPendingFixture(t)
	calls := 0
	d.startSharedOwner = func(context.Context, agentbridge.StartRequest, string, string) error {
		calls++
		return errSharedOwnerBusy
	}
	entry := map[string]any{"data": []any{map[string]any{"id": "mine", "clientUserMessageId": d.shared.clientID}}}
	_, _, _ = d.OnRaw(t.Context(), makeResponse(5, entry), io)
	_, _, _ = d.finishSharedHistory(t.Context(), []any{sharedTurn("other", "", "interrupted", false)}, io)
	if calls != 1 || len(io.writes) != 0 {
		t.Fatal("started during another active turn")
	}
	_, _, _ = d.finishSharedHistory(t.Context(), nil, io)
	frame := assertSharedMethod(t, io, "thread/queue/list")
	d.startSharedOwner = func(context.Context, agentbridge.StartRequest, string, string) error { calls++; return nil }
	_, _, _ = d.OnRaw(t.Context(), makeResponse(int64(frame["id"].(float64)), entry), io)
	if calls != 2 || !d.shared.startRequested {
		t.Fatal("did not start after owner became idle")
	}
}

func TestSharedWriterConsumedEntryNeverResubmits(t *testing.T) {
	d, io := sharedPendingFixture(t)
	d.startSharedOwner = func(context.Context, agentbridge.StartRequest, string, string) error { return errSharedQueueConsumed }
	_, _, _ = d.OnRaw(t.Context(), makeResponse(5, map[string]any{"data": []any{map[string]any{"id": "mine", "clientUserMessageId": d.shared.clientID}}}), io)
	events, _, _ := d.finishSharedHistory(t.Context(), []any{sharedTurn(d.shared.clientID, "finished", "completed", true)}, io)
	if len(io.writes) != 0 || events[len(events)-1].Result.Output != "finished" {
		t.Fatalf("consumed entry mishandled: %+v", events)
	}

	d, io = sharedPendingFixture(t)
	d.shared.submitted = true // Reconciliation after a busy start races consumption.
	_, _, _ = d.OnRaw(t.Context(), makeResponse(5, map[string]any{"data": []any{map[string]any{"id": "someone-else", "clientUserMessageId": "other"}}}), io)
	if len(io.writes) != 0 || !d.shared.startRequested {
		t.Fatal("missing acknowledged entry must never be re-added")
	}
}

func TestSharedWriterBusyRecoverySearchesAllQueuePages(t *testing.T) {
	d, io := sharedPendingFixture(t)
	d.shared.submitted = true
	calls := 0
	d.startSharedOwner = func(context.Context, agentbridge.StartRequest, string, string) error { calls++; return nil }
	_, _, _ = d.OnRaw(t.Context(), makeResponse(5, map[string]any{"data": []any{}, "nextCursor": "next"}), io)
	frame := assertSharedMethod(t, io, "thread/queue/list")
	_, _, _ = d.OnRaw(t.Context(), makeResponse(int64(frame["id"].(float64)), map[string]any{"data": []any{map[string]any{"id": "mine", "clientUserMessageId": d.shared.clientID}}}), io)
	if calls != 1 {
		t.Fatal("pending request on later page was skipped")
	}
}

func TestSharedWriterOwnerFailureCleansOnlyOwnEntryAndExplainsAction(t *testing.T) {
	for _, outcome := range []string{"deleted", "consumed", "cleanup_failed"} {
		t.Run(outcome, func(t *testing.T) {
			d, io := sharedPendingFixture(t)
			d.startSharedOwner = func(context.Context, agentbridge.StartRequest, string, string) error {
				return errors.New("private provider diagnostic")
			}
			events, _, _ := d.OnRaw(t.Context(), makeResponse(5, map[string]any{"data": []any{map[string]any{"id": "mine", "clientUserMessageId": d.shared.clientID}}}), io)
			if len(events) != 0 {
				t.Fatal("failed before pending entry cleanup acknowledgement")
			}
			frame := assertSharedMethod(t, io, "thread/queue/delete")
			if frame["params"].(map[string]any)["queuedSubmissionId"] != "mine" {
				t.Fatal("deleted another entry")
			}
			id := int64(frame["id"].(float64))
			response := makeResponse(id, map[string]any{"deleted": outcome == "deleted"})
			if outcome == "cleanup_failed" {
				response = agentbridge.RawEvent{Type: "error", Payload: map[string]any{"id": id, "error": map[string]any{"message": "private provider diagnostic"}}}
			}
			events, _, _ = d.OnRaw(t.Context(), response, io)
			result := events[len(events)-1].Result
			if result.Status != agentbridge.ResultFailed || !strings.Contains(result.Error, "Codex 앱을 열고") || strings.Contains(result.Error, "private") {
				t.Fatalf("unactionable/unsafe error: %+v", result)
			}
			if outcome != "deleted" && !strings.Contains(result.Error, "대기열") {
				t.Fatal("uncertain submission must be explained")
			}
			_, _, _ = d.OnPoll(t.Context(), io)
			if len(io.writes) != 0 {
				t.Fatal("polling after failure")
			}
		})
	}
}

type sharedBrokenIO struct{ *recordingIO }

func (sharedBrokenIO) WriteStdin(context.Context, []byte) error {
	return errors.New("broken cleanup pipe")
}

func TestSharedWriterCleanupWriteFailureIsActionable(t *testing.T) {
	d, io := sharedPendingFixture(t)
	d.shared.submissionID, d.shared.submitted = "mine", true
	d.startSharedOwner = func(context.Context, agentbridge.StartRequest, string, string) error {
		return errSharedOwnerUnavailable
	}
	events, _, err := d.startSharedSubmission(t.Context(), sharedBrokenIO{io})
	if err != nil || !d.shared.done || len(events) == 0 || !strings.Contains(events[len(events)-1].Result.Error, "대기열") {
		t.Fatalf("cleanup failure lost: %+v %v", events, err)
	}
}
