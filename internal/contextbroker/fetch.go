package contextbroker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// Query is one assignment-context read sent to the control-plane with the
// daemon's device credential and the run's lease token.
type Query struct {
	AgentID      string
	AssignmentID string
	LeaseToken   string
	View         string
	Target       string
	Cursor       string
	RequestID    string
}

// ResultSource is one source the control-plane recorded as read.
type ResultSource struct {
	Kind    string `json:"kind"`
	ID      string `json:"id"`
	Part    string `json:"part"`
	Version string `json:"version,omitempty"`
}

// Result is the control-plane answer for one view of the assignment snapshot.
type Result struct {
	Data       json.RawMessage `json:"data"`
	Sources    []ResultSource  `json:"sources"`
	CapturedAt time.Time       `json:"captured_at"`
	Truncated  bool            `json:"truncated"`
	// Live marks views read from current control-plane state at CapturedAt
	// instead of the assignment-time snapshot.
	Live bool `json:"live"`
}

// FetchError is a non-2xx control-plane answer. Message is the server's
// error text, which is safe to show (it never contains tokens).
type FetchError struct {
	Status  int
	Message string
}

func (e *FetchError) Error() string {
	return fmt.Sprintf("assignment context: status %d: %s", e.Status, e.Message)
}

// Fetcher reads the assignment context snapshot. The SaaS control plane
// implements it; other control planes leave it nil.
type Fetcher interface {
	FetchAssignmentContext(ctx context.Context, q Query) (Result, error)
}

// Views understood by the control-plane endpoint.
const (
	viewBase      = "base"
	viewTask      = "task"
	viewTree      = "tree"
	viewResources = "resources"
	viewComments  = "comments"
	viewAgents    = "agents"
	viewAgent     = "agent"
	viewMeetings  = "meetings"
	viewMeeting   = "meeting"
	viewContent   = "content"
)

// viewForCommand maps CLI read commands onto control-plane views: snapshot
// views for work context, live views for the agent directory. Activity,
// search and changes are not served.
func viewForCommand(command string) (string, bool) {
	switch command {
	case "context get":
		return viewBase, true
	case "context part":
		return viewContent, true
	case "work get":
		return viewTask, true
	case "work tree":
		return viewTree, true
	case "work relations":
		return viewResources, true
	case "work comments":
		return viewComments, true
	case "agents list":
		return viewAgents, true
	case "agents show":
		return viewAgent, true
	case "meeting list":
		return viewMeetings, true
	case "meeting get":
		return viewMeeting, true
	}
	return "", false
}
