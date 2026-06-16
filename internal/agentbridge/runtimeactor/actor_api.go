package runtimeactor

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func (a *Actor) buildHeartbeat(inFlight map[string]*runningTask) Heartbeat {
	ids := make([]string, 0, len(inFlight))
	for id := range inFlight {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return Heartbeat{
		RuntimeID:      a.cfg.RuntimeID,
		UptimeSeconds:  int64(a.cfg.Now().Sub(a.startedAt).Seconds()),
		DeviceName:     a.cfg.DeviceName,
		SlotLimit:      a.cfg.MaxConcurrent,
		SlotsInUse:     len(inFlight),
		RunningTaskIDs: ids,
	}
}

func (a *Actor) drainAndShutdown(level lifecycle.ShutdownLevel, inFlight map[string]*runningTask, completeCh <-chan string) {
	for _, t := range inFlight {
		t.handle.session.Cancel(ErrActorStopped)
	}
	if level.IsForced() {
		a.stopErrCh <- nil
		return
	}
	deadline := time.After(5 * time.Second)
	for len(inFlight) > 0 {
		select {
		case id := <-completeCh:
			delete(inFlight, id)
		case next := <-a.stopReqCh:
			if lifecycle.NormalizeShutdownLevel(next).IsForced() {
				a.stopErrCh <- nil
				return
			}
		case <-deadline:
			a.stopErrCh <- fmt.Errorf("runtimeactor: %d session(s) did not terminate", len(inFlight))
			return
		}
	}
	a.stopErrCh <- nil
}

// ----- Public methods (mailbox-only) -----

// Submit posts a TaskRequest to the actor. Returns a SessionHandle or
// a typed error.
//
// Note on the stoppedCh check inside the reply-wait select: the
// mailbox is buffered, so a send can succeed even after Stop has fully
// shut the actor down (the actor is no longer reading). Without the
// stoppedCh guard on the wait, callers would block forever waiting
// for a reply that will never be written. The same pattern applies to
// Cancel below.
func (a *Actor) Submit(ctx context.Context, req bridge.TaskRequest) (*SessionHandle, error) {
	reply := make(chan submitReply, 1)
	select {
	case a.mailbox <- envelope{submit: &submitMsg{ctx: ctx, req: req, reply: reply}}:
	case <-a.stoppedCh:
		return nil, ErrActorStopped
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	select {
	case res := <-reply:
		return res.handle, res.err
	case <-a.stoppedCh:
		return nil, ErrActorStopped
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Cancel asks the actor to cancel an in-flight task.
func (a *Actor) Cancel(ctx context.Context, taskID, reason string) error {
	reply := make(chan error, 1)
	select {
	case a.mailbox <- envelope{cancel: &cancelMsg{taskID: taskID, reason: reason, reply: reply}}:
	case <-a.stoppedCh:
		return ErrActorStopped
	case <-ctx.Done():
		return ctx.Err()
	}
	select {
	case err := <-reply:
		return err
	case <-a.stoppedCh:
		return ErrActorStopped
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Status returns a synchronous status snapshot.
func (a *Actor) Status(ctx context.Context) (Status, error) {
	reply := make(chan statusReply, 1)
	select {
	case a.statusCh <- reply:
	case <-a.stoppedCh:
		return Status{RuntimeID: a.cfg.RuntimeID, Health: "stopped"}, nil
	case <-ctx.Done():
		return Status{}, ctx.Err()
	}
	select {
	case res := <-reply:
		return res.status, nil
	case <-a.stoppedCh:
		return Status{RuntimeID: a.cfg.RuntimeID, Health: "stopped"}, nil
	case <-ctx.Done():
		return Status{}, ctx.Err()
	}
}

// HeartbeatPayload returns the publish-ready heartbeat.
func (a *Actor) HeartbeatPayload(ctx context.Context) (Heartbeat, error) {
	reply := make(chan statusReply, 1)
	select {
	case a.statusCh <- reply:
	case <-a.stoppedCh:
		return Heartbeat{RuntimeID: a.cfg.RuntimeID}, nil
	case <-ctx.Done():
		return Heartbeat{}, ctx.Err()
	}
	select {
	case res := <-reply:
		return res.hb, nil
	case <-a.stoppedCh:
		return Heartbeat{RuntimeID: a.cfg.RuntimeID}, nil
	case <-ctx.Done():
		return Heartbeat{}, ctx.Err()
	}
}

// ----- helpers -----

func indexAdapters(in []agentbridge.Adapter) map[string]agentbridge.Adapter {
	out := make(map[string]agentbridge.Adapter, len(in))
	for _, a := range in {
		out[a.Name()] = a
	}
	return out
}

func capabilityIndexForProvider(caps []Capability, provider string) int {
	for i, c := range caps {
		if c.Provider == provider {
			return i
		}
	}
	return -1
}

func metaProfile(meta map[string]string) string {
	if meta == nil {
		return ""
	}
	return meta["profile"]
}

func toProcessCommand(c agentbridge.StartCommand) process.Command {
	return process.Command{
		Executable: c.Executable,
		Args:       c.Args,
		Env:        c.Env,
		Dir:        c.Dir,
	}
}
