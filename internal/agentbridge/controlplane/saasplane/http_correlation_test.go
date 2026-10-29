package saasplane

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPFailurePreservesSafeRequestID(t *testing.T) {
	id := "0123456789abcdef0123456789abcdef"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Riido-Request-ID", id)
		w.WriteHeader(401)
	}))
	defer server.Close()
	plane := newTestPlane(t, server.URL, []AgentBinding{{AgentID: "fixture", RuntimeProvider: "codex"}})
	defer plane.Close()
	err := plane.doJSON(context.Background(), http.MethodPost, "/test", nil, nil)
	var status httpStatusError
	if !errors.As(err, &status) || status.RequestCorrelationID() != id {
		t.Fatalf("missing correlation: %v", err)
	}
	for _, value := range []string{"secret-token", "0123456789abcdef0123456789abcdeg", ""} {
		if got := (httpStatusError{RequestID: value}).RequestCorrelationID(); got != "" {
			t.Fatal("unsafe header retained")
		}
	}
}
