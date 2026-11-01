package userauth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestCredentialBindingSurvivesRefresh(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, recoveryResponse) }))
	defer server.Close()
	client, err := NewClient(configFor(server))
	if err != nil {
		t.Fatal(err)
	}
	credential := validCredential()
	credential.ExpiresAt = time.Now()
	credential.Binding = &CredentialBinding{Resource: server.URL, ClientID: "riido-cli", AccountID: "verified-account", KeyFingerprint: strings.Repeat("a", 64), AuthOrigin: AuthOriginDesktop}
	store := newFakeStore(credential)
	if _, err := NewManager(client, store).AccessToken(t.Context()); err != nil {
		t.Fatal(err)
	}
	if store.credential.Binding == nil || *store.credential.Binding != *credential.Binding {
		t.Fatal("refresh discarded verified custody binding")
	}
}

func TestCredentialBindingRejectsWrongClientBeforeUseOrMutation(t *testing.T) {
	for _, operation := range []string{"access", "logout", "replacement", "new-login"} {
		t.Run(operation, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { calls++; w.WriteHeader(http.StatusOK) }))
			defer server.Close()
			client, err := NewClient(configFor(server))
			if err != nil {
				t.Fatal(err)
			}
			wrong := validCredential()
			wrong.Binding = &CredentialBinding{Resource: "https://other.example.test", ClientID: "riido-cli", AccountID: "verified-account", KeyFingerprint: strings.Repeat("a", 64), AuthOrigin: AuthOriginBrowser}
			store := newFakeStore(wrong)
			manager := NewManager(client, store)
			switch operation {
			case "access":
				_, err = manager.AccessToken(t.Context())
			case "logout":
				err = manager.Logout(t.Context())
			case "replacement":
				next := validCredential()
				next.AccessToken = "new-access"
				err = manager.Login(t.Context(), next)
			case "new-login":
				err = manager.Login(t.Context(), wrong)
			}
			if err == nil || calls != 0 || !store.present || !reflect.DeepEqual(store.credential, wrong) {
				t.Fatal("binding crossed client or mutated stored login", err, calls)
			}
		})
	}
}
