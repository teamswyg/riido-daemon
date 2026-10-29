package contextbroker

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

func TestCapturedBudgetPreservesContentAndContinuation(t *testing.T) {
	for _, tc := range []struct {
		name, command, prefix, suffix string
	}{
		{"part", "context part", `{"source_id":"work-one","part":"editor","version":"v1","offset":0,"complete":false,"data":{"content":"`, `"},"next_cursor":"page-two"}`},
		{"work", "work get", `{"resource":{"id":"work-one","kind":"component"},"parts":[{"manifest":{"complete":true,"cursors":["page-one"]},"page":{"data":{"content":"`, `"},"complete":true}}]}`},
		{"meeting", "meeting get", `{"resource":{"id":"work-one","kind":"meeting"},"parts":[{"manifest":{"complete":true},"page":{"data":{"content":"`, `"},"complete":true}}]}`},
		{"context", "context get", `{"read_scope":{"schema_version":"riido-read-scope.v1","resources":[]},"document":{"content":"`, `"}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := json.RawMessage(tc.prefix + strings.Repeat("본문", 1000) + tc.suffix)
			fetcher := &fakeFetcher{result: Result{Data: body, Sources: []ResultSource{{Kind: "task", ID: "work-one", Version: "v1"}}}}
			broker := New("", "")
			broker.Fetcher = fetcher
			req := request(issueForTest(t, broker, saasBinding()), tc.command)
			req.Args = map[string]string{"work": "work-one", "cursor": "page-one", "budget_chars": "350"}
			env := broker.Handle(req, activeSet("run-1"))
			got, err := json.Marshal(env.Data)
			if env.Error != nil || err != nil || !bytes.Equal(got, body) {
				t.Fatal("budget changed versioned captured content or made it unavailable")
			}
			if len(env.Sources) != 1 || env.Sources[0].Truncated || env.Sources[0].Status != contextcli.SourceAvailable {
				t.Fatal("complete delivered source was reported as shortened")
			}
			if len(env.Warnings) == 0 || env.Warnings[len(env.Warnings)-1].Code != "BUDGET_PRESERVED_PAGE" {
				t.Fatal("over-budget captured response needs an explicit preservation warning")
			}
			if tc.name == "part" && (env.Page == nil || env.Page.NextCursor != "page-two") {
				t.Fatal("captured next cursor changed")
			}
		})
	}
}

func TestCapturedBudgetWithinLimitHasNoBudgetWarning(t *testing.T) {
	broker := New("", "")
	broker.Fetcher = &fakeFetcher{result: Result{Data: json.RawMessage(`{"source_id":"work-one","part":"editor","data":{"content":"full"},"complete":true}`)}}
	req := request(issueForTest(t, broker, saasBinding()), "context part")
	req.Args = map[string]string{"work": "work-one", "cursor": "page-one", "budget_chars": "1000"}
	env := broker.Handle(req, activeSet("run-1"))
	if env.Error != nil || len(env.Warnings) != 1 || env.Warnings[0].Code != "ASSIGNMENT_SNAPSHOT" {
		t.Fatal("a fitting captured page must not produce a budget warning")
	}
}

func TestScopedCommentsDoNotAdvertiseLegacyCursorPagination(t *testing.T) {
	broker := New("", "")
	broker.Fetcher = &fakeFetcher{result: Result{Data: json.RawMessage(`{"resource":{"id":"work-one","kind":"component"},"parts":[{"manifest":{"part":"comments","complete":true,"cursors":["first","next"],"next_action":"riido context part work-one --cursor next"}}]}`)}}
	req := request(issueForTest(t, broker, saasBinding()), "work comments")
	env := broker.Handle(req, activeSet("run-1"))
	if env.Error != nil || env.Page != nil {
		t.Fatal("scoped inline comments must use their content manifest, not a complete legacy page")
	}
}

func TestInvalidCapturedCursorReturnsUsageWithoutServerDetails(t *testing.T) {
	broker := New("", "")
	broker.Fetcher = &fakeFetcher{err: &FetchError{Status: http.StatusBadRequest, Message: "private cursor details must not be echoed"}}
	req := request(issueForTest(t, broker, saasBinding()), "work comments")
	req.Args = map[string]string{"cursor": "wrong-kind"}
	env := broker.Handle(req, activeSet("run-1"))
	if env.Error == nil || env.Error.Code != contextcli.CodeUsage {
		t.Fatal("invalid cursor must be actionable input failure instead of a retryable outage")
	}
	if strings.Contains(env.Error.Message, "private cursor") || !strings.Contains(env.Error.NextAction, "context part") {
		t.Fatal("invalid cursor must give safe continuation guidance")
	}
}
