package runtimeactor

import "sort"

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
