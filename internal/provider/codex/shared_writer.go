package codex

import (
	"context"
	"crypto/sha256"
	"errors"
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
	clientID       string
	submissionID   string
	submitted      bool
	startRequested bool
	startError     string
	done           bool
	stopping       bool
	turnID         string
	queueCursors   map[string]bool
	imported       map[string]bool
	output         map[string]string
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
		return d.startSharedSubmission(ctx, io)
	case codexMethodQueueDelete:
		s.done = true
		if s.startError != "" {
			message := s.startError
			if deleted, _ := result["deleted"].(bool); !deleted {
				message += ". 요청이 이미 실행되었을 수 있으니 재요청 전에 Codex 앱의 결과와 대기열을 확인해주세요"
			}
			return d.failedEvents(message), nil, nil
		}
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
			return d.startSharedSubmission(ctx, io)
		}
	}
	if cursor := stringField(result, "nextCursor"); cursor != "" {
		if d.shared.queueCursors[cursor] || len(d.shared.queueCursors) >= 100 {
			return d.failedEvents("codex: invalid native queue pagination"), nil, nil
		}
		d.shared.queueCursors[cursor] = true
		return nil, nil, d.requestSharedQueue(ctx, cursor, io)
	}
	if d.shared.submitted {
		// Our acknowledged entry has been consumed or removed. Reconcile by
		// client ID in history; never add it again or start somebody else's entry.
		d.shared.startRequested = true
		return nil, nil, nil
	}
	// Native insertion is not idempotent. Submit once after inspecting history
	// and pending entries; never retry a write whose acknowledgement is uncertain.
	_, err := d.sendRequest(ctx, io, codexMethodQueueAdd, map[string]any{
		"threadId": d.threadID, "clientUserMessageId": d.shared.clientID,
		"input": []map[string]any{{"type": "text", "text": d.sharedPrompt()}},
	})
	return nil, nil, err
}

func (d *protocolDriver) startSharedSubmission(ctx context.Context, io agentbridge.ProtocolIO) ([]agentbridge.Event, []agentbridge.Command, error) {
	s := d.shared
	if s.startRequested || s.stopping {
		return nil, nil, nil
	}
	err := d.startSharedOwner(ctx, d.req, d.threadID, s.submissionID)
	if errors.Is(err, errSharedOwnerBusy) {
		// The owner can become busy between the history read and queue/start.
		// Recheck idle history and the same pending entry before trying again.
		return nil, nil, nil
	}
	if err != nil && !errors.Is(err, errSharedQueueConsumed) {
		s.startError, s.stopping = sharedOwnerUnavailable, true
		if err := d.deleteSharedQueue(ctx, io); err != nil {
			s.done = true
			//nolint:nilerr // The terminal result carries the actionable failure instead of a raw pipe error.
			return d.failedEvents(s.startError + ". 대기 중인 요청을 정리하지 못했습니다. 재요청 전에 Codex 앱의 대기열을 확인해주세요"), nil, nil
		}
		return nil, nil, nil
	}
	s.startRequested = true
	return []agentbridge.Event{{Kind: agentbridge.EventProgress, Text: "Codex 앱의 같은 대화에서 요청을 이어갑니다. 앱의 실행 설정을 사용하며, 실행 중 중지는 Codex 앱에서 할 수 있습니다."}}, nil, nil
}

func (d *protocolDriver) sharedPrompt() string {
	if d.req.Cwd == "" {
		return d.req.Prompt
	}
	// Opening a saved chat can change the desktop's selected cwd. Preserve its
	// settings while carrying the supervisor-validated task directory explicitly.
	return d.req.Prompt + fmt.Sprintf("\n\nRiido task working directory: %q. Resolve this task's relative file paths in that directory; preserve the current Codex app permissions and approval settings.", d.req.Cwd)
}
