package userauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

const (
	fixtureIntent    = "ERERERERERERERERERERERERERERERERERERERERERE"
	fixtureChallenge = "IiIiIiIiIiIiIiIiIiIiIiIiIiIiIiIiIiIiIiIiIiI"
)

func fixtureInstallationKey() InstallationKey {
	return InstallationKey{
		Algorithm:      "p256-sha256-x962-der-v1",
		PublicKey:      "BGsX0fLhLEJH-Lzm5WOkQPJ3A32BLeszoPShOUXYmMKWT-NC4v4af5uO5-tKfA-eFivOM1drMV7Oy7ZAaDe_UfU",
		KeyFingerprint: "698bea63dc44a344663ff1429aea10842df27b6b991ef25866b2c6c02cdcc5be",
	}
}

func writeFixtureIntent(w http.ResponseWriter, key InstallationKey, expiry int64) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{"installation_intent": fixtureIntent, "challenge": fixtureChallenge, "expires_at": expiry, "key_fingerprint": key.KeyFingerprint})
}

func TestInstallationAuthorizationBindsIntentAndTypedProofToExchange(t *testing.T) {
	key := fixtureInstallationKey()
	expiry := time.Now().Add(10 * time.Minute).Unix()
	var request map[string]string
	var exchanged url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth/cli/installation-intents":
			if r.Method != http.MethodPost || r.Header.Get("Authorization") != "" || r.URL.RawQuery != "" || r.Header.Get("Content-Type") != "application/json" {
				t.Error("intent did not use unauthenticated bounded JSON POST")
			}
			if json.NewDecoder(r.Body).Decode(&request) != nil {
				t.Error("invalid intent body")
			}
			writeFixtureIntent(w, key, expiry)
		case "/oauth/token":
			if r.ParseForm() != nil {
				t.Error("invalid token form")
			}
			exchanged = r.PostForm
			fmt.Fprint(w, `{"access_token":"fixture-access","refresh_token":"fixture-refresh","token_type":"Bearer","expires_in":600,"scope":"read"}`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	client, _ := NewClient(configFor(server))
	var proof InstallationProof
	signed := 0
	flow, err := client.BeginInstallationAuthorization(context.Background(), key, func(_ context.Context, p InstallationProof) (string, error) {
		proof = p
		signed++
		return "MAYCAQECAQE", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	defer flow.Close()
	u, _ := url.Parse(flow.URL)
	q := u.Query()
	if request == nil || len(request) != 10 || request["public_key"] != key.PublicKey || request["key_algorithm"] != key.Algorithm || request["client_id"] != "riido-cli" || request["scope"] != "read" || request["response_type"] != "code" || request["code_challenge_method"] != "S256" || request["resource"] != server.URL || request["redirect_uri"] != flow.RedirectURI || request["state"] != q.Get("state") || request["code_challenge"] != q.Get("code_challenge") {
		t.Fatal("intent does not bind the generated OAuth request and protected key")
	}
	if len(q) != 9 || q.Get("installation_intent") != fixtureIntent || strings.Contains(flow.URL, key.PublicKey) || strings.Contains(flow.URL, fixtureChallenge) {
		t.Fatal("browser URL did not contain only original OAuth fields and intent")
	}
	response, err := http.Get(flow.RedirectURI + "?state=" + url.QueryEscape(q.Get("state")) + "&code=fixture-code")
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	credential, err := flow.Wait(context.Background())
	if err != nil || credential.AccessToken != "fixture-access" || signed != 1 {
		t.Fatalf("bound flow failed: %v signed=%d", err, signed)
	}
	if proof.Intent != fixtureIntent || proof.ChallengeHash != "9f72ea0cf49536e3c66c787f705186df9a4378083753ae9536d65b3ad7fcddc4" || proof.KeyFingerprint != key.KeyFingerprint || proof.Code != "fixture-code" || proof.State != q.Get("state") || proof.CodeChallenge != q.Get("code_challenge") || proof.RedirectURI != flow.RedirectURI || proof.Resource != server.URL || proof.ClientID != "riido-cli" || proof.ExpiresAt != expiry {
		t.Fatal("signer proof differs from original request")
	}
	if len(exchanged) != 6 || exchanged.Get("installation_signature") != "MAYCAQECAQE" || exchanged.Get("code") != "fixture-code" || exchanged.Get("redirect_uri") != flow.RedirectURI || exchanged.Get("grant_type") != "authorization_code" || exchanged.Get("client_id") != "riido-cli" || exchanged.Get("code_verifier") == "" {
		t.Fatal("token form lost proof or exposed installation selector")
	}
}

func TestInstallationAuthorizationRefusesInvalidAdmissionBeforeBrowser(t *testing.T) {
	key := fixtureInstallationKey()
	expiry := time.Now().Add(time.Minute).Unix()
	valid := fmt.Sprintf(`{"installation_intent":%q,"challenge":%q,"expires_at":%d,"key_fingerprint":%q}`, fixtureIntent, fixtureChallenge, expiry, key.KeyFingerprint)
	for _, body := range []string{
		strings.Replace(valid, fixtureIntent, "bad", 1), strings.Replace(valid, fixtureChallenge, "bad", 1), strings.Replace(valid, key.KeyFingerprint, strings.Repeat("0", 64), 1),
		strings.Replace(valid, fmt.Sprint(expiry), "1", 1), strings.Replace(valid, fmt.Sprint(expiry), "null", 1), strings.Replace(valid, `"expires_at":`, `"expires_at":1,"expires_at":`, 1),
		strings.Replace(valid, `"challenge":`, `"unknown":"private","challenge":`, 1), strings.TrimSuffix(valid, "}") + ",}", valid + ` {}`, strings.Repeat("x", 8193),
	} {
		t.Run(fmt.Sprint(len(body)), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusCreated); fmt.Fprint(w, body) }))
			defer server.Close()
			client, _ := NewClient(configFor(server))
			flow, err := client.BeginInstallationAuthorization(context.Background(), key, func(context.Context, InstallationProof) (string, error) {
				t.Error("invalid admission reached signer")
				return "", nil
			})
			if flow != nil {
				flow.Close()
			}
			if err == nil || flow != nil {
				t.Fatal("invalid admission returned browser flow")
			}
		})
	}
}

func TestInstallationSignerRefusalNeverExchangesCode(t *testing.T) {
	key := fixtureInstallationKey()
	tokens := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/cli/installation-intents" {
			writeFixtureIntent(w, key, time.Now().Add(time.Minute).Unix())
			return
		}
		tokens++
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()
	client, _ := NewClient(configFor(server))
	denied := errors.New("synthetic protection refusal")
	flow, err := client.BeginInstallationAuthorization(context.Background(), key, func(context.Context, InstallationProof) (string, error) { return "", denied })
	if err != nil {
		t.Fatal(err)
	}
	defer flow.Close()
	u, _ := url.Parse(flow.URL)
	r, err := http.Get(flow.RedirectURI + "?state=" + u.Query().Get("state") + "&code=fixture-code")
	if err != nil {
		t.Fatal(err)
	}
	r.Body.Close()
	_, err = flow.Wait(context.Background())
	if !errors.Is(err, denied) || tokens != 0 {
		t.Fatalf("signer refusal bypassed: error=%v tokens=%d", err, tokens)
	}
}
