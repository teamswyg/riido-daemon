package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
)

func repositoryHTTPFixture(t *testing.T, deferred bool, status int) (*standalonecli.Service, *[]string) {
	t.Helper()
	service, calls, _ := repositoryHTTPFixtureWithPadding(t, deferred, status, strings.Repeat("한", 45000))
	return service, calls
}

func repositoryHTTPFixtureWithPadding(t *testing.T, deferred bool, status int, padding string) (*standalonecli.Service, *[]string, string) {
	t.Helper()
	metadata := map[string]any{
		"details": map[string]any{"id": "work-one"}, "unselected": padding,
		"repositories": []any{map[string]any{"id": "repo-one", "fullName": "team/private", "isPrivate": true, "repositoryUrl": "https://github.example/team/private", "source": "connected"}},
		"github":       []any{map[string]any{"id": "pr-one", "title": "제목\x1b[31m\n가짜 줄", "number": 1, "state": "open", "htmlUrl": "https://github.example/team/private/pull/1", "updatedAt": nil, "commentsCount": 0, "aiShortSummary": ""}},
	}
	return metadataHTTPFixture(t, "repositories", []string{"repositories", "github"}, metadata, deferred, status)
}

func TestUserRepositoriesActualHTTPInlineAndDeferred(t *testing.T) {
	for _, deferred := range []bool{false, true} {
		t.Run(fmt.Sprint(deferred), func(t *testing.T) {
			service, calls := repositoryHTTPFixture(t, deferred, 200)
			h := newContextHarness(t, nil)
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
			if err := runRepo(h.io, []string{"list", "--work", "work-one", "--team-id", "team-one", "--json"}); err != nil {
				t.Fatal(err)
			}
			env := decodeEnvelope(t, h.stdout.Bytes())
			raw, _ := json.Marshal(env.Data)
			var result struct {
				LocalAuthentication struct {
					Used         bool
					Availability string
				} `json:"local_repository_authentication"`
				Mode string
				Live bool
				Repo struct {
					Complete     bool
					Repositories []struct {
						FullName  string
						IsPrivate bool
					}
					PullRequests    []struct{ Title string }
					MetadataVersion *string
				}
			}
			if json.Unmarshal(raw, &result) != nil || env.Error != nil || result.LocalAuthentication.Used || result.LocalAuthentication.Availability != "not_checked" || result.Mode != "user" || !result.Live || !result.Repo.Complete || len(result.Repo.Repositories) != 1 || !result.Repo.Repositories[0].IsPrivate || len(result.Repo.PullRequests) != 1 || result.Repo.PullRequests[0].Title != "제목\x1b[31m\n가짜 줄" {
				t.Fatalf("bad output: %s", raw)
			}
			want := []string{"repositories"}
			if deferred {
				want = append(want, "metadata", "metadata", "metadata")
			}
			if !reflect.DeepEqual(*calls, want) || len(h.daemon.requests) != 0 {
				t.Fatalf("calls=%v IPC=%d", *calls, len(h.daemon.requests))
			}
			if (result.Repo.MetadataVersion != nil) != deferred || strings.Contains(h.stdout.String(), "fixture-access") {
				t.Fatal("version or secret output")
			}
		})
	}
}

func TestUserRepositoriesTextEscapesRemoteValues(t *testing.T) {
	service, _ := repositoryHTTPFixture(t, false, 200)
	h := newContextHarness(t, nil)
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
	if err := runRepo(h.io, []string{"list", "--work", "work-one", "--team-id", "team-one"}); err != nil {
		t.Fatal(err)
	}
	if strings.ContainsRune(h.stdout.String(), '\x1b') || strings.Contains(h.stdout.String(), "\n가짜 줄") || !strings.Contains(h.stdout.String(), "github.example") {
		t.Fatal("unsafe/lost text")
	}
}

