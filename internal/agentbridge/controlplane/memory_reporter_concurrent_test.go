package controlplane

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestMemoryReporterConcurrentClaims(t *testing.T) {
	reporter := NewMemoryReporter()
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := range 16 {
		wg.Go(func() {
			<-start
			task := fmt.Sprintf("task-%d", i)
			_ = reporter.StartTask(context.Background(), task)
			for range 100 {
				_ = reporter.ReportEvent(context.Background(), task, agentbridge.Event{Kind: agentbridge.EventProgress, Text: "working"})
				_ = reporter.Recorded(task)
			}
			_ = reporter.CompleteTask(context.Background(), task, agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "done"})
		})
	}
	close(start)
	wg.Wait()
	for i := range 16 {
		record := reporter.Recorded(fmt.Sprintf("task-%d", i))
		if !record.Started || len(record.Events) != 100 || record.Result.Output != "done" {
			t.Fatalf("claim %d corrupted", i)
		}
	}
}
