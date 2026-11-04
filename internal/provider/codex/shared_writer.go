package codex

import (
	"context"
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

const (
	codexMethodThreadRead  codexMethod = "thread/read"
	codexMethodQueueList   codexMethod = "thread/queue/list"
	codexMethodQueueAdd    codexMethod = "thread/queue/add"
	codexMethodQueueDelete codexMethod = "thread/queue/delete"
	sharedStopNotice                   = "요청은 이미 Codex 앱에서 실행 중입니다. Riido의 응답 수신만 중지되었습니다. 실행 중지는 Codex 앱에서 해주세요"
)

type sharedWriter struct {
	clientID     string
	submissionID string
	submitted    bool
	done         bool
	stopping     bool
	turnID       string
	queueCursors map[string]bool
	imported     map[string]bool
	output       map[string]string
}

func (d *protocolDriver) attachSharedWriter(ctx context.Context, io agentbridge.ProtocolIO) ([]agentbridge.Event, []agentbridge.Command, error) {
	if d.req.TaskID == "" || d.req.Prompt == "" {
		return d.failedEvents("codex: shared writer requires request identity and prompt"), nil, nil
	}
	d.threadID = d.req.ResumeSessionID
	key := fmt.Sprintf("%s\x00%s\x00%s\x00%s", d.threadID, d.req.TaskID, d.req.Metadata["run_id"], d.req.Prompt)
	d.shared = &sharedWriter{clientID: fmt.Sprintf("riido-%x", sha256.Sum256([]byte(key))), imported: map[string]bool{}, output: map[string]string{}}
	_, err := d.sendRequest(ctx, io, codexMethodThreadRead, map[string]any{"threadId": d.threadID, "includeTurns": false})
	return nil, nil, err
}

func (d *protocolDriver) PollInterval() time.Duration { return time.Second }

func (d *protocolDriver) OnPoll(ctx context.Context, io agentbridge.ProtocolIO) ([]agentbridge.Event, []agentbridge.Command, error) {
	if d.shared == nil || d.shared.done || d.shared.stopping || len(d.pending) > 0 {
		return nil, nil, nil
	}
	d.historyTurns, d.historyCursors = nil, map[string]bool{}
	return nil, nil, d.requestHistoryPage(ctx, "", io)
}

func (d *protocolDriver) handleSharedResponse(ctx context.Context, method codexMethod, result map[string]any, io agentbridge.ProtocolIO) ([]agentbridge.Event, []agentbridge.Command, error) {
	s := d.shared
	if s == nil {
		return d.failedEvents("codex: unexpected shared writer response"), nil, nil
	}
	switch method {
	case codexMethodThreadRead:
		thread := mapField(result, "thread")
		if stringField(thread, "id") != d.threadID {
			return d.failedEvents("codex: shared thread identity mismatch"), nil, nil
		}
		if d.req.Cwd != "" && !filepath.IsAbs(d.req.Cwd) {
			return d.failedEvents("codex: shared task working directory must be absolute"), nil, nil
		}
		d.historyCursors = map[string]bool{}
		return d.threadIdentifiedEvents(), nil, d.requestHistoryPage(ctx, "", io)
	case codexMethodQueueList:
		return d.handleSharedQueueList(ctx, result, io)
	case codexMethodQueueAdd:
		queued := mapField(result, "queuedSubmission")
		if stringField(queued, "clientUserMessageId") != s.clientID || stringField(queued, "id") == "" {
			return d.failedEvents("codex: invalid queue acknowledgement; check Codex app before retrying"), nil, nil
		}
		s.submissionID, s.submitted = stringField(queued, "id"), true
		if s.stopping {
			return nil, nil, d.deleteSharedQueue(ctx, io)
		}
		return []agentbridge.Event{{Kind: agentbridge.EventProgress, Text: "Codex 앱의 같은 대화에서 요청을 이어갑니다. 앱의 실행 설정을 사용하며, 실행 중 중지는 Codex 앱에서 할 수 있습니다."}}, nil, nil
	case codexMethodQueueDelete:
		s.done = true
		note := ""
		if deleted, _ := result["deleted"].(bool); !deleted {
			note = sharedStopNotice
		}
		return []agentbridge.Event{{Kind: agentbridge.EventResult, Result: agentbridge.Result{Status: agentbridge.ResultCancelled, Error: note}}}, nil, nil
	default:
		return nil, nil, nil
	}
}

func (d *protocolDriver) requestSharedQueue(ctx context.Context, cursor string, io agentbridge.ProtocolIO) error {
	params := map[string]any{"threadId": d.threadID, "limit": 100}
	if cursor != "" {
		params["cursor"] = cursor
	}
	_, err := d.sendRequest(ctx, io, codexMethodQueueList, params)
	return err
}

func (d *protocolDriver) handleSharedQueueList(ctx context.Context, result map[string]any, io agentbridge.ProtocolIO) ([]agentbridge.Event, []agentbridge.Command, error) {
	data, ok := result["data"].([]any)
	if !ok {
		return d.failedEvents("codex: invalid native queue list"), nil, nil
	}
	for _, value := range data {
		entry, _ := value.(map[string]any)
		if stringField(entry, "clientUserMessageId") == d.shared.clientID {
			d.shared.submissionID = stringField(entry, "id")
			if d.shared.submissionID == "" {
				return d.failedEvents("codex: queued request identity missing"), nil, nil
			}
			d.shared.submitted = true
			return nil, nil, nil
		}
	}
	if cursor := stringField(result, "nextCursor"); cursor != "" {
		if d.shared.queueCursors[cursor] || len(d.shared.queueCursors) >= 100 {
			return d.failedEvents("codex: invalid native queue pagination"), nil, nil
		}
		d.shared.queueCursors[cursor] = true
		return nil, nil, d.requestSharedQueue(ctx, cursor, io)
	}
	// Native insertion is not idempotent. Submit once after inspecting history
	// and pending entries; never retry a write whose acknowledgement is uncertain.
	_, err := d.sendRequest(ctx, io, codexMethodQueueAdd, map[string]any{
		"threadId": d.threadID, "clientUserMessageId": d.shared.clientID,
		"input": []map[string]any{{"type": "text", "text": d.sharedPrompt()}},
	})
	return nil, nil, err
}

func (d *protocolDriver) sharedPrompt() string {
	if d.req.Cwd == "" {
		return d.req.Prompt
	}
	// Opening a saved chat can change the desktop's selected cwd. Preserve its
	// settings while carrying the supervisor-validated task directory explicitly.
	return d.req.Prompt + fmt.Sprintf("\n\nRiido task working directory: %q. Resolve this task's relative file paths in that directory; preserve the current Codex app permissions and approval settings.", d.req.Cwd)
}