func TestUserRepositoriesContinuationPermissionAndRateErrors(t *testing.T) {
	for _, tc := range []struct{ status, exit int }{{403, 4}, {429, 7}} {
		t.Run(fmt.Sprint(tc.status), func(t *testing.T) {
			service, calls := repositoryHTTPFixture(t, true, tc.status)
			h := newContextHarness(t, nil)
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
			if got := exitCodeOf(t, runRepo(h.io, []string{"list", "--work", "work-one", "--team-id", "team-one", "--json"})); got != tc.exit {
				t.Fatalf("exit=%d", got)
			}
			env := decodeEnvelope(t, h.stdout.Bytes())
			if env.Data != nil || env.Error == nil || len(*calls) != 2 || strings.Contains(h.stdout.String(), "private-error") {
				t.Fatal("partial output or retry escaped")
			}
		})
	}
}

func TestUserRepositoriesModeAndArgumentsBeforeCredentials(t *testing.T) {
	for _, tc := range []struct {
		name string
		vars map[string]string
		args []string
		exit int
	}{
		{"agent", agentVars(), []string{"list", "--work", "RIID-1"}, 8},
		{"missing", nil, []string{"list"}, 2},
		{"key", nil, []string{"list", "--work", "RIID-1", "--team-id", "team-one"}, 8},
		{"cursor", nil, []string{"list", "--work", "work-one", "--team-id", "team-one", "--cursor", "x"}, 8},
		{"unknown", nil, []string{"get"}, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := newContextHarness(t, tc.vars)
			opened := false
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { opened = true; return nil, nil }
			if got := exitCodeOf(t, runRepo(h.io, append(tc.args, "--json"))); got != tc.exit || opened || len(h.daemon.requests) != 0 {
				t.Fatalf("exit=%d opened=%v", got, opened)
			}
		})
	}
	h := newContextHarness(t, nil)
	if err := runRepo(h.io, []string{"list", "--help"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(h.stderr.String(), "riido repo list --work <ID> --team-id <TEAM-ID>") {
		t.Fatal("missing help")
	}
	found := false
	for _, c := range contextcli.Capabilities("fixture").Commands {
		if c.Name == "repo list" {
			found = true
			if c.Status != contextcli.CommandAvailable || !reflect.DeepEqual(c.Modes, []contextcli.Mode{contextcli.ModeUser}) {
				t.Fatal("wrong modes")
			}
		}
	}
	if !found {
		t.Fatal("missing capability")
	}
}

func TestUserRepositoriesLimitPreservesExecutableOriginalVersion(t *testing.T) {
	service, calls, version := repositoryHTTPFixtureWithPadding(t, true, http.StatusOK, strings.Repeat("x", 16<<20))
	h := newContextHarness(t, nil)
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
	err := runRepo(h.io, []string{"list", "--work", "work-one", "--team-id", "team-one", "--json"})
	if got := exitCodeOf(t, err); got != 9 {
		t.Fatalf("exit=%d", got)
	}
	env := decodeEnvelope(t, h.stdout.Bytes())
	want := fmt.Sprintf("riido work body %q --team-id %q --part metadata --version %q --offset 0 --limit 65536 --json", "work-one", "team-one", version)
	if env.Error == nil || env.Error.NextAction != want || env.Data != nil {
		t.Fatalf("manual paging lost original version: %+v", env.Error)
	}
	// Execute the advertised continuation using the same authenticated HTTP API.
	continuation := newContextHarness(t, nil)
	continuation.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
	before := len(*calls)
	if err := runWork(continuation.io, []string{"body", "work-one", "--team-id", "team-one", "--part", "metadata", "--version", version, "--offset", "0", "--limit", "65536", "--json"}); err != nil {
		t.Fatal(err)
	}
	pageEnvelope := decodeEnvelope(t, continuation.stdout.Bytes())
	raw, _ := json.Marshal(pageEnvelope.Data)
	var data struct {
		Body struct {
			Version    string
			Offset     int
			NextOffset *int
			Complete   bool
		}
	}
	if json.Unmarshal(raw, &data) != nil || data.Body.Version != version || data.Body.Offset != 0 || data.Body.NextOffset == nil || data.Body.Complete || len(*calls) != before+1 {
		t.Fatal("advertised work body continuation failed")
	}
	for i, call := range *calls {
		if (i == 0 && call != "repositories") || (i > 0 && call != "metadata") {
			t.Fatal("unrelated read blocks fallback")
		}
	}
}
