package standalonecli

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestLoginObtainsInstallationIntentBeforeOpeningBrowser(t *testing.T) {
	intents := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/cli/installation-intents" {
			intents++
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()
	service, err := New(Config{APIOrigin: server.URL, Store: &testStore{}})
	if err != nil {
		t.Fatal(err)
	}
	opened := false
	service.openBrowser = func(_ context.Context, address string) error {
		opened = true
		parsed, _ := url.Parse(address)
		if parsed.Query().Get("installation_intent") == "" {
			t.Error("browser received an unbound OAuth request")
		}
		return errors.New("synthetic browser refusal")
	}
	_, err = service.Login(context.Background())
	if err == nil || intents != 1 || opened {
		t.Fatalf("unavailable installation admission reached browser: intents=%d opened=%t", intents, opened)
	}
}
