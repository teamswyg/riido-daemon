package userauth

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/url"
	"slices"
	"strings"
	"time"
	"unicode/utf8"
)

func (c *Client) desktopToken(ctx context.Context, form url.Values) (Credential, error) {
	var raw json.RawMessage
	if err := c.formRequest(ctx, c.config.Endpoints.TokenURL, form, &raw); err != nil {
		return Credential{}, err
	}
	var result tokenResponse
	if json.Unmarshal(raw, &result) != nil {
		return Credential{}, failure("INVALID_RESPONSE")
	}
	credential, err := credentialFromToken(result)
	if err != nil {
		return Credential{}, err
	}
	if !validDesktopTokenShape(raw) || result.AuthOrigin != AuthOriginDesktop || result.Scope != "read" || result.RefreshToken == "" {
		// This direct exchange issued a new grant; it has never replaced a saved
		// session. Attempt its cleanup without reading/revoking either old slot.
		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		_ = c.Revoke(cleanup, revocationToken(credential))
		return Credential{}, failure("AUTH_ORIGIN_MISMATCH")
	}
	return credential, nil
}

// Tokens come directly from the configured server, so additive metadata may
// evolve independently. Required fields still have one exact, non-null spelling.
func validDesktopTokenShape(raw []byte) bool {
	if !utf8.Valid(raw) {
		return false
	}
	required := []string{"access_token", "refresh_token", "token_type", "expires_in", "scope", "auth_origin"}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	if token, err := decoder.Token(); err != nil || token != json.Delim('{') {
		return false
	}
	seen := make(map[string]bool)
	for decoder.More() {
		token, err := decoder.Token()
		key, ok := token.(string)
		if err != nil || !ok || seen[key] {
			return false
		}
		for _, field := range required {
			if strings.EqualFold(key, field) && key != field {
				return false
			}
		}
		var value json.RawMessage
		if decoder.Decode(&value) != nil || (slices.Contains(required, key) && bytes.Equal(bytes.TrimSpace(value), []byte("null"))) {
			return false
		}
		seen[key] = true
	}
	if token, err := decoder.Token(); err != nil || token != json.Delim('}') || decoder.Decode(new(any)) != io.EOF {
		return false
	}
	for _, field := range required {
		if !seen[field] {
			return false
		}
	}
	return true
}
