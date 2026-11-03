package userauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/url"
)

const refreshRecoveryGrant = "urn:riido:params:oauth:grant-type:refresh-recovery"

func newRefreshRecoverySecret() (string, error) {
	var random [32]byte
	if _, err := rand.Read(random[:]); err != nil {
		return "", failure("STORAGE_ERROR")
	}
	return base64.RawURLEncoding.EncodeToString(random[:]), nil
}

func (c *Client) RefreshRecover(ctx context.Context, refresh, secret string) (Credential, error) {
	raw, err := base64.RawURLEncoding.DecodeString(secret)
	if refresh == "" || err != nil || len(raw) != 32 || base64.RawURLEncoding.EncodeToString(raw) != secret {
		return Credential{}, failure("INVALID_CONFIG")
	}
	var result tokenResponse
	form := url.Values{"grant_type": {refreshRecoveryGrant}, "refresh_token": {refresh}, "recovery_secret": {secret}}
	if err := c.formRequest(ctx, c.config.Endpoints.TokenURL, form, &result); err != nil {
		return Credential{}, err
	}
	if result.RefreshRecovery != "v1" || result.RefreshToken == "" || result.RefreshToken == refresh {
		return Credential{}, failure("INVALID_RESPONSE")
	}
	return credentialFromToken(result)
}
