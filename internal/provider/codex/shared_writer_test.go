package codex

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func sharedFixture(t *testing.T) (*protocolDriver, *recordingIO) {
	t.Helper()
	driver, _ := NewProtocolDriver(agentbridge.StartRequest{TaskID: "run", ResumeSessionID: "native", Cwd: "/work", Prompt: "continue"})
	d := driver.(*protocolDriver)
	io := newRecordingIO()
	_ = d.OnStart(t.Context(), io)
	_ = io.next(t, time.Second)
	_, _, _ = d.OnRaw(t.Context(), makeResponse(1, nil), io)
	_ = io.next(t, time.Second)
	_ = io.next(t, time.Second)
	events, _, err := d.OnRaw(t.Context(), agentbridge.RawEvent{Type: "error", Payload: map[string]any{"id": int64(2), "error": map[string]any{"message": "thread native already has an active writer"}}}, io)
	if err != nil || len(events) != 0 {
		t.Fatalf("attach: %+v %v", events, err)
	}
	assertSharedMethod(t, io, "thread/read")
	_, _, _ = d.OnRaw(t.Context(), makeResponse(3, map[string]any{"thread": map[string]any{"id": "native", "cwd": "/work"}}), io)
	assertSharedMethod(t, io, "thread/turns/list")
	return d, io
}

func assertSharedMethod(t *testing.T, io *recordingIO, method string) map[string]any {
	t.Helper()
	var frame map[string]any
	if err := json.Unmarshal(io.next(t, time.Second), &frame); err != nil {
		t.Fatal(err)
	}
	if frame["method"] != method {
		t.Fatalf("want %s: %+v", method, frame)
	}
	return frame
}

func sharedTurn(clientID, answer, status string, finished bool) map[string]any {
	turn := nativeHistoryTurn("matched", "continue", answer)
	turn["status"] = status
	turn["items"].([]any)[0].(map[string]any)["clientId"] = clientID
	if finished {
		turn["completedAt"] = float64(123)
	}
	return turn
}

func TestSharedWriterContinuesSameIDAndCorrelatesOnlyOwnCompletedTurn(t *testing.T) {
	d, io := sharedFixture(t)
	_, _, _ = d.OnRaw(t.Context(), makeResponse(4, map[string]any{"data": []any{}}), io)
	assertSharedMethod(t, io, "thread/queue/list")
	_, _, _ = d.OnRaw(t.Context(), makeResponse(5, map[string]any{"data": []any{}}), io)
	frame := assertSharedMethod(t, io, "thread/queue/add")
	params := frame["params"].(map[string]any)
	if params["threadId"] != "native" || params["clientUserMessageId"] != d.shared.clientID {
		t.Fatal(params)
	}
	input := params["input"].([]any)[0].(map[string]any)["text"].(string)
	if !strings.HasPrefix(input, "continue") || !strings.Contains(input, `"/work"`) {
		t.Fatal("task cwd lost", input)
	}
	_, _, _ = d.OnRaw(t.Context(), makeResponse(6, map[string]any{"queuedSubmission": map[string]any{"id": "submission", "clientUserMessageId": d.shared.clientID}}), io)

	other := sharedTurn("another-client", "unrelated output", "completed", true)
	other["id"] = "other-turn"
	active := sharedTurn(d.shared.clientID, "", "interrupted", false)
	events, _, err := d.finishSharedHistory(t.Context(), []any{other, active}, io)
	if err != nil {
		t.Fatal(err)
	}
	for _, event := range events {
		if event.Kind == agentbridge.EventResult || strings.Contains(event.Text, "unrelated") {
			t.Fatalf("premature/mixed result: %+v", events)
		}
	}
	active["completedAt"], active["status"] = float64(124), "completed"
	active["items"].([]any)[2].(map[string]any)["text"] = "preserved context"
	events, _, err = d.finishSharedHistory(t.Context(), []any{other, active}, io)
	if err != nil || events[len(events)-1].Result.Output != "preserved context" || events[len(events)-1].Result.Status != agentbridge.ResultCompleted {
		t.Fatalf("completion: %+v %v", events, err)
	}
	_, _, _ = d.OnPoll(t.Context(), io)
	if len(io.writes) != 0 {
		t.Fatal("poll after completion")
	}
}

func TestSharedWriterReusesQueuedRequestAndDoesNotResubmit(t *testing.T) {
	d, io := sharedFixture(t)
	_, _, _ = d.OnRaw(t.Context(), makeResponse(4, map[string]any{"data": []any{}}), io)
	assertSharedMethod(t, io, "thread/queue/list")
	_, _, _ = d.OnRaw(t.Context(), makeResponse(5, map[string]any{"data": []any{map[string]any{"id": "existing", "clientUserMessageId": d.shared.clientID}}}), io)
	if !d.shared.submitted || d.shared.submissionID != "existing" || len(io.writes) != 0 {
		t.Fatal("queued request duplicated")
	}
	_, _, _ = d.OnPoll(t.Context(), io)
	assertSharedMethod(t, io, "thread/turns/list")
	_, _, _ = d.OnPoll(t.Context(), io)
	if len(io.writes) != 0 {
		t.Fatal("overlapping poll")
	}
}

