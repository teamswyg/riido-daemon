package supervisor

import (
	"context"
	"sync"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

const maxProgressQueueBytes = 256 * 1024

const progressReportInterval = 100 * time.Millisecond

// One claim owns this cancellable display lane. The actor journals original
// events and the session owns complete output; neither waits for this HTTP lane.
type progressDelivery struct {
	mu         sync.Mutex
	queue      []agentbridge.Event
	bytes      int
	superseded int64
	wake       chan struct{}
	done       chan struct{}
	cancel     context.CancelFunc
	metrics    responseMetrics // worker-owned until done is closed
}

func newProgressDelivery(ctx context.Context, send func(context.Context, agentbridge.Event) error) *progressDelivery {
	ctx, cancel := context.WithCancel(ctx)
	d := &progressDelivery{wake: make(chan struct{}, 1), done: make(chan struct{}), cancel: cancel}
	go func() {
		defer close(d.done)
		var nextReport time.Time
		for {
			select {
			case <-ctx.Done():
				return
			case <-d.wake:
			}
			for ctx.Err() == nil {
				if !waitProgressReport(ctx, time.Until(nextReport)) {
					return
				}
				ev, ok := d.next()
				if !ok {
					break
				}
				d.metrics.reportProgress(ctx, func() error { return send(ctx, ev) })
				nextReport = time.Now().Add(progressReportInterval)
			}
		}
	}()
	return d
}

// Cancel HTTP before crossing a lifecycle/terminal boundary. ReportEvent's
// context contract makes this independent of the remote response deadline.
func (d *progressDelivery) stop() (responseMetrics, int64) {
	d.cancel()
	<-d.done
	d.mu.Lock()
	defer d.mu.Unlock()
	d.superseded += int64(len(d.queue))
	d.queue = nil
	d.bytes = 0
	return d.metrics, d.superseded
}

func (a *Actor) stopProgressDelivery(task *runningTask) {
	if task == nil || task.delivery == nil {
		return
	}
	m, superseded := task.delivery.stop()
	task.delivery = nil
	task.timing.mergeReports(m)
	task.timing.progressSuperseded += superseded
}

// An unused worker has no timers and can be interrupted by the claim context.
func (a *Actor) queueProgressDelivery(ctx context.Context, task *runningTask, ev agentbridge.Event) bool {
	if task.handle == nil || task.terminalResult != nil || !plainDeliveryEvent(ev) {
		return false
	}
	if task.delivery == nil {
		task.delivery = newProgressDelivery(ctx, func(ctx context.Context, ev agentbridge.Event) error {
			return a.cfg.Reporter.ReportEvent(ctx, task.taskID, ev)
		})
	}
	task.delivery.enqueue(ev)
	return true
}

func waitProgressReport(ctx context.Context, delay time.Duration) bool {
	if delay <= 0 {
		return ctx.Err() == nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
