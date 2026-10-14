package saasplane

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPlaneReportsSuccessfulResponseDecodeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{"))
	}))
	t.Cleanup(server.Close)

	agents := []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}}
	plane := newTestPlane(t, server.URL, agents)
	defer plane.Close()

	var out map[string]string
	err := plane.getJSON(context.Background(), "/broken", &out)
	if err == nil {
		t.Fatal("getJSON returned nil error for malformed success response")
	}
	if !strings.Contains(err.Error(), "decode GET /broken 200 OK") {
		t.Fatalf("error = %q, want endpoint decode context", err.Error())
	}
}
