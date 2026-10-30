package contextbroker

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

type sequenceFetcher struct {
	results []Result
	errs    []error
	queries []Query
}

func (f *sequenceFetcher) FetchAssignmentContext(_ context.Context, q Query) (Result, error) {
	i := len(f.queries)
	f.queries = append(f.queries, q)
	if i >= len(f.results) {
		return Result{}, &FetchError{Status: http.StatusInternalServerError, Message: "unexpected"}
	}
	return f.results[i], f.errs[i]
}

func contentPartRequest(t *testing.T, b *Broker, args map[string]string) contextcli.Envelope {
	t.Helper()
	req := request(issueForTest(t, b, saasBinding()), "context part")
	req.Args = args
	return b.Handle(req, activeSet("run-1"))
}

const batchData = `{"source_id":"work-one","part":"description","version":"v1","pages":[{"cursor":"page-one","data":{"content":"ab"}},{"cursor":"page-two","data":{"content":"cd"}}],"complete":false,"next_cursor":"page-three"}`

func TestContentReadRequestsConsecutivePagesByDefault(t *testing.T) {
	fetcher := &sequenceFetcher{results: []Result{{Data: json.RawMessage(batchData)}}, errs: []error{nil}}
	b := New("", "")
	b.Fetcher = fetcher
	env := contentPartRequest(t, b, map[string]string{"work": "work-one", "cursor": "page-one"})
	if env.Error != nil || len(fetcher.queries) != 1 || fetcher.queries[0].MaxBytes != 48<<10 {
		t.Fatalf("err=%+v queries=%+v", env.Error, fetcher.queries)
	}
	if env.Page == nil || env.Page.NextCursor != "page-three" || env.Page.Limit != 2 || !env.Page.Truncated {
		t.Fatalf("batch continuation lost: %+v", env.Page)
	}
	for _, warning := range env.Warnings {
		if warning.Code == "BUDGET_PRESERVED_PAGE" {
			t.Fatal("a batch sized by the budget must not warn about preserving it")
		}
	}
}

func TestContentReadBudgetSizesTheBatchWithinServerRange(t *testing.T) {
	for budget, want := range map[string]int{"1000": 8 << 10, "100000": 100000, "999999": 256 << 10} {
		fetcher := &sequenceFetcher{results: []Result{{Data: json.RawMessage(batchData)}}, errs: []error{nil}}
		b := New("", "")
		b.Fetcher = fetcher
		contentPartRequest(t, b, map[string]string{"work": "work-one", "cursor": "page-one", "budget_chars": budget})
		if len(fetcher.queries) != 1 || fetcher.queries[0].MaxBytes != want {
			t.Fatalf("budget %s → %+v", budget, fetcher.queries)
		}
	}
}

func TestContentReadFallsBackToSinglePageOnOlderControlPlane(t *testing.T) {
	fetcher := &sequenceFetcher{
		results: []Result{{}, {Data: json.RawMessage(`{"source_id":"work-one","cursor":"page-one","data":{"content":"ab"},"next_cursor":"page-two"}`)}},
		errs:    []error{&FetchError{Status: http.StatusBadRequest, Message: `decode json: unknown field "max_bytes"`}, nil},
	}
	b := New("", "")
	b.Fetcher = fetcher
	env := contentPartRequest(t, b, map[string]string{"work": "work-one", "cursor": "page-one"})
	if env.Error != nil || len(fetcher.queries) != 2 || fetcher.queries[1].MaxBytes != 0 || fetcher.queries[1].Cursor != "page-one" {
		t.Fatalf("err=%+v queries=%+v", env.Error, fetcher.queries)
	}
	if env.Page == nil || env.Page.NextCursor != "page-two" || env.Page.Limit != 1 {
		t.Fatalf("single-page continuation lost: %+v", env.Page)
	}
}

func TestContentReadDoesNotResendOtherRefusals(t *testing.T) {
	fetcher := &sequenceFetcher{results: []Result{{}}, errs: []error{&FetchError{Status: http.StatusBadRequest, Message: "assignment_id, lease_token and a supported view are required"}}}
	b := New("", "")
	b.Fetcher = fetcher
	env := contentPartRequest(t, b, map[string]string{"work": "work-one", "cursor": "page-one"})
	if env.Error == nil || len(fetcher.queries) != 1 {
		t.Fatalf("err=%+v queries=%d", env.Error, len(fetcher.queries))
	}
}
