package userauth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/url"
)

// DesktopAuthorizationRequest contains public request data only. The verifier
// and the app's login token never cross this transport.
type DesktopAuthorizationRequest struct {
	ResponseType        string `json:"response_type"`
	ClientID            string `json:"client_id"`
	Resource            string `json:"resource"`
	RedirectURI         string `json:"redirect_uri"`
	Scope               string `json:"scope"`
	State               string `json:"state"`
	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`
	KeyAlgorithm        string `json:"key_algorithm"`
	PublicKey           string `json:"public_key"`
}

func (DesktopAuthorizationRequest) String() string { return "[redacted Desktop authorization request]" }
func (DesktopAuthorizationRequest) GoString() string {
	return "[redacted Desktop authorization request]"
}

type desktopAuthorization struct {
	key  InstallationKey
	sign InstallationSigner
}

func (c *Client) BeginDesktopAuthorization(ctx context.Context, key InstallationKey, sign InstallationSigner) (*Authorization, error) {
	if err := c.validateInstallationConfiguration(key, sign); err != nil {
		return nil, err
	}
	return c.beginAuthorization(ctx, &desktopAuthorization{key: key, sign: sign})
}

func (a *Authorization) DesktopRequest() (DesktopAuthorizationRequest, error) {
	if a.desktop == nil {
		return DesktopAuthorizationRequest{}, failure("INVALID_CONFIG")
	}
	u, _ := url.Parse(a.URL)
	q := u.Query()
	return DesktopAuthorizationRequest{ResponseType: "code", ClientID: a.client.config.ClientID, Resource: a.client.config.Resource, RedirectURI: a.RedirectURI, Scope: "read", State: a.state, CodeChallenge: q.Get("code_challenge"), CodeChallengeMethod: "S256", KeyAlgorithm: a.desktop.key.Algorithm, PublicKey: a.desktop.key.PublicKey}, nil
}

// LaunchURL selects only known Riido app environments. Arbitrary API endpoints
// remain supported by independent browser login, not app-token delegation.
func (r DesktopAuthorizationRequest) LaunchURL() (string, error) {
	var scheme string
	switch r.Resource {
	case "https://monolith.riido.io":
		scheme = "riido-cli-auth"
	case "https://monolith.api.riido.io":
		scheme = "riido-staging-cli-auth"
	default:
		return "", failure("DESKTOP_UNAVAILABLE")
	}
	raw, err := json.Marshal(r)
	if err != nil {
		return "", failure("INVALID_CONFIG")
	}
	return scheme + "://cli/authorize/v1?request=" + base64.RawURLEncoding.EncodeToString(raw), nil
}
