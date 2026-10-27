package userauth

import (
	"context"
	"net/url"
	"strings"
	"time"
)

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	Scope        string `json:"scope"`
}

func (c *Client) exchange(ctx context.Context, code, verifier, redirect string) (Credential, error) {
	return c.token(ctx, url.Values{"grant_type": {"authorization_code"}, "code": {code}, "code_verifier": {verifier}, "redirect_uri": {redirect}})
}

func (c *Client) Refresh(ctx context.Context, refreshToken string) (Credential, error) {
	if refreshToken == "" {
		return Credential{}, ErrReauthRequired
	}
	result, err := c.token(ctx, url.Values{"grant_type": {"refresh_token"}, "refresh_token": {refreshToken}})
	if err == nil && result.RefreshToken == "" {
		result.RefreshToken = refreshToken
	}
	return result, err
}

func (c *Client) token(ctx context.Context, form url.Values) (Credential, error) {
	var result tokenResponse
	if err := c.formRequest(ctx, c.config.Endpoints.TokenURL, form, &result); err != nil {
		return Credential{}, err
	}
	if result.AccessToken == "" || !strings.EqualFold(result.TokenType, "Bearer") || result.ExpiresIn <= 0 || result.ExpiresIn > 31*24*60*60 {
		return Credential{}, failure("INVALID_RESPONSE")
	}
	return Credential{AccessToken: result.AccessToken, RefreshToken: result.RefreshToken, TokenType: "Bearer", ExpiresAt: time.Now().Add(time.Duration(result.ExpiresIn) * time.Second), Scope: result.Scope}, nil
}

func (c *Client) Revoke(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return c.formRequest(ctx, c.config.Endpoints.RevokeURL, url.Values{"token": {token}}, nil)
}
