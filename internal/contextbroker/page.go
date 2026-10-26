package contextbroker

import (
	"encoding/json"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

// commentPageSize is the control-plane's fixed comment page size; --limit is
// not forwarded yet.
const commentPageSize = 20

// commentPage lifts the comments view's next_cursor into the envelope page so
// agents page with `riido work comments <key> --cursor <next>`.
func commentPage(data json.RawMessage) *contextcli.Page {
	var page struct {
		NextCursor string `json:"next_cursor"`
	}
	_ = json.Unmarshal(data, &page)
	return &contextcli.Page{Limit: commentPageSize, NextCursor: page.NextCursor, Truncated: page.NextCursor != ""}
}
