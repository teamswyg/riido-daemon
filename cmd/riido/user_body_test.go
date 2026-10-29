package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
	"github.com/teamswyg/riido-daemon/internal/userapi"
)

func TestUserBodyCallsVersionBoundAPIWithoutDaemon(t *testing.T) {
	version := "sha256:" + strings.Repeat("a", 64)
	calls := 0
	service := commentService(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		var req struct {
			Query     string
			Variables userapi.ReadBodyRequest
		}
		if json.NewDecoder(r.Body).Decode(&req) != nil || r.Header.Get("Authorization") != "Bearer fixture-access" || !strings.Contains(req.Query, "getComponentReadScopeBody") {
			t.Fatal("expected authenticated fixed body query")
		}
		want := userapi.ReadBodyRequest{ComponentID: "work-one", TeamID: "team-one", Part: "description", Version: version, Offset: 12, Limit: 64}
		if req.Variables != want {
			t.Fatalf("body request = %+v", req.Variables)
		}
		fmt.Fprintf(w, `{"data":{"getComponentReadScopeBody":{"resourceId":"work-one","part":"description","status":"available","source":"component_description","format":"html","version":%q,"offset":12,"content":"다음 페이지","nextOffset":null,"complete":true}}}`, version)
	})
	h := newContextHarness(t, nil)
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
	if err := runWork(h.io, []string{"body", "work-one", "--team-id", "team-one", "--part", "description", "--version", version, "--offset", "12", "--limit", "64", "--json"}); err != nil {
		t.Fatal(err)
	}
	data := decodeEnvelope(t, h.stdout.Bytes()).Data.(map[string]any)
	body := data["body"].(map[string]any)
	if calls != 1 || data["mode"] != "user" || body["content"] != "다음 페이지" || body["complete"] != true || body["nextOffset"] != nil || len(h.daemon.requests) != 0 {
		t.Fatal("body page or mode boundary was lost")
	}
}

func TestUserBodyRejectsInvalidPagesBeforeCredentialRead(t *testing.T) {
	version := "sha256:" + strings.Repeat("a", 64)
	for _, flags := range [][]string{
		{"--version", ""},
		{"--version", "not-a-version"},
		{"--offset", "-1"},
		{"--offset", "2147483648"},
		{"--offset", "1.5"},
		{"--limit", "0"},
		{"--limit", "65537"},
		{"--limit", ""},
		{"--part", "board"},
	} {
		h := newContextHarness(t, nil)
		h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) {
			t.Fatal("invalid body page opened credentials")
			return nil, nil
		}
		args := append([]string{"body", "work-one", "--team-id", "team-one", "--version", version, "--json"}, flags...)
		if code := exitCodeOf(t, runWork(h.io, args)); code != 2 {
			t.Fatalf("%v exit = %d", flags, code)
		}
		if len(h.daemon.requests) != 0 {
			t.Fatal("user page reached daemon")
		}
	}
}

func TestUserBodyConflictIsSanitizedAndAgentNeverUsesUserCredentials(t *testing.T) {
	version := "sha256:" + strings.Repeat("a", 64)
	service := commentService(t, func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"data":{"getComponentReadScopeBody":null},"errors":[{"message":"private-fixture-detail","extensions":{"code":"conflict"}}]}`)
	})
	h := newContextHarness(t, nil)
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
	args := []string{"body", "work-one", "--team-id", "team-one", "--version", version, "--json"}
	if code := exitCodeOf(t, runWork(h.io, args)); code != 10 {
		t.Fatalf("conflict exit = %d", code)
	}
	if strings.Contains(h.stdout.String()+h.stderr.String(), "private-fixture-detail") {
		t.Fatal("server detail leaked")
	}
	agent := newContextHarness(t, agentVars())
	agent.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) {
		t.Fatal("agent opened user credentials")
		return nil, nil
	}
	if code := exitCodeOf(t, runWork(agent.io, args)); code != 8 || len(agent.daemon.requests) != 0 {
		t.Fatal("agent used a live body read")
	}
}

func TestUserBodyInvalidServerOffsetIsUsageFailure(t *testing.T) {
	service := commentService(t, func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"errors":[{"message":"private-fixture-offset","extensions":{"code":"invalid_request"}}]}`)
	})
	h := newContextHarness(t, nil)
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
	args := []string{"body", "work-one", "--team-id", "team-one", "--version", "sha256:" + strings.Repeat("a", 64), "--offset", "1", "--json"}
	if code := exitCodeOf(t, runWork(h.io, args)); code != 2 {
		t.Fatalf("invalid server byte boundary exit=%d", code)
	}
	if strings.Contains(h.stdout.String()+h.stderr.String(), "private-fixture-offset") {
		t.Fatal("server details leaked")
	}
}
