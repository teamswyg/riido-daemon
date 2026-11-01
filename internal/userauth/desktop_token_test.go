package userauth

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestDesktopTokenAllowsAdditiveMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth/token" {
			t.Error("valid grant was revoked")
			return
		}
		fmt.Fprint(w, strings.TrimSuffix(desktopTokenFixture, "}")+`,"metadata":{"v":2},"optional":null}`)
	}))
	defer server.Close()
	client, _ := NewClient(configFor(server))
	if credential, err := client.desktopToken(context.Background(), url.Values{}); err != nil || credential.AccessToken != "desktop-access" {
		t.Fatal("additive metadata broke login", err)
	}
}

func TestDesktopTokenShapeRejectsAmbiguity(t *testing.T) {
	for name, body := range map[string]string{
		"duplicate":           strings.TrimSuffix(desktopTokenFixture, "}") + `,"auth_origin":"desktop"}`,
		"alias":               strings.TrimSuffix(desktopTokenFixture, "}") + `,"AUTH_ORIGIN":"desktop"}`,
		"null":                strings.Replace(desktopTokenFixture, `"scope":"read"`, `"scope":null`, 1),
		"missing":             strings.Replace(desktopTokenFixture, `,"scope":"read"`, "", 1),
		"trailing":            desktopTokenFixture + `{}`,
		"duplicate-extension": strings.TrimSuffix(desktopTokenFixture, "}") + `,"meta":1,"meta":2}`,
	} {
		t.Run(name, func(t *testing.T) {
			if validDesktopTokenShape([]byte(body)) {
				t.Fatal("ambiguous token accepted")
			}
		})
	}
}
