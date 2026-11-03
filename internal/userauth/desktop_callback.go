package userauth

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"
)

const desktopCallbackLimit = 8 << 10

type desktopCallbackReply struct {
	Intent         string `json:"installation_intent"`
	Challenge      string `json:"challenge"`
	ExpiresAt      int64  `json:"expires_at"`
	KeyFingerprint string `json:"key_fingerprint"`
	Code           string `json:"code"`
	State          string `json:"state"`
	RedirectURL    string `json:"redirect_url"`
}

func (a *Authorization) desktopCallback(w http.ResponseWriter, r *http.Request) {
	if r.Host != a.host || r.URL.Path != "/callback" || r.URL.RawPath != "" || r.URL.IsAbs() || r.URL.RawQuery != "" || r.URL.ForceQuery || r.Method != http.MethodPost || r.Header.Get("Content-Type") != "application/json" || len(r.Header.Values("Origin"))+len(r.Header.Values("Cookie"))+len(r.Header.Values("Authorization")) != 0 {
		http.Error(w, "Invalid authorization response.", http.StatusBadRequest)
		return
	}
	raw, err := io.ReadAll(http.MaxBytesReader(w, r.Body, desktopCallbackLimit))
	if err != nil {
		http.Error(w, "Invalid authorization response.", http.StatusBadRequest)
		return
	}
	result, err := a.decodeDesktopCallback(raw)
	if err != nil {
		http.Error(w, "Invalid authorization response.", http.StatusBadRequest)
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.accepted || a.ctx.Err() != nil {
		http.Error(w, "Authorization response already received.", http.StatusConflict)
		return
	}
	a.accepted = true
	a.result <- result
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("Authorization response received. Return to your terminal."))
}

func (a *Authorization) decodeDesktopCallback(raw []byte) (callbackResult, error) {
	var errorReply struct {
		State string `json:"state"`
		Error string `json:"error"`
	}
	if json.Unmarshal(raw, &errorReply) != nil {
		return callbackResult{}, failure("AUTH_CALLBACK")
	}
	if errorReply.Error != "" {
		if _, err := deviceLinkObject(raw, "state", "error"); err != nil || subtle.ConstantTimeCompare([]byte(errorReply.State), []byte(a.state)) != 1 {
			return callbackResult{}, failure("AUTH_CALLBACK")
		}
		codes := map[string]string{"access_denied": "AUTH_DENIED", "login_required": "DESKTOP_LOGIN_REQUIRED", "temporarily_unavailable": "DESKTOP_UNAVAILABLE"}
		code, ok := codes[errorReply.Error]
		if !ok {
			return callbackResult{}, failure("AUTH_CALLBACK")
		}
		return callbackResult{err: failure(code)}, nil
	}
	if _, err := deviceLinkObject(raw, "installation_intent", "challenge", "expires_at", "key_fingerprint", "code", "state", "redirect_url"); err != nil {
		return callbackResult{}, failure("AUTH_CALLBACK")
	}
	var reply desktopCallbackReply
	if json.Unmarshal(raw, &reply) != nil || !deviceLinkRandom(reply.Intent) || !deviceLinkRandom(reply.Challenge) || !deviceLinkRandom(reply.Code) || reply.KeyFingerprint != a.desktop.key.KeyFingerprint || subtle.ConstantTimeCompare([]byte(reply.State), []byte(a.state)) != 1 {
		return callbackResult{}, failure("AUTH_CALLBACK")
	}
	remaining := time.Until(time.Unix(reply.ExpiresAt, 0))
	if remaining <= 0 || remaining > 15*time.Minute {
		return callbackResult{}, failure("AUTH_CALLBACK")
	}
	expected := a.RedirectURI + "?" + url.Values{"code": {reply.Code}, "state": {a.state}}.Encode()
	if reply.RedirectURL != expected {
		return callbackResult{}, failure("AUTH_CALLBACK")
	}
	request, err := a.DesktopRequest()
	if err != nil {
		return callbackResult{}, err
	}
	challenge, _ := base64.RawURLEncoding.DecodeString(reply.Challenge)
	digest := sha256.Sum256(challenge)
	proof := InstallationProof{Resource: request.Resource, ClientID: request.ClientID, Intent: reply.Intent, ChallengeHash: hex.EncodeToString(digest[:]), KeyFingerprint: reply.KeyFingerprint, Code: reply.Code, RedirectURI: a.RedirectURI, CodeChallenge: request.CodeChallenge, State: a.state, ExpiresAt: reply.ExpiresAt}
	return callbackResult{code: reply.Code, installation: &installationAuthorization{proof: proof, sign: a.desktop.sign}}, nil
}
