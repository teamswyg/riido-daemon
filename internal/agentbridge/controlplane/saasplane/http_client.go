package saasplane

import (
	"bytes"
	"context"
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
			return decodeJSONResponse(resp, out, method, path)
		}

		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		_ = resp.Body.Close()
		lastErr = httpStatusError{
			Path:       path,
			Status:     resp.Status,
			StatusCode: resp.StatusCode,
			Body:       strings.TrimSpace(string(b)),
		}
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
