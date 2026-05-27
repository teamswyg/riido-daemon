package workdir

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/teamswyg/riido-contracts/ir"
)

// RunEventSink appends CanonicalEvent JSONL records under a run's ir/
// directory. It is a filesystem adapter for the EventIngestor Sink port.
type RunEventSink struct {
	path string
}

func NewRunEventSink(ws Workspace) (*RunEventSink, error) {
	if strings.TrimSpace(ws.IR) == "" {
		return nil, errors.New("workdir: ir dir is required")
	}
	return &RunEventSink{path: filepath.Join(ws.IR, "events.jsonl")}, nil
}

func (s *RunEventSink) AppendEvent(ctx context.Context, ev ir.CanonicalEvent) error {
	return s.AppendEvents(ctx, []ir.CanonicalEvent{ev})
}

func (s *RunEventSink) AppendEvents(ctx context.Context, events []ir.CanonicalEvent) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if len(events) == 0 {
		return nil
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	for _, ev := range events {
		if err := enc.Encode(ev); err != nil {
			return fmt.Errorf("workdir: encode ir event: %w", err)
		}
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("workdir: create ir event dir: %w", err)
	}
	f, err := os.OpenFile(s.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("workdir: open ir event log: %w", err)
	}
	if _, err := f.Write(buf.Bytes()); err != nil {
		_ = f.Close()
		return fmt.Errorf("workdir: write ir event log: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("workdir: close ir event log: %w", err)
	}
	return nil
}

func (s *RunEventSink) Path() string { return s.path }
