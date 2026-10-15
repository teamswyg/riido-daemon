package session

import (
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/provider/claude"
	"github.com/teamswyg/riido-daemon/internal/provider/cursor"
)

func TestNativeSessionIdentityGuardBeforePublication(t *testing.T) {
	for _, test := range []struct {
		name, resume string
		pins         []string
		want         string
		blocked      bool
	}{
		{"new pin", "", []string{"original", "original", ""}, "original", false},
		{"resumed pin", "original", []string{"original", "original"}, "original", false},
		{"resume mismatch", "original", []string{"foreign"}, "original", true},
		{"pin changed", "", []string{"original", "foreign"}, "original", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			adapter := &recordingAdapter{name: "fake", parser: &recordingParser{}, translateFn: func(agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
				var events []agentbridge.Event
				for _, pin := range test.pins {
					events = append(events, agentbridge.Event{Kind: agentbridge.EventSessionIdentified, SessionID: pin})
				}
				events = append(events, agentbridge.Event{Kind: agentbridge.EventResult, Result: completedResult()})
				return events, nil, nil
			}}
			scenario := startToolGateScenario(t, "identity", adapter, func(cfg *Config) { cfg.Request.ResumeSessionID = test.resume })
			go scenario.running.EmitStdout([]byte("x"))
			res := waitResult(t, scenario.session, time.Second)
			if res.SessionID != test.want || (res.Status == agentbridge.ResultBlocked) != test.blocked {
				t.Fatalf("result=%+v", res)
			}
			pins := 0
			for ev := range scenario.session.Events() {
				if ev.Kind == agentbridge.EventSessionIdentified {
					pins++
					if ev.SessionID != test.want {
						t.Fatalf("foreign pin published: %+v", ev)
					}
				}
				if strings.Contains(ev.Err, "foreign") {
					t.Fatal("foreign identity exposed")
				}
			}
			if pins > 1 {
				t.Fatalf("duplicate pins published: %d", pins)
			}
		})
	}
}

func TestResumeIdentitySurvivesEarlyProviderExit(t *testing.T) {
	adapter := &recordingAdapter{name: "fake", parser: &recordingParser{}}
	scenario := startToolGateScenario(t, "early-exit", adapter, func(cfg *Config) { cfg.Request.ResumeSessionID = "original" })
	scenario.running.EmitExit(1, nil)
	res := waitResult(t, scenario.session, time.Second)
	if res.SessionID != "original" || res.Status == agentbridge.ResultCompleted {
		t.Fatalf("result=%+v", res)
	}
}

func TestNativeResultSessionIdentityIsChecked(t *testing.T) {
	for _, provider := range []string{"claude", "cursor"} {
		for _, resume := range []string{"", "expected"} {
			t.Run(provider+"/"+resume, func(t *testing.T) {
				adapter := &recordingAdapter{name: "fake", parser: &recordingParser{}, translateFn: func(agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
					raw := agentbridge.RawEvent{Type: "result", Source: agentbridge.RawSourceStdout, Payload: map[string]any{"session_id": "native", "subtype": "success", "result": "reply"}}
					if provider == "claude" {
						return claude.Translate(raw)
					}
					return cursor.Translate(raw)
				}}
				scenario := startToolGateScenario(t, "result-session", adapter, func(cfg *Config) { cfg.Request.ResumeSessionID = resume })
				go scenario.running.EmitStdout([]byte("x"))
				res := waitResult(t, scenario.session, time.Second)
				if resume == "" {
					if res.SessionID != "native" || res.Status != agentbridge.ResultCompleted {
						t.Fatalf("lost result-only pin: %+v", res)
					}
				} else if res.SessionID != "expected" || res.Status != agentbridge.ResultBlocked || res.Output != "" {
					t.Fatalf("foreign result accepted: %+v", res)
				}
			})
		}
	}
}

func TestNativeSessionGuardPreservesTerminalUsageTail(t *testing.T) {
	adapter := &recordingAdapter{name: "fake", parser: &recordingParser{}, translateFn: func(agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
		return []agentbridge.Event{
			{Kind: agentbridge.EventSessionIdentified, SessionID: "session"},
			{Kind: agentbridge.EventResult, Result: completedResult()},
			{Kind: agentbridge.EventUsageDelta, Usage: agentbridge.Usage{PromptTokens: 7}},
		}, nil, nil
	}}
	scenario := startToolGateScenario(t, "usage-tail", adapter, nil)
	go scenario.running.EmitStdout([]byte("x"))
	res := waitResult(t, scenario.session, time.Second)
	if res.Usage.PromptTokens != 7 || res.SessionID != "session" || res.Status != agentbridge.ResultCompleted {
		t.Fatalf("terminal usage lost: %+v", res)
	}
}

func TestNativeResumeCannotCompleteWithoutIdentityConfirmation(t *testing.T) {
	adapter := &recordingAdapter{name: "fake", parser: &recordingParser{}, translateFn: func(agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
		return []agentbridge.Event{{Kind: agentbridge.EventResult, Result: agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "unverified reply"}}}, nil, nil
	}}
	scenario := startToolGateScenario(t, "unconfirmed", adapter, func(cfg *Config) { cfg.Request.ResumeSessionID = "expected" })
	go scenario.running.EmitStdout([]byte("x"))
	res := waitResult(t, scenario.session, time.Second)
	if res.Status != agentbridge.ResultBlocked || res.SessionID != "expected" || res.Output != "" {
		t.Fatalf("unconfirmed resume succeeded: %+v", res)
	}
}
