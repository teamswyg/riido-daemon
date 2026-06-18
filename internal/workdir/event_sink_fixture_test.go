package workdir

import (
	"time"

	"github.com/teamswyg/riido-contracts/ir"
)

func testCanonicalEvent(id string) ir.CanonicalEvent {
	return ir.CanonicalEvent{
		EventID:             id,
		OccurredAt:          time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC),
		EventSchemaVersion:  1,
		Scope:               ir.EventScopeTask,
		Type:                ir.EventTaskCreated,
		ActorKind:           ir.ActorDaemon,
		ActorID:             "daemon-1",
		RiidoDaemonVersion:  "riido-daemon.test.v1",
		PolicyBundleVersion: "policy-bundle.test.v1",
		TaskID:              "task-1",
		FSMVersion:          1,
	}
}
