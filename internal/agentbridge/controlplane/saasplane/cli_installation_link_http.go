package saasplane

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"net/url"
	"time"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

// A dedicated transport prevents caller-controlled endpoints, redirected device
// credentials, and reflected server bodies from crossing the local IPC boundary.
func (c *CLIInstallationLinkClient) postConfirmation(ctx context.Context, proof contextcli.CLIInstallationLinkProof, signature string) error {
	body, err := json.Marshal(struct {
		SchemaVersion string `json:"schema_version"`
		Signature     string `json:"signature"`
	}{"riido-cli-installation-link-confirm.v1", signature})
	if err != nil {
		return cliLinkFailure("LINK_INVALID_REQUEST")
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL+"/v1/daemon/cli-installation-links/"+url.PathEscape(proof.IntentID)+"/confirm", bytes.NewReader(body))
	if err != nil {
		return cliLinkFailure("LINK_INVALID_REQUEST")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Riido-Device-Id", c.cfg.DeviceID)
	req.Header.Set("X-Riido-Device-Secret", c.cfg.DeviceSecret)
	client := *http.DefaultClient
	if c.cfg.HTTPClient != nil {
		client = *c.cfg.HTTPClient
	}
	client.Jar = nil
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := client.Do(req)
	if err != nil {
		return cliLinkFailure("LINK_UNAVAILABLE")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return cliLinkHTTPFailure(resp.StatusCode)
	}
	contentType, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil || contentType != "application/json" {
		return cliLinkFailure("LINK_INVALID_RESPONSE")
	}
	payload, err := io.ReadAll(io.LimitReader(resp.Body, (16<<10)+1))
	if err != nil || len(payload) > 16<<10 || !validCLILinkACK(payload, proof) {
		return cliLinkFailure("LINK_INVALID_RESPONSE")
	}
	return nil
}

func cliLinkHTTPFailure(status int) error {
	switch status {
	case http.StatusUnauthorized, http.StatusForbidden:
		return cliLinkFailure("LINK_PERMISSION_DENIED")
	case http.StatusNotFound, http.StatusMethodNotAllowed:
		return cliLinkFailure("LINK_UNSUPPORTED")
	case http.StatusBadRequest:
		return cliLinkFailure("LINK_INVALID_REQUEST")
	case http.StatusConflict:
		return cliLinkFailure("LINK_CONFLICT")
	case http.StatusGone:
		return cliLinkFailure("LINK_EXPIRED")
	case http.StatusRequestTimeout, http.StatusTooManyRequests:
		return cliLinkFailure("LINK_UNAVAILABLE")
	default:
		if status >= 500 {
			return cliLinkFailure("LINK_UNAVAILABLE")
		}
		return cliLinkFailure("LINK_INVALID_RESPONSE")
	}
}
