package userauth

import (
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	config   Config
	http     *http.Client
	resource *url.URL
}

func NewClient(config Config) (*Client, error) {
	if strings.TrimSpace(config.ClientID) == "" {
		return nil, failure("INVALID_CONFIG")
	}
	for _, raw := range []string{config.Endpoints.AuthorizeURL, config.Endpoints.TokenURL, config.Endpoints.RevokeURL, config.Resource} {
		if _, err := safeURL(raw); err != nil {
			return nil, err
		}
	}
	resource, _ := safeURL(config.Resource)
	if resource.RawQuery != "" || (resource.Path != "" && resource.Path != "/") {
		return nil, failure("INVALID_CONFIG")
	}
	if config.AuthTimeout <= 0 {
		config.AuthTimeout = 5 * time.Minute
	}
	config.Scopes = append([]string(nil), config.Scopes...)
	hc := http.Client{Timeout: 30 * time.Second}
	if config.HTTPClient != nil {
		hc = *config.HTTPClient
		if hc.Timeout == 0 {
			hc.Timeout = 30 * time.Second
		}
	}
	hc.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	return &Client{config: config, http: &hc, resource: resource}, nil
}

func safeURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || u.User != nil || u.Fragment != "" || u.Opaque != "" {
		return nil, failure("INVALID_CONFIG")
	}
	if u.Scheme != "https" {
		ip := net.ParseIP(u.Hostname())
		if u.Scheme != "http" || ip == nil || !ip.IsLoopback() {
			return nil, failure("INVALID_CONFIG")
		}
	}
	return u, nil
}

func (c *Client) permits(u *url.URL) bool {
	if u == nil {
		return false
	}
	if _, err := safeURL(u.String()); err != nil {
		return false
	}
	return u.Scheme == c.resource.Scheme && strings.EqualFold(u.Host, c.resource.Host)
}
