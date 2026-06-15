package controlplane

import "github.com/teamswyg/riido-daemon/pkg/failure"

const controlPlaneErrorLayer failure.Layer = "controlplane"

var (
	ErrControlPlaneInput       = failure.NewSentinel(controlPlaneErrorLayer, "input")
	ErrControlPlaneRuntime     = failure.NewSentinel(controlPlaneErrorLayer, "runtime")
	ErrControlPlaneQueue       = failure.NewSentinel(controlPlaneErrorLayer, "queue")
	ErrControlPlaneReporter    = failure.NewSentinel(controlPlaneErrorLayer, "reporter")
	ErrControlPlaneRegistry    = failure.NewSentinel(controlPlaneErrorLayer, "registry")
	ErrControlPlanePersistence = failure.NewSentinel(controlPlaneErrorLayer, "persistence")
)

func controlPlaneErrorf(kind failure.Sentinel, op string, format string, args ...any) error {
	return failure.New(kind, op, failure.Format(format, args...))
}

func controlPlaneWrapf(kind failure.Sentinel, op string, cause error, format string, args ...any) error {
	return failure.Wrap(kind, op, failure.Format(format, args...), cause)
}
