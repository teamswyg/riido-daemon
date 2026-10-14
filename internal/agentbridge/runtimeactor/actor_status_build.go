package runtimeactor

import "sort"

func (a *Actor) buildStatus(caps []Capability, inFlight map[string]*runningTask) Status {
	ids := make([]string, 0, len(inFlight))
	for id := range inFlight {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	tasks := make([]TaskStatus, 0, len(ids))
	for _, id := range ids {
		t := inFlight[id]
		tasks = append(tasks, TaskStatus{
			TaskID:   t.taskID,
			Provider: t.provider,
			State:    "running",
		})
	}
	return Status{
		RuntimeID:       a.cfg.RuntimeID,
		StartedAt:       a.startedAt,
		UptimeSeconds:   int64(a.cfg.Now().Sub(a.startedAt).Seconds()),
		Health:          "ok",
		Owner:           a.cfg.Owner,
		DeviceName:      a.cfg.DeviceName,
		Agents:          append([]AgentStatus(nil), a.cfg.Agents...),
		Models:          append([]RuntimeModel(nil), a.cfg.Models...),
		Capabilities:    append([]Capability(nil), caps...),
		MaxConcurrent:   a.cfg.MaxConcurrent,
		RunningSessions: len(inFlight),
		RunningTasks:    tasks,
	}
}
