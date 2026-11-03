package userapi

import (
	"encoding/json"
	"errors"
	"testing"
)

func threadArrayFixture(t *testing.T, field string, value any) string {
	t.Helper()
	var fixture map[string]any
	if err := json.Unmarshal([]byte(threadHistoryFixture), &fixture); err != nil {
		t.Fatal(err)
	}
	if field == "messages" {
		fixture["threads"].([]any)[0].(map[string]any)[field] = value
	} else {
		fixture[field] = value
	}
	raw, err := json.Marshal(fixture)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func TestThreadsRequiredArraysRejectNull(t *testing.T) {
	for _, field := range []string{"messages", "assigned_agent_ids", "threads"} {
		t.Run(field, func(t *testing.T) {
			r := ThreadReader{Client: &threadFixtureClient{body: threadArrayFixture(t, field, nil), status: 200}}
			list, err := r.List(t.Context(), "workspace-one", "task-one")
			if list != nil || !errors.Is(err, ErrInvalidResponse) {
				t.Fatal("null array became complete empty projection", err)
			}
			show, err := r.Show(t.Context(), "workspace-one", "task-one", "thread-one")
			if show != nil || !errors.Is(err, ErrInvalidResponse) {
				t.Fatal("null array became retained show or absence", err)
			}
		})
	}
}

func TestThreadsRequiredArraysAcceptActualEmptyArrays(t *testing.T) {
	for _, field := range []string{"messages", "assigned_agent_ids", "threads"} {
		t.Run(field, func(t *testing.T) {
			r := ThreadReader{Client: &threadFixtureClient{body: threadArrayFixture(t, field, []any{}), status: 200}}
			list, err := r.List(t.Context(), "workspace-one", "task-one")
			if err != nil || !list.ProjectionComplete || list.Complete {
				t.Fatal("valid empty array rejected", err)
			}
			if field == "messages" && list.Threads[0].MessageCount != 0 {
				t.Fatal("empty messages counted incorrectly")
			}
		})
	}
}
