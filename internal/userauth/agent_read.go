package userauth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"
)

var (
	ErrAgentReadUnsupported = errors.New("agent reads require a supported Riido API origin")
	ErrAgentReadIncomplete  = errors.New("agent directory exceeds the response limit")
)

// ReadTaskAgents holds one protected login until the caller has validated the
// public response. Neither raw OAuth nor a response-selected URL reaches CP.
func (m *Manager) ReadTaskAgents(ctx context.Context, workspace, task, agent, cursor string, consume func([]byte, int) error) error {
	if _, err := m.agentReadPreflight(workspace, task, agent, cursor, consume); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	return m.WithReadSession(ctx, func(session SessionReadClient) error {
		return session.ReadTaskAgents(ctx, workspace, task, agent, cursor, consume)
	})
}

func (m *Manager) agentReadPreflight(workspace, task, agent, cursor string, consume func([]byte, int) error) (string, error) {
	origin := threadReadOrigin(m.client)
	if origin == "" {
		return "", ErrAgentReadUnsupported
	}
	if consume == nil || !threadReadID(workspace) || !threadReadID(task) || (agent != "" && !threadReadID(agent)) || !agentReadCursor(cursor) || (agent != "" && cursor != "") {
		return "", failure("UNSAFE_REQUEST")
	}
	return origin, nil
}

func (s *sessionReadClient) ReadTaskAgents(ctx context.Context, workspace, task, agent, cursor string, consume func([]byte, int) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	m := s.manager
	if !s.active || m.generation.Load() != s.generation {
		return ErrLoggedOut
	}
	origin, err := m.agentReadPreflight(workspace, task, agent, cursor, consume)
	if err != nil {
		return err
	}
	credential, err := m.accessCredentialLocked(ctx, "", s.sessionID, s.generation)
	if err != nil {
		return err
	}
	body, err := json.Marshal(map[string]string{"workspace_id": workspace, "task_id": task})
	if err != nil {
		return failure("UNSAFE_REQUEST")
	}
	raw, status, err := m.threadReadRequest(ctx, http.MethodPost, m.client.config.Resource+"/oauth/cli/agent-read-tickets", credential.AccessToken, body, deviceLinkLimit)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return consume(nil, status)
	}
	ticket, err := decodeAgentReadTicket(raw, origin)
	if err != nil {
		return err
	}
	if err := s.checkLogin(ctx); err != nil {
		return err
	}
	endpoint := origin + "/v1/cli/workspaces/" + workspace + "/tasks/" + task + "/agents"
	if agent != "" {
		endpoint += "/" + agent
	} else if cursor != "" {
		endpoint += "?cursor=" + cursor
	}
	raw, status, err = m.threadReadRequest(ctx, http.MethodGet, endpoint, ticket, nil, threadReadBodyLimit)
	if errors.Is(err, ErrThreadReadIncomplete) {
		return ErrAgentReadIncomplete
	}
	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if m.generation.Load() != s.generation {
		return ErrLoggedOut
	}
	return consume(raw, status)
}

func agentReadCursor(cursor string) bool {
	if cursor == "" {
		return true
	}
	n, err := strconv.Atoi(cursor)
	return err == nil && n >= 0 && strconv.Itoa(n) == cursor
}
