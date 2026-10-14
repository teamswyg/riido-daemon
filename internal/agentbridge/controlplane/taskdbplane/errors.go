package taskdbplane

import "github.com/teamswyg/riido-daemon/pkg/failure"

const taskDBPlaneErrorLayer failure.Layer = "taskdbplane"

var (
	ErrTaskDBPlaneInput       = failure.NewSentinel(taskDBPlaneErrorLayer, "input")
	ErrTaskDBPlaneRuntime     = failure.NewSentinel(taskDBPlaneErrorLayer, "runtime")
	ErrTaskDBPlaneTaskState   = failure.NewSentinel(taskDBPlaneErrorLayer, "task-state")
	ErrTaskDBPlaneRegistry    = failure.NewSentinel(taskDBPlaneErrorLayer, "registry")
	ErrTaskDBPlaneLease       = failure.NewSentinel(taskDBPlaneErrorLayer, "lease")
	ErrTaskDBPlanePersistence = failure.NewSentinel(taskDBPlaneErrorLayer, "persistence")
)

func planeErrorf(kind failure.Sentinel, op, format string, args ...any) error {
	return failure.New(kind, op, failure.Format(format, args...))
}

func planeWrapf(kind failure.Sentinel, op string, cause error, format string, args ...any) error {
	return failure.Wrap(kind, op, failure.Format(format, args...), cause)
}
