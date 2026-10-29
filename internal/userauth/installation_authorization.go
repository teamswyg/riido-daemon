package userauth

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net/url"
	"time"
)

type installationAuthorization struct {
	proof InstallationProof
	sign  InstallationSigner
}

func (c *Client) BeginInstallationAuthorization(ctx context.Context, key InstallationKey, sign InstallationSigner) (*Authorization, error) {
	if err := c.validateInstallationConfiguration(key, sign); err != nil {
		return nil, err
	}
	a, err := c.BeginAuthorization(ctx)
	if err != nil {
		return nil, err
	}
	address, _ := url.Parse(a.URL)
	query := address.Query()
	intent, err := c.createInstallationIntent(a.ctx, key, query)
	if err != nil {
		a.Close()
		return nil, err
	}
	if err := a.ctx.Err(); err != nil {
		a.Close()
		return nil, err
	}
	challenge, _ := base64.RawURLEncoding.DecodeString(intent.Challenge)
	digest := sha256.Sum256(challenge)
	a.installation = &installationAuthorization{sign: sign, proof: InstallationProof{
		Resource: c.config.Resource, ClientID: c.config.ClientID, Intent: intent.ID,
		ChallengeHash: hex.EncodeToString(digest[:]), KeyFingerprint: key.KeyFingerprint,
		RedirectURI: a.RedirectURI, CodeChallenge: query.Get("code_challenge"), State: a.state, ExpiresAt: intent.ExpiresAt,
	}}
	query.Set("installation_intent", intent.ID)
	address.RawQuery = query.Encode()
	a.URL = address.String()
	return a, nil
}

func (c *Client) validateInstallationConfiguration(key InstallationKey, sign InstallationSigner) error {
	if sign == nil || key.Algorithm != "p256-sha256-x962-der-v1" {
		return failure("STORAGE_ERROR")
	}
	raw, err := decodeInstallationURL(key.PublicKey, 65)
	if err != nil {
		return failure("STORAGE_ERROR")
	}
	if _, err = ecdsa.ParseUncompressedPublicKey(elliptic.P256(), raw); err != nil {
		return failure("STORAGE_ERROR")
	}
	digest := sha256.Sum256(raw)
	if hex.EncodeToString(digest[:]) != key.KeyFingerprint {
		return failure("STORAGE_ERROR")
	}
	if c.config.ClientID != "riido-cli" || len(c.config.Scopes) != 1 || c.config.Scopes[0] != "read" ||
		c.resource.Path != "" || c.resource.RawPath != "" || c.resource.RawQuery != "" || c.resource.ForceQuery {
		return failure("INVALID_CONFIG")
	}
	for _, endpoint := range []string{c.config.Endpoints.AuthorizeURL, c.config.Endpoints.TokenURL, c.config.Endpoints.RevokeURL} {
		u, _ := url.Parse(endpoint)
		if !c.permits(u) || u.RawQuery != "" || u.ForceQuery {
			return failure("INVALID_CONFIG")
		}
	}
	return nil
}

func decodeInstallationURL(value string, size int) ([]byte, error) {
	raw, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil || len(raw) != size || base64.RawURLEncoding.EncodeToString(raw) != value {
		return nil, failure("INVALID_RESPONSE")
	}
	return raw, nil
}

func (a *Authorization) waitContext() (context.Context, context.CancelFunc) {
	if a.installation != nil {
		return context.WithDeadline(a.ctx, time.Unix(a.installation.proof.ExpiresAt, 0))
	}
	return context.WithCancel(a.ctx)
}
