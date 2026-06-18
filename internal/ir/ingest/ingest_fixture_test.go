package ingest

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
)

type memorySink struct {
	events      []ir.CanonicalEvent
	appendCalls int
	batchSizes  []int
}

func (s *memorySink) AppendEvent(_ context.Context, ev ir.CanonicalEvent) error {
	s.events = append(s.events, ev)
	return nil
}

func (s *memorySink) AppendEvents(_ context.Context, events []ir.CanonicalEvent) error {
	s.appendCalls++
	s.batchSizes = append(s.batchSizes, len(events))
	s.events = append(s.events, events...)
	return nil
}

func sequentialEventIDs(ids ...string) func(time.Time) (string, error) {
	next := 0
	return func(time.Time) (string, error) {
		if next >= len(ids) {
			return "", errors.New("no event id left")
		}
		id := ids[next]
		next++
		return id, nil
	}
}

func assertSinkBatch(t *testing.T, sink *memorySink, batchSize int) {
	t.Helper()
	if sink.appendCalls != 1 || !slices.Contains(sink.batchSizes, batchSize) {
		t.Fatalf("sink batches: calls=%d sizes=%v", sink.appendCalls, sink.batchSizes)
	}
}
