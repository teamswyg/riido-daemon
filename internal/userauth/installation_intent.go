package userauth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"
)

type installationIntent struct {
	ID             string `json:"installation_intent"`
	Challenge      string `json:"challenge"`
	ExpiresAt      int64  `json:"expires_at"`
	KeyFingerprint string `json:"key_fingerprint"`
}

func (c *Client) createInstallationIntent(ctx context.Context, key InstallationKey, query url.Values) (installationIntent, error) {
	var result installationIntent
	payload := map[string]string{"key_algorithm": key.Algorithm, "public_key": key.PublicKey}
	for _, field := range []string{"response_type", "client_id", "resource", "redirect_uri", "scope", "state", "code_challenge", "code_challenge_method"} {
		payload[field] = query.Get(field)
	}
	body, err := json.Marshal(payload)
	if err != nil || len(body) > 8192 {
		return result, failure("INVALID_CONFIG")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.Resource+"/oauth/cli/installation-intents", bytes.NewReader(body))
	if err != nil {
		return result, failure("INVALID_CONFIG")
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	response, err := c.http.Do(request)
	if err != nil {
		if ctx.Err() != nil {
			return result, ctx.Err()
		}
		return result, &Error{Code: "HTTP_ERROR", Retryable: true}
	}
	defer response.Body.Close()
	body, err = io.ReadAll(io.LimitReader(response.Body, 4097))
	if err != nil || len(body) > 4096 {
		if ctx.Err() != nil {
			return result, ctx.Err()
		}
		return result, failure("INVALID_RESPONSE")
	}
	if response.StatusCode != http.StatusCreated {
		return result, oauthError(response.StatusCode, body)
	}
	if !strictInstallationIntent(body) || json.Unmarshal(body, &result) != nil || result.KeyFingerprint != key.KeyFingerprint ||
		result.ExpiresAt <= 0 || !time.Unix(result.ExpiresAt, 0).After(time.Now()) {
		return installationIntent{}, failure("INVALID_RESPONSE")
	}
	for _, value := range []string{result.ID, result.Challenge} {
		if _, err := decodeInstallationURL(value, 32); err != nil {
			return installationIntent{}, err
		}
	}
	return result, nil
}

func strictInstallationIntent(body []byte) bool {
	decoder := json.NewDecoder(bytes.NewReader(body))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return false
	}
	seen := map[string]bool{}
	for decoder.More() {
		token, err = decoder.Token()
		field, ok := token.(string)
		if err != nil || !ok || seen[field] {
			return false
		}
		switch field {
		case "installation_intent", "challenge", "expires_at", "key_fingerprint":
		default:
			return false
		}
		var value json.RawMessage
		if decoder.Decode(&value) != nil || bytes.Equal(value, []byte("null")) {
			return false
		}
		seen[field] = true
	}
	if token, err = decoder.Token(); err != nil || token != json.Delim('}') || len(seen) != 4 {
		return false
	}
	var extra any
	return errors.Is(decoder.Decode(&extra), io.EOF)
}
