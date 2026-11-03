package userauth

import (
	"strings"
	"time"
)

func decodeAgentReadTicket(raw []byte, origin string) (string, error) {
	fields, err := deviceLinkObject(raw, "schema_version", "access_token", "token_type", "resource", "purpose", "expires_at")
	if err != nil {
		return "", failure("INVALID_RESPONSE")
	}
	var ticket struct {
		SchemaVersion string `json:"schema_version"`
		AccessToken   string `json:"access_token"`
		TokenType     string `json:"token_type"`
		Resource      string `json:"resource"`
		Purpose       string `json:"purpose"`
		ExpiresAt     string `json:"expires_at"`
	}
	if decodeDeviceLink(fields, &ticket) != nil || ticket.SchemaVersion != "riido-cli-agent-read.v1" || ticket.Resource != origin || ticket.Purpose != "task_agents_read" || ticket.TokenType != "Bearer" || !strings.HasPrefix(ticket.AccessToken, "riido_cli_agent_v1_") || !deviceLinkRandom(strings.TrimPrefix(ticket.AccessToken, "riido_cli_agent_v1_")) {
		return "", failure("INVALID_RESPONSE")
	}
	expires, err := time.Parse(time.RFC3339Nano, ticket.ExpiresAt)
	now := time.Now()
	if err != nil || !expires.After(now) || expires.After(now.Add(65*time.Second)) {
		return "", failure("INVALID_RESPONSE")
	}
	return ticket.AccessToken, nil
}
