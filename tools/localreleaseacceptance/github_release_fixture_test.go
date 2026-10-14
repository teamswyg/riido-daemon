package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func releaseAPIServer(t *testing.T, body string) string {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, body)
	}))
	t.Cleanup(server.Close)
	return server.URL
}

func releaseBody(assets ...string) string {
	names := make([]string, 0, len(assets))
	for _, asset := range assets {
		names = append(names, fmt.Sprintf(`{"name":%q}`, asset))
	}
	return fmt.Sprintf(`[{"tag_name":"v-test","draft":false,"assets":[%s]}]`, strings.Join(names, ","))
}
