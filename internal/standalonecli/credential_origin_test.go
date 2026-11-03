package standalonecli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/userauth"
)

func custodyBinding(origin string) *userauth.CredentialBinding {
	return &userauth.CredentialBinding{Resource: protectedTestResource, ClientID: "riido-cli", AccountID: "verified-account", KeyFingerprint: strings.Repeat("a", 64), AuthOrigin: origin}
}

func TestCredentialOriginStrictBindingAndContext(t *testing.T) {
	s, n, credential, _ := newProtectedFixture(t)
	credential.Binding = custodyBinding("browser")
	raw, err := encodeCredential(credential)
	if err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(string) string{
		func(x string) string {
			return strings.Replace(x, `"auth_origin":"browser"`, `"auth_origin":"desktop","auth_origin":"browser"`, 1)
		},
		func(x string) string {
			return strings.Replace(x, `"auth_origin":"browser"`, `"Auth_Origin":"browser"`, 1)
		},
		func(x string) string {
			return strings.Replace(x, `"auth_origin":"browser"`, `"auth_origin":"unknown"`, 1)
		},
		func(x string) string {
			return strings.Replace(x, `"account_id":"verified-account"`, `"account_id":null`, 1)
		},
		func(x string) string { return strings.Replace(x, `"client_id":"riido-cli",`, ``, 1) },
		func(x string) string {
			return strings.Replace(x, `"auth_origin":"browser"`, `"auth_origin":"browser","unowned":"value"`, 1)
		},
		func(x string) string {
			return strings.Replace(x, `"account_id":"verified-account"`, `"account_id":""`, 1)
		},
		func(x string) string {
			return strings.Replace(x, `"resource":"https://api.example.test"`, `"resource":"https://api.example.test/path"`, 1)
		},
	} {
		mutated := []byte(change(string(raw)))
		if bytes.Equal(mutated, raw) {
			t.Fatal("invalid case did not change input")
		}
		n.raw = mutated
		if _, err := s.Load(t.Context()); err == nil {
			t.Fatal("invalid binding accepted")
		}
		if err := saveProtectedCredential(t, s, credential); err == nil || !bytes.Equal(n.raw, mutated) {
			t.Fatal("unknown binding overwritten")
		}
	}
	n.raw = raw
	for _, other := range []*ProtectedStore{{Native: n, Resource: s.Resource, AuthOrigin: "desktop"}, {Native: n, Resource: "https://another.test"}, {Native: n, Resource: s.Resource, AuthOrigin: "invalid"}} {
		if _, err := other.Load(t.Context()); err == nil {
			t.Fatal("wrong origin/resource read credential")
		}
	}
	legacy := credential
	legacy.Binding = nil
	unbound, err := json.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	n.raw = unbound
	if err = saveProtectedCredential(t, s, credential); err == nil || !bytes.Equal(n.raw, unbound) {
		t.Fatal("same legacy session acquired invented provenance")
	}
}

func TestCredentialOriginBoundRoundTrip(t *testing.T) {
	_, _, credential, _ := newProtectedFixture(t)
	credential.Binding = custodyBinding(userauth.AuthOriginBrowser)
	raw, err := encodeCredential(credential)
	if err != nil {
		t.Fatal(err)
	}
	got, err := decodeCredential(raw)
	if err != nil || got.Binding == nil || *got.Binding != *credential.Binding {
		t.Fatal("verified binding did not survive protected round trip", err)
	}
	if !bytes.Contains(raw, []byte(`"schema_version":"riido-cli-protected-session.v4"`)) {
		t.Fatal("new bound record is readable by an old unbound client")
	}
}

func TestCredentialOriginDesktopRejectsLegacyWithoutRewrite(t *testing.T) {
	s, n, _, _ := newProtectedFixture(t)
	s.AuthOrigin = userauth.AuthOriginDesktop
	before := bytes.Clone(n.raw)
	if _, err := s.Load(t.Context()); err == nil {
		t.Fatal("Desktop read legacy browser credential")
	}
	if !bytes.Equal(before, n.raw) || n.saves != 0 || n.deletes != 0 {
		t.Fatal("rejected record changed")
	}
}
