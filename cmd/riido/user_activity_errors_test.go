package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
)

func TestUserActivityErrorsKeepSanitizedExitCodes(t *testing.T) {
	for _, tc := range []struct {
		name, code   string
		status, exit int
	}{
		{"permission-http", "", 403, 4},
		{"rate", "", 429, 7},
		{"server", "", 503, 6},
		{"permission-graphql", "permission_denied", 200, 4},
		{"missing", "not_found", 200, 5},
		{"authentication", "authentication_required", 200, 3},
		{"unknown", "private-code", 200, 6},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			service := commentService(t, func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.URL.Path != "/graphql" {
					t.Error("error unexpectedly refreshed")
				}
				w.WriteHeader(tc.status)
				fmt.Fprintf(w, `{"data":{"getComponentActivityLogs":{"activityLogs":[],"private":"private-body"}},"errors":[{"message":"private-detail","extensions":{"code":%q}}]}`, tc.code)
			})
			h := newContextHarness(t, nil)
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
			if got := exitCodeOf(t, runWork(h.io, activityArgs())); got != tc.exit {
				t.Fatalf("exit=%d want=%d", got, tc.exit)
			}
			env := decodeEnvelope(t, h.stdout.Bytes())
			if calls != 1 || env.Data != nil || env.Error == nil || strings.Contains(h.stdout.String()+h.stderr.String(), "private-") {
				t.Fatal("error leaked data or retried")
			}
		})
	}
}

func TestUserActivityRefreshRepeatsIdenticalRead(t *testing.T) {
	reads, refreshes := 0, 0
	var original string
	service := commentService(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			refreshes++
			fmt.Fprint(w, `{"access_token":"refreshed-access","refresh_token":"refreshed-refresh","token_type":"Bearer","expires_in":3600,"scope":"read"}`)
			return
		}
		reads++
		raw, _ := io.ReadAll(r.Body)
		if reads == 1 {
			original = string(raw)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if original != string(raw) || r.Header.Get("Authorization") != "Bearer refreshed-access" {
			t.Error("refresh changed read")
		}
		fmt.Fprint(w, `{"data":{"getComponentActivityLogs":{"activityLogs":[],"meta":{"total":0,"currentPage":1,"nextPage":null,"prevPage":null,"lastPage":0}}}}`)
	})
	h := newContextHarness(t, nil)
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
	if err := runWork(h.io, activityArgs()); err != nil {
		t.Fatal(err)
	}
	if reads != 2 || refreshes != 1 || len(h.daemon.requests) != 0 {
		t.Fatalf("reads=%d refreshes=%d", reads, refreshes)
	}
}
