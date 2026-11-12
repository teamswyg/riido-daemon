package codex

import (
	"context"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (d *protocolDriver) BeginIdleRecovery(ctx context.Context, io agentbridge.ProtocolIO) (bool, error) {
	if d.threadID == "" || d.owned.id == "" || d.shared != nil {
		return false, nil
	}
	id, err := d.sendRequest(ctx, io, codexMethodThreadRead, map[string]any{"threadId": d.threadID, "includeTurns": true})
	d.idleRecoveryID = id
	return err == nil, err
}

func (d *protocolDriver) OnIdleRecoveryRaw(raw agentbridge.RawEvent) (agentbridge.Result, bool) {
	kind := rawFrameType(raw.Type)
	id, ok := rpcID(raw.Payload)
	if !ok || d.idleRecoveryID == 0 || id != d.idleRecoveryID || (kind != rawFrameResponse && kind != rawFrameError) {
		return agentbridge.Result{}, false
	}
	delete(d.pending, id)
	d.idleRecoveryID = 0
	if kind == rawFrameError {
		return agentbridge.Result{}, true
	}
	thread := mapField(mapField(raw.Payload, "result"), "thread")
	if stringField(thread, "id") != d.threadID {
		return agentbridge.Result{}, true
	}
	turns, _ := thread["turns"].([]any)
	for _, value := range turns {
		turn, _ := value.(map[string]any)
		if stringField(turn, "id") != d.owned.id {
			continue
		}
		if stringField(turn, "status") != "completed" || intField(turn, "completedAt") <= 0 {
			break
		}
		items, _ := turn["items"].([]any)
		if output := finalAnswer(items); strings.TrimSpace(output) != "" {
			d.owned.done = true
			return agentbridge.Result{Status: agentbridge.ResultCompleted, SessionID: d.threadID, Output: output}, true
		}
		break
	}
	return agentbridge.Result{}, true
}