func TestSharedWriterWaitsForOtherActiveTurnAndIgnoresNotifications(t *testing.T) {
	d, io := sharedFixture(t)
	_, _, _ = d.OnRaw(t.Context(), makeResponse(4, map[string]any{"data": []any{sharedTurn("other", "", "interrupted", false)}}), io)
	if len(io.writes) != 0 {
		t.Fatal("queued before native history settled")
	}
	events, _, _ := d.OnRaw(t.Context(), makeNotification("turn/completed", map[string]any{"output": "other answer"}), io)
	if len(events) != 0 {
		t.Fatal("unowned notification accepted")
	}
}

func TestSharedWriterPendingCancellationAcknowledgesOnlyOwnQueueEntry(t *testing.T) {
	for _, deleted := range []bool{true, false} {
		t.Run(map[bool]string{true: "pending", false: "already_consumed"}[deleted], func(t *testing.T) {
			d, io := sharedFixture(t)
			d.pending = map[int64]pendingRequest{}
			d.shared.submissionID, d.shared.submitted = "mine", true
			wait, err := d.OnTerminate(t.Context(), io)
			if !wait || err != nil {
				t.Fatal(wait, err)
			}
			frame := assertSharedMethod(t, io, "thread/queue/delete")
			if frame["params"].(map[string]any)["queuedSubmissionId"] != "mine" {
				t.Fatal(frame)
			}
			events, _, _ := d.OnRaw(t.Context(), makeResponse(int64(frame["id"].(float64)), map[string]any{"deleted": deleted}), io)
			if events[0].Result.Status != agentbridge.ResultCancelled || (events[0].Result.Error == "") != deleted {
				t.Fatal(events)
			}
		})
	}
}

func TestSharedWriterRejectsWrongIdentityAndUnsupportedQueue(t *testing.T) {
	d, io := sharedFixture(t)
	events, _, _ := d.handleSharedResponse(t.Context(), codexMethodThreadRead, map[string]any{"thread": map[string]any{"id": "wrong", "cwd": "/work"}}, io)
	if events[len(events)-1].Result.Status != agentbridge.ResultFailed {
		t.Fatal(events)
	}
	d.pending[9] = pendingRequest{method: codexMethodQueueAdd}
	events, _, _ = d.OnRaw(t.Context(), agentbridge.RawEvent{Type: "error", Payload: map[string]any{"id": int64(9), "error": map[string]any{"message": "unsupported method"}}}, io)
	if events[len(events)-1].Result.Status != agentbridge.ResultFailed || len(io.writes) != 0 {
		t.Fatal("unsupported queue should fail without a new thread")
	}
}

func TestSharedWriterImportsExternalHistoryOnceAndPreservesMessageIDs(t *testing.T) {
	d, io := sharedFixture(t)
	d.pending = map[int64]pendingRequest{}
	d.req.Prompt = "# Riido AI Agent Assignment\n- task_id: task\ncontinue"
	d.req.Metadata = map[string]string{"task_id": "task"}
	owned := nativeHistoryTurn("riido", "# Riido AI Agent Assignment\n- task_id: task\nfirst", "first answer")
	owned["completedAt"] = float64(1)
	external := nativeHistoryTurn("external", "outside question", "outside answer")
	external["completedAt"] = float64(2)
	events, _, err := d.finishSharedHistory(t.Context(), []any{owned, external}, io)
	if err != nil || len(events) != 2 || events[0].NativeHistory.MessageID != "external:user" || events[1].NativeHistory.Body != "outside answer" {
		t.Fatalf("import: %+v %v", events, err)
	}
	assertSharedMethod(t, io, "thread/queue/list")
	d.shared.submitted = true
	events, _, err = d.finishSharedHistory(t.Context(), []any{owned, external}, io)
	if err != nil || len(events) != 0 {
		t.Fatal("history repeated", events, err)
	}
}

func TestSharedWriterWaitsForFirstNativeAnchorToFinish(t *testing.T) {
	d, io := sharedFixture(t)
	d.req.Prompt = "# Riido AI Agent Assignment\n- task_id: task\ncontinue"
	d.req.Metadata = map[string]string{"task_id": "task"}
	active := sharedTurn("other", "", "interrupted", false)
	events, _, err := d.finishSharedHistory(t.Context(), []any{active}, io)
	if err != nil || len(events) != 0 || len(io.writes) != 0 {
		t.Fatal("failed before first anchor settled", events, err)
	}
}
