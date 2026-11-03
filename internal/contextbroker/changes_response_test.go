package contextbroker

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

const changedFixture = `{"snapshot_id":"asg-1","since":"previous","changes":[{"resource_kind":"component","source_id":"task-1","part":"editor","status":"updated","title":"현재 제목","team_id":"team-1","key_number":"42","reasons":[{"kind":"root","relation":null}],"current":{"resource_kind":"component","source_id":"task-1","part":"editor","version":"v1","digest":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","status":"available","complete":true,"cursors":["opaque.first","opaque.second"]}}],"complete":true,"full_refresh_required":false}`

func readChangesTest(t *testing.T, result Result, budget string) contextcli.Envelope {
	t.Helper()
	b := New("", "")
	b.Fetcher = &fakeFetcher{result: result}
	req := request(issueForTest(t, b, saasBinding()), "context changes")
	req.Args = map[string]string{"since": "previous"}
	if budget != "" {
		req.Args["budget_chars"] = budget
	}
	return b.Handle(req, activeSet("run-1"))
}

func changedResult() Result {
	return Result{Data: json.RawMessage(changedFixture), CapturedAt: time.Now(), Sources: []ResultSource{{Kind: "task", ID: "task-1", Part: "title"}}}
}

func TestChangesProjectsMetadataWithoutBodies(t *testing.T) {
	r := changedResult()
	r.Data = json.RawMessage(strings.ReplaceAll(changedFixture, `"resource_kind":`, `"body":"DO-NOT-OUTPUT","resource_kind":`))
	env := readChangesTest(t, r, "")
	raw, _ := json.Marshal(env)
	if env.Error != nil || strings.Contains(string(raw), "DO-NOT-OUTPUT") || len(env.Sources) != 1 {
		t.Fatal("unsafe comparison projection")
	}
	for _, want := range []string{"opaque.first", "opaque.second", "sha256:aaaa", "현재 제목", "task-1"} {
		if !strings.Contains(string(raw), want) {
			t.Fatal("metadata lost", want)
		}
	}
}

func TestChangesBudgetNeverShortensComparison(t *testing.T) {
	r := changedResult()
	data, err := changesMetadata(r, Query{AssignmentID: "asg-1", Since: "previous"})
	if err != nil {
		t.Fatal(err)
	}
	for _, budget := range []int{1, utf8.RuneCount(data) - 1, utf8.RuneCount(data)} {
		env := readChangesTest(t, r, strconv.Itoa(budget))
		if budget < utf8.RuneCount(data) {
			if env.Error == nil || env.Error.Code.ExitCode() != contextcli.ExitIncomplete || env.Error.NextAction != "riido context get" || env.Error.Retryable || env.Data != nil || len(env.Sources) != 0 {
				t.Fatal("budget did not fail closed", env.Error)
			}
		} else if env.Error != nil {
			t.Fatal("exact budget rejected", env.Error.Code)
		}
	}
}

func TestChangesPartialAndMissingDoNotInventDeletion(t *testing.T) {
	for _, changes := range []string{`[]`, `[{"resource_kind":"meeting","source_id":"old","part":"board","status":"unavailable","reason":"not_in_current_capture"}]`, strings.Replace(strings.Split(strings.Split(changedFixture, `"changes":`)[1], `,"complete":true,"full_refresh_required"`)[0], `"status":"updated"`, `"status":"unavailable","reason":"incomplete_comparison"`, 1)} {
		r := Result{CapturedAt: time.Now(), Data: json.RawMessage(`{"snapshot_id":"asg-1","since":"previous","changes":` + changes + `,"complete":false,"full_refresh_required":true,"next_action":"untrusted command"}`)}
		env := readChangesTest(t, r, "")
		raw, _ := json.Marshal(env.Data)
		if env.Error != nil || !strings.Contains(string(raw), `"complete":false`) || !strings.Contains(string(raw), `"next_action":"riido context get"`) || strings.Contains(string(raw), "untrusted command") {
			t.Fatal("partial recovery lost")
		}
	}
}

func TestChangesRejectsMalformedIdentityAndAudit(t *testing.T) {
	cases := map[string]func(*Result){
		"snapshot": func(r *Result) { r.Data = json.RawMessage(strings.Replace(changedFixture, "asg-1", "other", 1)) },
		"baseline": func(r *Result) { r.Data = json.RawMessage(strings.Replace(changedFixture, "previous", "other", 1)) },
		"current identity": func(r *Result) {
			r.Data = json.RawMessage(strings.Replace(changedFixture, `"version":"v1"`, `"version":"v1","source_id":"other"`, 1))
		},
		"missing bool": func(r *Result) {
			r.Data = json.RawMessage(strings.Replace(changedFixture, `,"full_refresh_required":false`, "", 1))
		},
		"null changes": func(r *Result) {
			r.Data = json.RawMessage(strings.Replace(changesFixture, `"changes":[]`, `"changes":null`, 1))
		},
		"missing changes":  func(r *Result) { r.Data = json.RawMessage(strings.Replace(changesFixture, `"changes":[],`, "", 1)) },
		"complete partial": func(r *Result) { r.Truncated = true },
		"live":             func(r *Result) { r.Live = true },
		"time":             func(r *Result) { r.CapturedAt = time.Time{} },
		"digest":           func(r *Result) { r.Data = json.RawMessage(strings.Replace(changedFixture, "sha256:a", "sha256:z", 1)) },
		"body audit":       func(r *Result) { r.Sources[0].Part = "body" },
		"other source":     func(r *Result) { r.Sources[0].ID = "other" },
		"wrong kind":       func(r *Result) { r.Sources[0].Kind = "related_meeting" },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			r := changedResult()
			mutate(&r)
			env := readChangesTest(t, r, "")
			if env.Error == nil || env.Error.Code != contextcli.CodeUnavailable || env.Data != nil || len(env.Sources) != 0 {
				t.Fatal("untrusted response accepted")
			}
		})
	}
}

func TestChangesUncapturedPartKeepsFullRefreshRecovery(t *testing.T) {
	r := Result{CapturedAt: time.Now(), Data: json.RawMessage(`{"snapshot_id":"asg-1","since":"previous","changes":[{"resource_kind":"meeting","source_id":"meeting-1","part":"board","status":"unavailable","reason":"incomplete_comparison","title":"미팅","current":{"resource_kind":"meeting","source_id":"meeting-1","part":"board","version":"","status":"unavailable","complete":false,"cursors":[]}}],"complete":false,"full_refresh_required":true,"next_action":"riido context get"}`)}
	env := readChangesTest(t, r, "")
	raw, _ := json.Marshal(env.Data)
	if env.Error != nil || !strings.Contains(string(raw), `"full_refresh_required":true`) || !strings.Contains(string(raw), `"version":""`) {
		t.Fatal("valid uncaptured part lost recovery")
	}
}

func TestChangesMalformedCursorsCannotClaimComplete(t *testing.T) {
	for _, cursors := range []string{`null`, `[]`, `[""]`, `["same","same"]`} {
		r := changedResult()
		r.Data = json.RawMessage(strings.Replace(changedFixture, `["opaque.first","opaque.second"]`, cursors, 1))
		env := readChangesTest(t, r, "")
		if env.Error == nil || env.Data != nil {
			t.Fatal("complete comparison accepted unusable cursors")
		}
	}
}
