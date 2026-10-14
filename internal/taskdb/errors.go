package taskdb

import "github.com/teamswyg/riido-daemon/pkg/failure"

const taskDBErrorLayer failure.Layer = "taskdb"

var (
	ErrTaskDBInput       = failure.NewSentinel(taskDBErrorLayer, "input")
	ErrTaskDBState       = failure.NewSentinel(taskDBErrorLayer, "state")
	ErrTaskDBGuard       = failure.NewSentinel(taskDBErrorLayer, "guard")
	ErrTaskDBReplay      = failure.NewSentinel(taskDBErrorLayer, "replay")
	ErrTaskDBPersistence = failure.NewSentinel(taskDBErrorLayer, "persistence")
	ErrTaskDBSchema      = failure.NewSentinel(taskDBErrorLayer, "schema")
)

func taskDBErrorf(kind failure.Sentinel, op, format string, args ...any) error {
	return failure.New(kind, op, failure.Format(format, args...))
}

func taskDBWrapf(kind failure.Sentinel, op string, cause error, format string, args ...any) error {
	return failure.Wrap(kind, op, failure.Format(format, args...), cause)
}
