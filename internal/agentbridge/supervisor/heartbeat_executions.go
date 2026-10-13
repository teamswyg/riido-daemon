package supervisor

import "sort"

// OwnedExecutions is an immutable observation for host activity checks. Native
// process slots and end-to-end assignment ownership are distinct observations.
func (a *Actor) OwnedExecutions() map[string][]string {
	a.heartbeatMu.RLock()
	defer a.heartbeatMu.RUnlock()
	out := make(map[string][]string, len(a.heartbeatExecutions))
	for runtime, ids := range a.heartbeatExecutions {
		out[runtime] = append([]string(nil), ids...)
	}
	return out
}

// The actor publishes immutable execution ownership snapshots. The independent
// heartbeat goroutine never reads the actor's mutable in-flight task map.
func (a *Actor) publishHeartbeatExecutions(inFlight map[string]*runningTask) {
	byRuntime := map[string][]string{}
	for id, task := range inFlight {
		byRuntime[task.runtimeID] = append(byRuntime[task.runtimeID], id)
	}
	a.heartbeatMu.Lock()
	a.heartbeatExecutions = byRuntime
	a.heartbeatMu.Unlock()
}

func (a *Actor) heartbeatTaskIDs(runtimeID string, nativeIDs []string) []string {
	seen := map[string]bool{}
	for _, id := range nativeIDs {
		seen[id] = true
	}
	a.heartbeatMu.RLock()
	for _, id := range a.heartbeatExecutions[runtimeID] {
		seen[id] = true
	}
	a.heartbeatMu.RUnlock()
	ids := make([]string, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
