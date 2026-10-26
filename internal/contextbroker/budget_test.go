package contextbroker

import (
	"encoding/json"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

func TestApplyBudgetShrinksLongestBodiesAndStaysValid(t *testing.T) {
	data := json.RawMessage(`{"document":{"content":"` + strings.Repeat("본문", 2000) + `"},"title":"짧은 제목","n":3}`)
	out, cut, fits := applyBudget(data, 600)
	if !cut || !fits || utf8.RuneCount(out) > 600 || !json.Valid(out) || !strings.Contains(string(out), "짧은 제목") || !strings.Contains(string(out), "…") {
		t.Fatalf("cut=%v fits=%v len=%d out=%s", cut, fits, utf8.RuneCount(out), out[:min(len(out), 120)])
	}
	if same, cut, fits := applyBudget(data, 0); cut || !fits || string(same) != string(data) {
		t.Fatal("no budget must leave data untouched")
	}
}

func TestReadHonorsBudgetAndKeepsCapabilityOnInconclusiveActivity(t *testing.T) {
	b := New("", "")
	b.Fetcher = &fakeFetcher{result: Result{Data: json.RawMessage(`{"document":{"content":"` + strings.Repeat("x", 5000) + `"}}`), Sources: []ResultSource{{Kind: "task", ID: "t"}}}}
	token := issueForTest(t, b, saasBinding())
	req := request(token, "context get")
	req.Args = map[string]string{"budget_chars": "300"}
	env := b.Handle(req, activeSet("run-1"))
	raw, _ := json.Marshal(env.Data)
	if env.Error != nil || len(raw) > 400 || !env.Sources[0].Truncated || env.Warnings[len(env.Warnings)-1].Code != "BUDGET_TRUNCATED" {
		t.Fatalf("budget not honored: %d bytes warnings=%+v", len(raw), env.Warnings)
	}
	unknown := func(string) (bool, bool) { return false, false }
	if got := errorCode(t, b.Handle(request(token, "whoami"), unknown)); got != contextcli.CodeUnavailable {
		t.Fatalf("inconclusive activity must be retryable, got %s", got)
	}
	if got := b.Handle(request(token, "whoami"), activeSet("run-1")); got.Error != nil {
		t.Fatalf("capability must survive an inconclusive check: %+v", got.Error)
	}
}
