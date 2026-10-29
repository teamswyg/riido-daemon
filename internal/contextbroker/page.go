package contextbroker

import (
	"encoding/json"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

// Fixed control-plane page sizes; --limit is not forwarded yet.
const (
	commentPageSize = 20
	agentPageSize   = 50
)

// cursorPage lifts a paged view's next_cursor into the envelope page so
// agents continue with `--cursor <next>` (work comments, agents list).
func cursorPage(view string, data json.RawMessage) *contextcli.Page {
	var page struct {
		NextCursor string `json:"next_cursor"`
	}
	_ = json.Unmarshal(data, &page)
	limit := commentPageSize
	switch view {
	case viewAgents:
		limit = agentPageSize
	case viewContent:
		limit = 1
	}
	return &contextcli.Page{Limit: limit, NextCursor: page.NextCursor, Truncated: page.NextCursor != ""}
}

// queryTarget is the read's target: a work item for work views, an agent ID
// for `agents show`, or a meeting ID for `meeting get`.
func queryTarget(args map[string]string) string {
	if meeting := strings.TrimSpace(args["meeting"]); meeting != "" {
		return meeting
	}
	if agent := strings.TrimSpace(args["agent"]); agent != "" {
		return agent
	}
	return strings.TrimSpace(args["work"])
}

// freshnessWarning tells the agent how current the answer is: live views are
// read now, every other view is the assignment-time snapshot.
func freshnessWarning(live bool, fetchedAt string) contextcli.Warning {
	if live {
		return contextcli.Warning{Code: "LIVE_STATUS", Message: fetchedAt + "에 조회한 현재 상태입니다. Agent를 멘션하는 시점에 서버가 다시 확인합니다"}
	}
	return contextcli.Warning{Code: "ASSIGNMENT_SNAPSHOT", Message: "배정 시점(" + fetchedAt + ")에 저장된 맥락입니다. 이후 변경은 반영되지 않습니다"}
}
