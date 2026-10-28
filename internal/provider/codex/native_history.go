package codex

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/provider/history"
)

const codexMethodThreadTurnsList codexMethod = "thread/turns/list"

func (d *protocolDriver) readNativeHistory(ctx context.Context, result map[string]any, io agentbridge.ProtocolIO) ([]agentbridge.Event, []agentbridge.Command, error) {
	thread := mapField(result, "thread")
	if stringField(mapField(thread, "status"), "type") == "active" {
		return d.failedEvents("codex: native session is active in another client"), nil, nil
	}
	turns, _ := thread["turns"].([]any)
	if len(turns) > 0 && stringField(result, "turnsBackwardsCursor") == "" && stringField(result, "itemsBackwardsCursor") == "" {
		return d.finishNativeHistory(ctx, turns, io)
	}
	d.historyCursors = map[string]bool{}
	return nil, nil, d.requestHistoryPage(ctx, "", io)
}

func (d *protocolDriver) requestHistoryPage(ctx context.Context, cursor string, io agentbridge.ProtocolIO) error {
	params := map[string]any{"threadId": d.threadID, "sortDirection": "asc", "itemsView": "full", "limit": 50}
	if cursor != "" {
		params["cursor"] = cursor
	}
	_, err := d.sendRequest(ctx, io, codexMethodThreadTurnsList, params)
	return err
}

func (d *protocolDriver) handleHistoryPage(ctx context.Context, raw agentbridge.RawEvent, io agentbridge.ProtocolIO) ([]agentbridge.Event, []agentbridge.Command, error) {
	result := mapField(raw.Payload, "result")
	turns, ok := result["data"].([]any)
	if !ok {
		return d.failedEvents("codex: invalid native history page"), nil, nil
	}
	d.historyTurns = append(d.historyTurns, turns...)
	if len(d.historyTurns) > 10000 {
		return d.failedEvents("codex: native history exceeds turn limit"), nil, nil
	}
	if cursor := stringField(result, "nextCursor"); cursor != "" {
		if d.historyCursors[cursor] {
			return d.failedEvents("codex: repeated native history cursor"), nil, nil
		}
		d.historyCursors[cursor] = true
		return nil, nil, d.requestHistoryPage(ctx, cursor, io)
	}
	return d.finishNativeHistory(ctx, d.historyTurns, io)
}

func (d *protocolDriver) finishNativeHistory(ctx context.Context, turns []any, io agentbridge.ProtocolIO) ([]agentbridge.Event, []agentbridge.Command, error) {
	messages, err := codexHistoryMessages(d.threadID, turns)
	if err != nil {
		return d.failedEvents(err.Error()), nil, nil
	}
	imported, err := history.External(d.req, messages)
	if err != nil {
		return d.failedEvents(err.Error()), nil, nil
	}
	events := append(d.threadIdentifiedEvents(), imported...)
	d.historyTurns, d.historyCursors = nil, nil
	if _, err := d.sendRequest(ctx, io, codexMethodTurnStart, d.turnStartParams()); err != nil {
		return nil, nil, err
	}
	return events, nil, nil
}

func codexHistoryMessages(sessionID string, turns []any) ([]agentbridge.NativeHistoryMessage, error) {
	var out []agentbridge.NativeHistoryMessage
	for _, value := range turns {
		data, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		var turn struct {
			ID     string `json:"id"`
			Status string `json:"status"`
			Items  []struct {
				ID      string `json:"id"`
				Type    string `json:"type"`
				Text    string `json:"text"`
				Content []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"content"`
			} `json:"items"`
		}
		if json.Unmarshal(data, &turn) != nil || turn.ID == "" || turn.Items == nil {
			return nil, errors.New("codex: incomplete native history turn")
		}
		if turn.Status == "inProgress" {
			return nil, errors.New("codex: native turn is still running")
		}
		for _, item := range turn.Items {
			role, text := "", item.Text
			switch item.Type {
			case "userMessage":
				role = "user"
				var parts []string
				for _, block := range item.Content {
					if block.Type == "text" {
						parts = append(parts, block.Text)
					} else {
						parts = append(parts, "[Non-text user input; attachment not imported.]")
					}
				}
				text = strings.Join(parts, "\n")
			case "agentMessage":
				role = "agent"
			default:
				continue
			}
			if item.ID == "" {
				return nil, errors.New("codex: native message identity missing")
			}
			out = append(out, agentbridge.NativeHistoryMessage{SessionID: sessionID, MessageID: turn.ID + ":" + item.ID, Role: role, Body: text})
		}
	}
	return out, nil
}
