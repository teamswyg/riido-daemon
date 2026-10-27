package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

func TestUserWorkGraphQLHTTP200ErrorsHaveAccurateExit(t *testing.T) {
	for code, want := range map[string]int{"authentication_required": 3, "permission_denied": 4, "not_found": 5, "unknown": 6} {
		t.Run(code, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.URL.Path != "/graphql" {
					t.Error("GraphQL domain error attempted token refresh")
				}
				fmt.Fprintf(w, `{"errors":[{"message":"fake-secret-provider-message","extensions":{"code":%q}}]}`, code)
			}))
			defer server.Close()
			store := &cliMemoryStore{present: true, credential: userauth.Credential{SessionID: "test-session", AccessToken: "fake-access", RefreshToken: "fake-refresh", ExpiresAt: time.Now().Add(time.Hour), TokenType: "Bearer"}}
			service, err := standalonecli.New(standalonecli.Config{APIOrigin: server.URL, Store: store})
			if err != nil {
				t.Fatal(err)
			}
			h := newContextHarness(t, nil)
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
			got := exitCodeOf(t, runWork(h.io, []string{"get", "work-one", "--team-id", "team-one", "--json"}))
			if got != want || calls != 1 {
				t.Fatalf("domain error exit=%d calls=%d", got, calls)
			}
			if strings.Contains(h.stdout.String()+h.stderr.String(), "fake-secret-provider-message") {
				t.Fatal("server message leaked")
			}
		})
	}
}
