package userauth

import (
	"crypto/subtle"
	"net/http"
)

func (a *Authorization) callback(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Security-Policy", "default-src 'none'")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if r.Host != a.host || r.URL.Path != "/callback" || r.Method != http.MethodGet || len(r.URL.RawQuery) > 16384 {
		http.Error(w, "Invalid authorization response.", http.StatusBadRequest)
		return
	}
	q, err := parseCallback(r)
	if err != nil || subtle.ConstantTimeCompare([]byte(q.state), []byte(a.state)) != 1 {
		http.Error(w, "Invalid authorization response.", http.StatusBadRequest)
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.accepted {
		http.Error(w, "Authorization response already received.", http.StatusConflict)
		return
	}
	if q.code == "" && q.failure == "" {
		http.Error(w, "Invalid authorization response.", http.StatusBadRequest)
		return
	}
	a.accepted = true
	result := callbackResult{code: q.code}
	if q.failure != "" {
		result.err = failure("AUTH_DENIED")
	}
	a.result <- result
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("Authorization response received. Return to your terminal."))
}

type callbackValues struct{ state, code, failure string }

func parseCallback(r *http.Request) (callbackValues, error) {
	q := r.URL.Query()
	for _, key := range []string{"state", "code", "error"} {
		if len(q[key]) > 1 {
			return callbackValues{}, failure("AUTH_CALLBACK")
		}
	}
	if q.Get("code") != "" && q.Get("error") != "" {
		return callbackValues{}, failure("AUTH_CALLBACK")
	}
	return callbackValues{state: q.Get("state"), code: q.Get("code"), failure: q.Get("error")}, nil
}
