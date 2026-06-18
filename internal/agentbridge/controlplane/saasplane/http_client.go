package saasplane

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (p *Plane) doJSON(ctx context.Context, method, path string, body []byte, out any) error {
	ctx, cancel := context.WithTimeout(ctx, p.cfg.RequestTimeout)
	defer cancel()

	attempts := 1
	if retryableJSONRequest(method, path) {
		attempts = jsonRequestMaxAttempts
	}

	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		var reader io.Reader
		if body != nil {
			reader = bytes.NewReader(body)
		}
		req, err := http.NewRequestWithContext(ctx, method, p.cfg.BaseURL+path, reader)
		if err != nil {
			return err
		}
		if method == http.MethodPost {
			req.Header.Set("Content-Type", "application/json")
		}
		p.attachAuthHeaders(req)
		resp, err := p.client.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			lastErr = err
			if attempt < attempts {
				if waitErr := waitJSONRetry(ctx, attempt); waitErr != nil {
					return waitErr
				}
				continue
			}
			return err
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			if out == nil {
				_ = resp.Body.Close()
				return nil
			}
			err = json.NewDecoder(resp.Body).Decode(out)
			_ = resp.Body.Close()
			return err
		}

		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		_ = resp.Body.Close()
		lastErr = fmt.Errorf("saasplane: %s returned %s: %s", path, resp.Status, strings.TrimSpace(string(b)))
		if attempt < attempts && retryableHTTPStatus(resp.StatusCode) {
			if waitErr := waitJSONRetry(ctx, attempt); waitErr != nil {
				return waitErr
			}
			continue
		}
		return lastErr
	}
	return lastErr
}
