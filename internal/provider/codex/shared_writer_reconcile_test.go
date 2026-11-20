package codex

import (
	"context"
	"errors"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestSharedWriterLostStartAcknowledgementRecoversConsumedTurn(t *testing.T) {
	d, io := sharedPendingFixture(t)
	starts := 0
	d.startSharedOwner = func(context.Context, agentbridge.StartRequest, string, string) error {
		starts++
		return errors.New("lost start response")
	}
	_, _, _ = d.OnRaw(t.Context(), makeResponse(5, map[string]any{"data": []any{map[string]any{"id": "mine", "clientUserMessageId": d.shared.clientID}}}), io)
	frame := assertSharedMethod(t, io, "thread/queue/delete")
	events, _, err := d.OnRaw(t.Context(), makeResponse(int64(frame["id"].(float64)), map[string]any{"deleted": false}), io)
	if err != nil || d.shared.done || d.shared.stopping {
		t.Fatalf("consumed request must still be observed: events=%+v err=%v", events, err)
	}
	for _, e := range events {
		if e.Kind == agentbridge.EventError || e.Kind == agentbridge.EventResult {
			t.Fatalf("declared failure before reading the consumed request: %+v", e)
		}
	}
	_, _, _ = d.finishSharedHistory(t.Context(), []any{sharedTurn("unrelated", "wrong result", "completed", true)}, io)
	if len(io.writes) != 0 || starts != 1 || d.shared.done {
		t.Fatal("uncertain start was duplicated or matched another request")
	}
	_, _, _ = d.OnPoll(t.Context(), io)
	frame = assertSharedMethod(t, io, "thread/turns/list")
	events, _, err = d.OnRaw(t.Context(), makeResponse(int64(frame["id"].(float64)), map[string]any{"data": []any{sharedTurn(d.shared.clientID, "recovered", "completed", true)}}), io)
	if err != nil || len(events) == 0 || events[len(events)-1].Result.Output != "recovered" || starts != 1 || len(io.writes) != 0 {
		t.Fatalf("failed to recover exactly one matching result: %+v %v", events, err)
	}
}
