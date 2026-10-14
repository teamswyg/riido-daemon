package workdir

// TaskID is the per-task identity that determines the workdir tree path.
type TaskID struct {
	Workspace string
	Task      string
	Run       string
}

// Workspace is the result of Prepare: the on-disk tree paths.
type Workspace struct {
	Root         string // <root>/<workspace>/tasks/<task>/runs/<run>/
	Workdir      string // <root>/<workspace>/tasks/<task>/runs/<run>/workdir/
	Output       string // <root>/<workspace>/tasks/<task>/runs/<run>/output/
	Logs         string // <root>/<workspace>/tasks/<task>/runs/<run>/logs/
	Artifacts    string // <root>/<workspace>/tasks/<task>/runs/<run>/artifacts/
	NativeConfig string // <root>/<workspace>/tasks/<task>/runs/<run>/native-config/
	IR           string // <root>/<workspace>/tasks/<task>/runs/<run>/ir/
}
