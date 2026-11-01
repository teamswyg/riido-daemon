package userauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDesktopCallbackRejectsMalformedWithoutConsumingFlow(t *testing.T) {
	for _, name := range []string{"host", "path", "query", "GET", "origin", "empty-origin", "cookie", "authorization", "content-type", "wrong-state", "wrong-key", "wrong-intent", "wrong-code", "redirect", "expired", "far-future", "duplicate", "case-alias", "unknown", "null", "trailing", "oversized", "mixed-error", "unknown-error"} {
		t.Run(name, func(t *testing.T) {
			calls, signs := 0, 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls++; fmt.Fprint(w, desktopTokenFixture) }))
			defer server.Close()
			c, _ := NewClient(configFor(server))
			a, err := c.BeginDesktopAuthorization(t.Context(), fixtureInstallationKey(), func(context.Context, InstallationProof) (string, error) { signs++; return "MAYCAQECAQE", nil })
			if err != nil {
				t.Fatal(err)
			}
			defer a.Close()
			valid := string(desktopReply(t, a))
			bad := valid
			change := func(r *http.Request) {
				switch name {
				case "host":
					r.Host = "evil.test"
				case "path":
					r.URL.Path = "/other"
				case "query":
					r.URL.RawQuery = "extra=1"
				case "GET":
					r.Method = http.MethodGet
				case "origin":
					r.Header.Set("Origin", "https://evil.test")
				case "empty-origin":
					r.Header.Set("Origin", "")
				case "cookie":
					r.Header.Set("Cookie", "app-token=not-accepted")
				case "authorization":
					r.Header.Set("Authorization", "Bearer not-accepted")
				case "content-type":
					r.Header.Set("Content-Type", "text/plain")
				}
			}
			switch name {
			case "wrong-state":
				bad = strings.Replace(valid, `"state":"`, `"state":"wrong`, 1)
			case "wrong-key":
				bad = strings.Replace(valid, fixtureInstallationKey().KeyFingerprint, strings.Repeat("0", 64), 1)
			case "wrong-intent":
				bad = strings.Replace(valid, fixtureIntent, "bad", 1)
			case "wrong-code":
				bad = strings.Replace(valid, `"code":"`, `"code":"bad`, 1)
			case "redirect":
				bad = strings.Replace(valid, `"redirect_url":"`, `"redirect_url":"https://evil.test/`, 1)
			case "expired", "far-future":
				var reply desktopCallbackReply
				if err := json.Unmarshal([]byte(valid), &reply); err != nil {
					t.Fatal(err)
				}
				reply.ExpiresAt = 1
				if name == "far-future" {
					reply.ExpiresAt = 1 << 60
				}
				raw, err := json.Marshal(reply)
				if err != nil {
					t.Fatal(err)
				}
				bad = string(raw)
			case "duplicate":
				bad = strings.TrimSuffix(valid, "}") + `,"state":"` + a.state + `"}`
			case "case-alias":
				bad = strings.Replace(valid, `"state":`, `"State":`, 1)
			case "unknown":
				bad = strings.TrimSuffix(valid, "}") + `,"parent_session_id":"untrusted"}`
			case "null":
				bad = strings.Replace(valid, `"challenge":"`+fixtureChallenge+`"`, `"challenge":null`, 1)
			case "trailing":
				bad = valid + ` {}`
			case "oversized":
				bad = strings.Repeat(" ", desktopCallbackLimit) + valid
			case "mixed-error":
				bad = strings.TrimSuffix(valid, "}") + `,"error":"access_denied"}`
			case "unknown-error":
				bad = `{"state":"` + a.state + `","error":"unknown-private-detail"}`
			}
			if got := postDesktopReply(t, a, []byte(bad), change); got != http.StatusBadRequest {
				t.Fatal("invalid callback accepted", got)
			}
			if calls != 0 || signs != 0 {
				t.Fatal("invalid callback contacted signer or server")
			}
			if got := postDesktopReply(t, a, []byte(valid), nil); got != http.StatusOK {
				t.Fatal("invalid callback consumed valid response", got)
			}
			if _, err := a.Wait(t.Context()); err != nil || signs != 1 || calls != 1 {
				t.Fatal("valid recovery failed", err, signs, calls)
			}
		})
	}
}
