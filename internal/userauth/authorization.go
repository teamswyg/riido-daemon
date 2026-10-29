package userauth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type callbackResult struct {
	code string
	err  error
}

type Authorization struct {
	URL                   string
	RedirectURI           string
	client                *Client
	verifier, state, host string
	result                chan callbackResult
	server                *http.Server
	ctx                   context.Context
	cancel                context.CancelFunc
	closeOnce             sync.Once
	mu                    sync.Mutex
	used, accepted        bool
	installation          *installationAuthorization
}

func (c *Client) BeginAuthorization(ctx context.Context) (*Authorization, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	state, err := randomString()
	if err != nil {
		return nil, err
	}
	verifier, err := randomString()
	if err != nil {
		return nil, err
	}
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return nil, failure("AUTH_CALLBACK")
	}
	authCtx, cancel := context.WithTimeout(ctx, c.config.AuthTimeout)
	a := &Authorization{client: c, state: state, verifier: verifier, host: listener.Addr().String(), result: make(chan callbackResult, 1), ctx: authCtx, cancel: cancel}
	a.RedirectURI = "http://" + a.host + "/callback"
	u, _ := url.Parse(c.config.Endpoints.AuthorizeURL)
	q := u.Query()
	challenge := sha256.Sum256([]byte(verifier))
	for k, v := range map[string]string{"response_type": "code", "client_id": c.config.ClientID, "redirect_uri": a.RedirectURI, "state": state, "code_challenge_method": "S256", "code_challenge": base64.RawURLEncoding.EncodeToString(challenge[:]), "resource": c.config.Resource, "scope": strings.Join(c.config.Scopes, " ")} {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	a.URL = u.String()
	a.server = &http.Server{Handler: http.HandlerFunc(a.callback), ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 5 * time.Second, MaxHeaderBytes: 16 << 10}
	go func() { _ = a.server.Serve(listener) }()
	go func() { <-authCtx.Done(); a.Close() }()
	return a, nil
}

func randomString() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", failure("AUTH_CALLBACK")
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
