package main

import (
	"net/http"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

func TestContextCLINativeIPCRoundTripRateLimit(t *testing.T) {
	f := newNativeContextFixture(t)
	f.failureStatus.Store(http.StatusTooManyRequests)
	for _, args := range [][]string{{"context", "get"}, {"agents", "list"}} {
		env, exit := f.run(t, nil, args...)
		if exit != 7 || env.Error == nil || env.Error.Code != contextcli.CodeRateLimited || env.Error.Retryable {
			t.Fatalf("compiled CLI run limit: exit=%d error=%+v", exit, env.Error)
		}
		if env.Data != nil || len(env.Sources) != 0 || env.Error.RequestID != env.RequestID || !strings.Contains(env.Error.NextAction, "새 실행") {
			t.Fatal("rate limit must preserve request identity and new-run action without sources")
		}
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.queries) != 2 {
		t.Fatal("compiled CLI must not retry a rate limited read")
	}
	for _, line := range f.audit {
		f.assertNoSecrets(t, line)
	}
}
