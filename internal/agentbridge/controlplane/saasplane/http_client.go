package saasplane

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
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

func (p *Plane) attachAuthHeaders(req *http.Request) {
	if p.cfg.DeviceSecret != "" {
		req.Header.Set("X-Riido-Device-Id", p.cfg.DeviceID)
		req.Header.Set("X-Riido-Device-Secret", p.cfg.DeviceSecret)
	}
	if p.cfg.BearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+p.cfg.BearerToken)
	}
}

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

func (p *Plane) saveAssignmentRuntime(ctx context.Context, assignment assignmentcontract.Assignment, runtimeID string) error {
	return p.withState(ctx, func(s *planeState) {
		executionID := assignmentExecutionID(assignment)
		s.assignmentsByExecution[executionID] = assignment
		if runtimeID != "" {
			s.runtimeIDsByExecution[executionID] = runtimeID
		}
	})
}

func (p *Plane) assignmentForExecution(ctx context.Context, executionID string) (assignmentcontract.Assignment, bool, error) {
	var assignment assignmentcontract.Assignment
	var ok bool
	err := p.withState(ctx, func(s *planeState) {
		assignment, ok = s.assignmentsByExecution[executionID]
	})
	return assignment, ok, err
}

func (p *Plane) agentBindings(ctx context.Context) ([]assignmentcontract.AgentRuntimeBinding, error) {
	now := time.Now()
	cached, ok, err := p.cachedAgentBindings(ctx, now)
	if err != nil || ok {
		return cached, err
	}
	var out AgentRuntimeBindingListResponse
	if err := p.getJSON(ctx, "/v1/daemon/agent-bindings", &out); err != nil {
		return nil, err
	}
	bindings := cloneAgentRuntimeBindings(out.Bindings)
	_ = p.withState(ctx, func(s *planeState) {
		s.agentBindingsCache = cloneAgentRuntimeBindings(bindings)
		s.agentBindingsCachedAt = now
	})
	return bindings, nil
}

func (p *Plane) cachedAgentBindings(ctx context.Context, now time.Time) ([]assignmentcontract.AgentRuntimeBinding, bool, error) {
	var bindings []assignmentcontract.AgentRuntimeBinding
	var ok bool
	err := p.withState(ctx, func(s *planeState) {
		if s.agentBindingsCachedAt.IsZero() || now.Sub(s.agentBindingsCachedAt) >= agentBindingCacheTTL {
			return
		}
		bindings = cloneAgentRuntimeBindings(s.agentBindingsCache)
		ok = true
	})
	return bindings, ok, err
}

func (p *Plane) invalidateAgentBindingsCache(ctx context.Context) {
	_ = p.withState(ctx, func(s *planeState) {
		s.agentBindingsCache = nil
		s.agentBindingsCachedAt = time.Time{}
	})
}

func cloneAgentRuntimeBindings(in []assignmentcontract.AgentRuntimeBinding) []assignmentcontract.AgentRuntimeBinding {
	if len(in) == 0 {
		return nil
	}
	return append([]assignmentcontract.AgentRuntimeBinding(nil), in...)
}

func (p *Plane) activeAssignmentIDsForHeartbeat(ctx context.Context, agentID string, runningTaskIDs []string) ([]string, error) {
	var executions []string
	seen := map[string]bool{}
	for _, executionID := range runningTaskIDs {
		executionID = strings.TrimSpace(executionID)
		if executionID != "" && !seen[executionID] {
			seen[executionID] = true
			executions = append(executions, executionID)
		}
	}
	if len(executions) == 0 {
		return nil, nil
	}
	var ids []string
	err := p.withState(ctx, func(s *planeState) {
		for _, executionID := range executions {
			assignment := s.assignmentsByExecution[executionID]
			if assignment.AgentID != agentID || assignment.ID == "" {
				continue
			}
			ids = append(ids, assignment.ID)
		}
	})
	if err != nil {
		return nil, err
	}
	return ids, nil
}
