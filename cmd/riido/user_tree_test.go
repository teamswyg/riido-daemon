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
	"github.com/teamswyg/riido-daemon/internal/userapi"
)

const userTreeFixture = `{"data":{"getComponentReadScope":{"schemaVersion":"riido-read-scope.v1","rootId":"work-one","teamId":"team-one","status":"available","selection":{"childrenLimit":10},"works":[{"id":"work-one","kind":"component","teamId":"team-one","reasons":[{"kind":"root"}],"summary":{"id":"work-one","kind":"component","title":"계층 작업\u001b[31m\n위조 줄","keyNumber":"RIID-5972","componentType":"task"}}]}}}`

func TestUserTreeAuthenticatedHTTPReceiptAndSafeText(t *testing.T) {
	for _, jsonOutput := range []bool{false, true} {
		t.Run(fmt.Sprint(jsonOutput), func(t *testing.T) {
			calls := 0
			service := commentService(t, func(w http.ResponseWriter, r *http.Request) {
				calls++
				var body struct {
					Query     string
					Variables map[string]string
				}
				if r.Method != http.MethodPost || r.URL.Path != "/graphql" || r.Header.Get("Authorization") != "Bearer fixture-access" || json.NewDecoder(r.Body).Decode(&body) != nil || !reflect.DeepEqual(body.Variables, map[string]string{"componentId": "work-one", "teamId": "team-one"}) || !strings.Contains(body.Query, "query RiidoCLIWorkTree") {
					t.Error("wrong authenticated read")
				}
				for _, forbidden := range []string{"mutation", "body", "details", "meetings", "contentPage"} {
					if strings.Contains(body.Query, forbidden) {
						t.Error("unexpected data selection")
					}
				}
				fmt.Fprint(w, userTreeFixture)
			})
			h := newContextHarness(t, nil)
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
			args := []string{"tree", "work-one", "--team-id", "team-one"}
			if jsonOutput {
				args = append(args, "--json")
			}
			if err := runWork(h.io, args); err != nil {
				t.Fatal(err)
			}
			if calls != 1 || len(h.daemon.requests) != 0 {
				t.Fatalf("HTTP=%d daemon=%d", calls, len(h.daemon.requests))
			}
			output := h.stdout.String()
			if strings.Contains(output, "fixture-access") || strings.ContainsRune(output, '\x1b') {
				t.Fatal("secret or terminal escape escaped")
			}
			if jsonOutput {
				env := decodeEnvelope(t, h.stdout.Bytes())
				raw, _ := json.Marshal(env.Data)
				var data struct {
					Mode string
					Live bool
					Tree userapi.WorkTree
				}
				_ = json.Unmarshal(raw, &data)
				if env.Error != nil || data.Mode != "user" || !data.Live || data.Tree.Root.Title != "계층 작업\x1b[31m\n위조 줄" || data.Tree.ChildrenLimit != 10 {
					t.Fatalf("unexpected result: %s", raw)
				}
			} else if !strings.Contains(output, "선택된 하위 작업: 0개 (최대 10개, 한 단계)") || strings.Contains(output, "\n위조 줄") {
				t.Fatal("unsafe or misleading text")
			}
		})
	}
}

func TestUserTreePartialWarningAndHTTPFailures(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
		exit   int
	}{
		{"partial", 200, strings.Replace(userTreeFixture, `"available"`, `"partial"`, 1), 0},
		{"permission", 403, `{"private":"hidden"}`, 4},
		{"authentication", 401, `{"error":"invalid_grant"}`, 3},
		{"rate-limit", 429, `{}`, 7},
		{"invalid", 200, `{"data":{"getComponentReadScope":null}}`, 6},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service := commentService(t, func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(tc.status); fmt.Fprint(w, tc.body) })
			h := newContextHarness(t, nil)
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
			err := runWork(h.io, []string{"tree", "work-one", "--team-id", "team-one", "--json"})
			if tc.exit == 0 {
				if err != nil {
					t.Fatal(err)
				}
			} else if got := exitCodeOf(t, err); got != tc.exit {
				t.Fatalf("exit=%d want=%d", got, tc.exit)
			}
			env := decodeEnvelope(t, h.stdout.Bytes())
			if tc.exit == 0 {
				if len(env.Warnings) != 1 || env.Warnings[0].Code != "CONTEXT_INCOMPLETE" {
					t.Fatal("missing partial warning")
				}
			} else if env.Data != nil || env.Error == nil {
				t.Fatal("failure returned partial data")
			}
			if len(h.daemon.requests) != 0 {
				t.Fatal("fallback to daemon")
			}
		})
	}
}

func TestUserTreeRejectsUnsupportedScopeBeforeCredentials(t *testing.T) {
	for _, flags := range [][]string{{"--workspace", "x"}, {"--cursor", "x"}, {"--page", "2"}, {"--budget", "1000"}, {"--limit", "20"}, {"--profile", "x"}} {
		t.Run(flags[0], func(t *testing.T) {
			h := newContextHarness(t, nil)
			opened := 0
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { opened++; return nil, userapi.ErrAuthentication }
			err := runWork(h.io, append([]string{"tree", "work-one", "--team-id", "team-one", "--json"}, flags...))
			if exitCodeOf(t, err) != 8 || opened != 0 || len(h.daemon.requests) != 0 {
				t.Fatal("unsupported option opened credential or daemon")
			}
		})
	}
	for _, tc := range []struct {
		args []string
		exit int
	}{{[]string{"tree", "work-one"}, 2}, {[]string{"tree", "RIID-5972", "--team-id", "team-one"}, 8}} {
		h := newContextHarness(t, nil)
		if got := exitCodeOf(t, runWork(h.io, append(tc.args, "--json"))); got != tc.exit {
			t.Fatalf("exit=%d want=%d", got, tc.exit)
		}
	}
}

func TestTreeAgentRouteAndAdvertisedUserCapability(t *testing.T) {
	h := newContextHarness(t, agentVars())
	h.daemon.replies[contextcli.IPCMethodContextRead] = `{"schema_version":"1","request_id":"daemon-1","data":{"fixture":true},"error":null}`
	if err := runWork(h.io, []string{"tree", "RIID-5972", "--json"}); err != nil {
		t.Fatal(err)
	}
	if h.daemon.lastRequest(t)["command"] != "work tree" {
		t.Fatal("agent route changed")
	}
	found := false
	for _, c := range contextcli.Capabilities("fixture").Commands {
		if c.Name == "work tree" {
			found = true
			if c.Status != contextcli.CommandAvailable || !reflect.DeepEqual(c.Modes, []contextcli.Mode{contextcli.ModeAgent, contextcli.ModeUser}) {
				t.Fatalf("capability=%+v", c)
			}
		}
	}
	if !found {
		t.Fatal("tree not advertised")
	}
	h = newContextHarness(t, nil)
	if err := runWork(h.io, []string{"tree", "--help"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(h.stderr.String(), "riido work tree <ID> --team-id <TEAM-ID>") {
		t.Fatal("user tree help missing")
	}
}
