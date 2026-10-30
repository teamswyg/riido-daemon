package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
)

func commentCommandArgs(command string) []string {
	args := []string{command, "work-one", "--team-id", "team-one", "--json"}
	if command == "replies" {
		args = append(args, "--comment-id", "comment-one")
	}
	return args
}

func TestUserCommentFailuresNeverExposePartialDataOrRetryPermission(t *testing.T) {
	for _, command := range []string{"comments", "replies"} {
		for _, tc := range []struct {
			name, code   string
			status, exit int
		}{
			{"forbidden", "", 403, 4},
			{"rate-limited", "", 429, 7},
			{"server", "", 500, 6},
			{"permission", "permission_denied", 200, 4},
			{"missing", "not_found", 200, 5},
			{"auth", "authentication_required", 200, 3},
			{"unknown", "private-unknown-code", 200, 6},
		} {
			t.Run(command+"/"+tc.name, func(t *testing.T) {
				calls := 0
				service := commentService(t, func(w http.ResponseWriter, r *http.Request) {
					calls++
					if r.URL.Path != "/graphql" {
						t.Error("domain/HTTP failure attempted refresh")
					}
					raw, _ := io.ReadAll(r.Body)
					if strings.Contains(string(raw), "RiidoCLICommentScope") {
						fmt.Fprint(w, `{"data":{"getComponent":{"id":"work-one"}}}`)
						return
					}
					w.Header().Set("Retry-After", "600")
					w.WriteHeader(tc.status)
					fmt.Fprintf(w, `{"data":{"private":"private-body"},"errors":[{"message":"private-detail","extensions":{"code":%q}}]}`, tc.code)
				})
				h := newContextHarness(t, nil)
				h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
				if got := exitCodeOf(t, runWork(h.io, commentCommandArgs(command))); got != tc.exit {
					t.Fatalf("exit=%d want=%d", got, tc.exit)
				}
				wantCalls := 1
				if command == "comments" {
					wantCalls = 2
				}
				envelope := decodeEnvelope(t, h.stdout.Bytes())
				if calls != wantCalls || envelope.Data != nil || envelope.Error == nil || strings.Contains(h.stdout.String()+h.stderr.String(), "private-") {
					t.Fatalf("partial/error leak or unexpected retry: calls=%d", calls)
				}
			})
		}
	}
}

func TestUserReadsHTTP429RetainsRateLimitedContract(t *testing.T) {
	for _, command := range []string{"whoami", "get"} {
		t.Run(command, func(t *testing.T) {
			calls := 0
			service := commentService(t, func(w http.ResponseWriter, _ *http.Request) {
				calls++
				w.Header().Set("Retry-After", "600")
				w.WriteHeader(http.StatusTooManyRequests)
			})
			h := newContextHarness(t, nil)
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
			var err error
			if command == "whoami" {
				err = runWhoami(h.io, []string{"--json"})
			} else {
				err = runWork(h.io, commentCommandArgs(command))
			}
			if got := exitCodeOf(t, err); got != 7 || calls != 1 {
				t.Fatalf("exit=%d calls=%d", got, calls)
			}
			envelope := decodeEnvelope(t, h.stdout.Bytes())
			if envelope.Error == nil || envelope.Error.Code != contextcli.CodeRateLimited || !envelope.Error.Retryable {
				t.Fatal("missing RATE_LIMITED error")
			}
		})
	}
}

func TestUserCommentRead401RefreshReplaysOnlyTheFailedQueryOnce(t *testing.T) {
	for _, command := range []string{"comments", "replies"} {
		t.Run(command, func(t *testing.T) {
			reads, refreshes, scopes := 0, 0, 0
			var firstBody string
			service := commentService(t, func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/oauth/token" {
					refreshes++
					fmt.Fprint(w, `{"access_token":"refreshed-access","refresh_token":"refreshed-refresh","token_type":"Bearer","expires_in":3600,"scope":"read","refresh_recovery":"v1"}`)
					return
				}
				raw, _ := io.ReadAll(r.Body)
				if strings.Contains(string(raw), "RiidoCLICommentScope") {
					scopes++
					fmt.Fprint(w, `{"data":{"getComponent":{"id":"work-one"}}}`)
					return
				}
				reads++
				if reads == 1 {
					firstBody = string(raw)
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				if string(raw) != firstBody || r.Header.Get("Authorization") != "Bearer refreshed-access" {
					t.Error("refresh changed the fixed query or did not replace auth")
				}
				if command == "comments" {
					fmt.Fprint(w, `{"data":{"getComponentComments":{"comments":[],"meta":{"total":0,"currentPage":1,"nextPage":null,"prevPage":null,"lastPage":0}}}}`)
				} else {
					fmt.Fprint(w, `{"data":{"getCommentReplies":[]}}`)
				}
			})
			h := newContextHarness(t, nil)
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
			if err := runWork(h.io, commentCommandArgs(command)); err != nil {
				t.Fatal(err)
			}
			wantScopes := 0
			if command == "comments" {
				wantScopes = 1
			}
			if reads != 2 || refreshes != 1 || scopes != wantScopes {
				t.Fatalf("reads=%d refreshes=%d scopes=%d", reads, refreshes, scopes)
			}
			var output map[string]any
			if json.Unmarshal(h.stdout.Bytes(), &output) != nil || strings.Contains(h.stdout.String()+h.stderr.String(), "refreshed-") {
				t.Fatal("invalid or unsanitized output")
			}
		})
	}
}
