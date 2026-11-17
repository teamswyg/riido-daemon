package codex

import (
	"errors"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (s *sharedWriter) waitingForReconnect() bool {
	if s == nil || s.done || s.startRequested || s.turnID != "" {
		return false
	}
	// Confirmed-connect failure cleanup is part of waiting, too. Termination
	// clears startError so a cancellation cleanup cannot extend idle grace.
	cleanup := s.startError != "" && errors.Is(s.startFailure, errSharedOwnerConnect)
	return (!s.stopping && (s.reconnecting || !s.retryAt.IsZero())) || cleanup
}

func (d *protocolDriver) waitSharedReconnect() []agentbridge.Event {
	s := d.shared
	s.reconnecting = false
	if s.stopping || s.done {
		return nil
	}
	if s.retryDelay == 0 {
		s.retryDelay = 2 * time.Second
	} else {
		s.retryDelay = min(2*s.retryDelay, 10*time.Second)
	}
	s.retryAt = time.Now().Add(s.retryDelay)
	if s.waitAnnounced {
		return nil
	}
	s.waitAnnounced = true
	return []agentbridge.Event{{Kind: agentbridge.EventProgress, Text: "Codex의 기존 대화 연결을 기다리고 있습니다. 연결이 복구되면 이 요청을 자동으로 이어갑니다. 답글을 다시 보내지 않아도 됩니다."}}
}
