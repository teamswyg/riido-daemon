package main

import (
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

type daemonObservation struct {
	runtimes []runtimeactor.Status
	metrics  daemonMetrics
	ready    bool
}

func (o daemonObservation) readyText() string {
	if o.ready {
		return "ready"
	}
	return "not-ready"
}

func observeDaemon(runtimes []*runtimeactor.Actor) daemonObservation {
	ctx, cancel := lifecycle.WithTimeout(lifecycle.Background(), 2*time.Second)
	defer cancel()

	obs := daemonObservation{
		runtimes: make([]runtimeactor.Status, 0, len(runtimes)),
		metrics:  daemonMetrics{RuntimeCount: len(runtimes)},
	}
	for _, rt := range runtimes {
		rtStatus, err := rt.Status(ctx.Context())
		if err != nil {
			continue
		}
		obs.observeRuntime(rtStatus)
	}
	obs.ready = obs.metrics.RuntimeResponding == obs.metrics.RuntimeCount
	return obs
}

func (o *daemonObservation) observeRuntime(rtStatus runtimeactor.Status) {
	o.runtimes = append(o.runtimes, rtStatus)
	o.metrics.RuntimeResponding++
	o.metrics.RunningTasks += rtStatus.RunningSessions
	for _, cap := range rtStatus.Capabilities {
		if cap.Available {
			o.metrics.ProviderAvailable++
		} else {
			o.metrics.ProviderUnavailable++
		}
	}
}
