package supervisor

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

type watchCleanupSource struct {
	task           bridge.TaskRequest
	claimed        bool
	watchStarted   chan struct{}
	watchCtxClosed chan struct{}
}

func newWatchCleanupSource() *watchCleanupSource {
	return &watchCleanupSource{
		task:           bridge.TaskRequest{ID: "t-watch-cleanup", Provider: "fake", Prompt: "hello"},
		watchStarted:   make(chan struct{}),
		watchCtxClosed: make(chan struct{}),
	}
}

func (s *watchCleanupSource) RegisterRuntime(context.Context, controlplane.RuntimeRegistration) error {
	return nil
}

func (s *watchCleanupSource) DeregisterRuntime(context.Context, string) error { return nil }

func (s *watchCleanupSource) Heartbeat(context.Context, controlplane.RuntimeHeartbeat) error {
	return nil
}

func (s *watchCleanupSource) ClaimTask(context.Context, string) (*bridge.TaskRequest, error) {
	if s.claimed {
		return nil, nil
	}
	s.claimed = true
	task := s.task
	return &task, nil
}

func (s *watchCleanupSource) WatchCancellation(ctx context.Context, _ string) (<-chan error, error) {
	ch := make(chan error)
	close(s.watchStarted)
	go func() {
		<-ctx.Done()
		close(s.watchCtxClosed)
	}()
	return ch, nil
}
