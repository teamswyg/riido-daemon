package userapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"
)

type agentReadFixtureClient struct {
	raw    []byte
	status int
	after  error
}

func (f agentReadFixtureClient) ReadTaskAgents(_ context.Context, _, _, _, _ string, consume func([]byte, int) error) error {
	if err := consume(f.raw, f.status); err != nil {
		return err
	}
	return f.after
}

const agentEntryFixture = `{"agent_id":"a","name":"Agent","role_status":"missing","work_status":"idle","runtime":{"connection":"online"},"callable":true}`

func agentPageFixture(total int, cursor string, count int, next string) []byte {
	entries := make([]json.RawMessage, count)
	for i := range entries {
		entries[i] = json.RawMessage(strings.Replace(agentEntryFixture, `"a"`, fmt.Sprintf(`"a%d"`, i), 1))
	}
	b, _ := json.Marshal(map[string]any{"agents": entries, "total": total, "cursor": cursor, "next_cursor": next, "fetched_at": "2026-11-01T00:00:00Z", "policy_checked": false})
	return b
}

func TestUserAgentPagesAndFinalScopeError(t *testing.T) {
	for _, tc := range []struct {
		total  int
		cursor string
		count  int
		next   string
	}{{53, "0", 50, "1"}, {53, "1", 3, ""}, {53, strconv.Itoa(int(^uint(0) >> 1)), 0, ""}, {0, "0", 0, ""}} {
		reader := AgentReader{Client: agentReadFixtureClient{raw: agentPageFixture(tc.total, tc.cursor, tc.count, tc.next), status: 200}}
		page, err := reader.List(context.Background(), "workspace", "task", tc.cursor)
		if err != nil || len(page.Agents) != tc.count || page.NextCursor != tc.next {
			t.Fatal("valid current page rejected", tc, err)
		}
	}
	stop := errors.New("session changed after parsing")
	reader := AgentReader{Client: agentReadFixtureClient{raw: agentPageFixture(1, "0", 1, ""), status: 200, after: stop}}
	if page, err := reader.List(context.Background(), "workspace", "task", ""); page != nil || !errors.Is(err, stop) {
		t.Fatal("parsed result escaped final session fence", page, err)
	}
	reader.Client = agentReadFixtureClient{raw: []byte(`{"agent":` + agentEntryFixture + `,"fetched_at":"2026-11-01T00:00:00Z","policy_checked":false}`), status: 200, after: stop}
	if detail, err := reader.Show(context.Background(), "workspace", "task", "a"); detail != nil || !errors.Is(err, stop) {
		t.Fatal("detail escaped final fence", detail, err)
	}
}

func TestUserAgentPageRejectsMalformedPublicProjection(t *testing.T) {
	valid := string(agentPageFixture(1, "0", 1, ""))
	for name, body := range map[string]string{
		"partial count":    strings.Replace(valid, `"total":1`, `"total":2`, 1),
		"extra page":       strings.Replace(valid, `"next_cursor":""`, `"next_cursor":"1"`, 1),
		"wrong page":       strings.Replace(valid, `"cursor":"0"`, `"cursor":"1"`, 1),
		"null agents":      strings.Replace(valid, `"agents":[`+strings.Replace(agentEntryFixture, `"a"`, `"a0"`, 1)+`]`, `"agents":null`, 1),
		"missing policy":   strings.Replace(valid, `"policy_checked":false,`, ``, 1),
		"policy lie":       strings.Replace(valid, `"policy_checked":false`, `"policy_checked":true`, 1),
		"self lie":         strings.Replace(valid, `"agent_id":"a0"`, `"agent_id":"a0","is_current_agent":true`, 1),
		"private field":    strings.Replace(valid, `"agent_id":"a0"`, `"agent_id":"a0","owner_principal_id":"private"`, 1),
		"null boolean":     strings.Replace(valid, `"callable":true`, `"callable":null`, 1),
		"wrong status":     strings.Replace(valid, `"work_status":"idle"`, `"work_status":"other"`, 1),
		"missing identity": strings.Replace(valid, `"agent_id":"a0",`, ``, 1),
		"case alias":       strings.Replace(valid, `"total":1`, `"Total":1`, 1),
		"duplicate":        strings.Replace(valid, `"total":1`, `"total":1,"total":1`, 1),
		"trailing":         valid + ` {}`,
		"too many":         string(agentPageFixture(51, "0", 51, "")),
	} {
		t.Run(name, func(t *testing.T) {
			if body == valid {
				t.Fatal("mutation did not change fixture")
			}
			reader := AgentReader{Client: agentReadFixtureClient{raw: []byte(body), status: 200}}
			if page, err := reader.List(context.Background(), "workspace", "task", ""); err == nil || page != nil {
				t.Fatal("invalid page accepted", page, err)
			}
		})
	}
}
