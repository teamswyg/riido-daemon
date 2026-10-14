package workdir

// Adapter is the port. The supervisor calls Prepare per claimed task and
// InjectRuntimeConfig before handing the workdir to the runtime actor.
type Adapter interface {
	Prepare(TaskID) (Workspace, error)
	InjectRuntimeConfig(Workspace, RuntimeConfig) error
}
