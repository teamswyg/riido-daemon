package supervisor

import (
	"bytes"
	"context"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestTimingSummaryRequiresTerminalACKAndEmitsOnce(t *testing.T) {
	var output bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&output)
	t.Cleanup(func() { log.SetOutput(previous) })
	reporter := &cancellationDeliveryReporter{failTerminal: true}
	a, task, tasks := cancellationDeliveryFixture(reporter)
	deliverCancellationTestResult(a, task, tasks, agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "answer-secret"})
	if strings.Contains(output.String(), "agent_execution_timing") {
		t.Fatal("unacknowledged terminal produced a successful timing summary")
	}
	reporter.failTerminal = false
	a.retryTerminalReports(context.Background(), tasks)
	a.retryTerminalReports(context.Background(), tasks)
	if strings.Count(output.String(), "agent_execution_timing") != 1 || len(tasks) != 0 {
		t.Fatal("terminal ACK did not emit exactly one summary")
	}
	if task.timing.reportAttempts != 2 || task.timing.reportFailures != 1 || strings.Contains(output.String(), "answer-secret") {
		t.Fatal("terminal retry accounting or privacy failed")
	}
}

func TestTimingContinuationCannotClaimSingleNativeTurn(t *testing.T) {
	start := time.Now()
	m := responseMetrics{providerResults: 2, providerStartedAt: start, providerFinishedAt: start.Add(time.Second)}
	fields := m.summary("codex", "v0.0.94", agentbridge.ResultCompleted, start.Add(2*time.Second))
	if fields["provider_elapsed_ms"] != int64(-1) || fields["native_turn_ms"] != int64(-1) || fields["provider_result_count"] != int64(2) {
		t.Fatal("explicit continuation reported as one native model turn")
	}
	if fields["provider_clock_scope"] != "adapter_including_backpressure" {
		t.Fatal("adapter interval mislabeled as native inference")
	}
}
