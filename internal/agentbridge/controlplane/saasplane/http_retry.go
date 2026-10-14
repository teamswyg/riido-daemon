package saasplane

import (
	"context"
	"net/http"
	"strings"
	"time"
)

func retryableJSONRequest(method, path string) bool {
	if method == http.MethodGet {
		return true
	}
	if method != http.MethodPost {
		return false
	}
	return strings.HasSuffix(path, "/poll") ||
		strings.HasSuffix(path, "/heartbeat") ||
		strings.HasSuffix(path, "/events") ||
		strings.Contains(path, "/tool-approvals") ||
		path == "/v1/daemon/runtime-snapshot"
}

func retryableHTTPStatus(status int) bool {
	switch status {
	case http.StatusRequestTimeout,
		http.StatusTooManyRequests,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func waitJSONRetry(ctx context.Context, attempt int) error {
	wait := time.Duration(attempt) * jsonRequestRetryBase
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
