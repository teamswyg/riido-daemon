package ingest

import (
	"time"

	"github.com/teamswyg/riido-contracts/ir"
)

func daemonTestConfig(sink *memorySink, now time.Time) Config {
	cfg := Config{
		Sink:                sink,
		RiidoDaemonVersion:  "riido-daemon.test.v1",
		PolicyBundleVersion: "policy-bundle.test.v1",
		ActorKind:           ir.ActorDaemon,
		ActorID:             "daemon-1",
		NewEventID: func(time.Time) (string, error) {
			return "018f0000-0000-7000-8000-000000000001", nil
		},
	}
	if !now.IsZero() {
		cfg.Now = func() time.Time { return now }
	}
	return cfg
}

func agentRedactionTestConfig(sink *memorySink) Config {
	return Config{
		Sink:                sink,
		RiidoDaemonVersion:  "riido-daemon.test.v1",
		PolicyBundleVersion: "policy-bundle.test.v1",
		ActorKind:           ir.ActorAgent,
		ActorID:             "run-1",
		NewEventID: sequentialEventIDs(
			"018f0000-0000-7000-8000-000000000101",
			"018f0000-0000-7000-8000-000000000102",
		),
	}
}
