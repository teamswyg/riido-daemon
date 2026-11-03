package userapi

import (
	"context"
	"errors"
	"strings"
	"testing"
)

const threadHistoryFixture = `{"schema_version":"riido-ai-server.v1","task_id":"task-one","assigned_agent_ids":["agent-one"],"threads":[{"thread_id":"thread-one","conversation_id":"conversation-one","task_id":"task-one","agent_id":"agent-one","agent_snapshot_id":"snapshot-one","run_id":"run-one","work_status":"done","assignment_state":"completed","messages":[{"message_id":"human-one","role":"user","body":"질문"},{"message_id":"reply-one","role":"agent","run_id":"run-one","seq":2,"result_message":"답변"},{"message_id":"progress-one","role":"progress","display_message":"진행"}]}],"agent_snapshots":{"snapshot-one":{"agent_id":"agent-one","workspace_id":"workspace-one","name":"에이전트"}}}`

type threadFixtureClient struct {
	body   string
	status int
	calls  int
}

func (c *threadFixtureClient) ReadTaskThreads(context.Context, string, string) ([]byte, int, error) {
	c.calls++
	return []byte(c.body), c.status, nil
}

func TestThreadsRetainedListAndShow(t *testing.T) {
	c := &threadFixtureClient{body: threadHistoryFixture, status: 200}
	r := ThreadReader{Client: c}
	list, err := r.List(t.Context(), "workspace-one", "task-one")
	if err != nil || list.Complete || !list.ProjectionComplete || list.HistoryScope != "server_retained" || len(list.Threads) != 1 || list.Threads[0].MessageCount != 3 || list.Threads[0].ConversationID != "conversation-one" {
		t.Fatal("retained list contract lost", err)
	}
	show, err := r.Show(t.Context(), "workspace-one", "task-one", "thread-one")
	if err != nil || show.Complete || !show.ProjectionComplete || show.Thread.Messages[1].ResultMessage != "답변" || show.Thread.Messages[2].Role != "progress" || show.AgentSnapshot == nil || show.AgentSnapshot.Name != "에이전트" || c.calls != 2 {
		t.Fatal("show lost history or cached authorization", err)
	}
	_, err = r.Show(t.Context(), "workspace-one", "task-one", "missing")
	if !errors.Is(err, ErrThreadNotRetained) {
		t.Fatal("missing retained thread needs scoped absence", err)
	}
}

func TestThreadsRejectIncompleteOrMismatchedProjection(t *testing.T) {
	for name, raw := range map[string]string{
		"truncated":                threadHistoryFixture[:len(threadHistoryFixture)-1],
		"wrong task":               strings.Replace(threadHistoryFixture, `"task_id":"task-one"`, `"task_id":"task-other"`, 1),
		"missing threads":          `{"schema_version":"riido-ai-server.v1","task_id":"task-one","assigned_agent_ids":[]}`,
		"missing messages":         strings.Replace(threadHistoryFixture, `"messages":`, `"other":`, 1),
		"unknown role":             strings.Replace(threadHistoryFixture, `"role":"agent"`, `"role":"admin"`, 1),
		"missing snapshot":         strings.Replace(threadHistoryFixture, `"agent_snapshot_id":"snapshot-one"`, `"agent_snapshot_id":"missing"`, 1),
		"wrong snapshot workspace": strings.Replace(threadHistoryFixture, `"workspace_id":"workspace-one"`, `"workspace_id":"other"`, 1),
		"duplicate top":            `{"task_id":"task-one",` + threadHistoryFixture[1:],
		"duplicate nested":         strings.Replace(threadHistoryFixture, `"role":"user"`, `"role":"agent","role":"user"`, 1),
		"case alias":               strings.Replace(threadHistoryFixture, `"task_id":`, `"Task_id":`, 1),
		"unknown schema":           strings.Replace(threadHistoryFixture, "riido-ai-server.v1", "future", 1),
	} {
		t.Run(name, func(t *testing.T) {
			c := &threadFixtureClient{body: raw, status: 200}
			r := ThreadReader{Client: c}
			got, err := r.Show(t.Context(), "workspace-one", "task-one", "thread-one")
			if got != nil || !errors.Is(err, ErrInvalidResponse) {
				t.Fatal("invalid projection accepted", err)
			}
		})
	}
}

func TestThreadsStatusAndEmptyRetention(t *testing.T) {
	for status, want := range map[int]error{401: ErrAuthentication, 403: ErrPermission, 404: ErrNotFound, 429: ErrRateLimited, 503: ErrUnavailable, 302: ErrUnavailable} {
		c := &threadFixtureClient{body: "secret", status: status}
		got, err := (ThreadReader{Client: c}).List(t.Context(), "workspace-one", "task-one")
		if got != nil || !errors.Is(err, want) {
			t.Fatal(status, err)
		}
	}
	c := &threadFixtureClient{body: `{"schema_version":"riido-ai-server.v1","task_id":"task-one","assigned_agent_ids":[],"threads":[]}`, status: 200}
	got, err := (ThreadReader{Client: c}).List(t.Context(), "workspace-one", "task-one")
	if err != nil || got.Complete || !got.ProjectionComplete || got.Threads == nil {
		t.Fatal("empty projection claimed full history", err)
	}
}
