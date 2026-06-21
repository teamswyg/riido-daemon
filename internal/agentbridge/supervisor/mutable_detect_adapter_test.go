package supervisor

import (
	"context"
	"sync"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

type mutableDetectAdapter struct {
	name string
	mu   sync.Mutex
	res  agentbridge.DetectResult
}

func newMutableDetectAdapter(name, version string) *mutableDetectAdapter {
	return &mutableDetectAdapter{name: name, res: driftDetectResult(name, version)}
}

func (a *mutableDetectAdapter) Name() string { return a.name }

func (a *mutableDetectAdapter) Detect(
	context.Context,
	agentbridge.DetectEnv,
) (agentbridge.DetectResult, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.res, nil
}

func (a *mutableDetectAdapter) setVersion(version string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.res = driftDetectResult(a.name, version)
}

func (a *mutableDetectAdapter) BuildStart(req agentbridge.StartRequest) (agentbridge.StartCommand, error) {
	return (&stubAdapter{name: a.name}).BuildStart(req)
}

func (a *mutableDetectAdapter) NewParser() agentbridge.Parser {
	return &stubParser{}
}

func (a *mutableDetectAdapter) Translate(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
	return (&stubAdapter{name: a.name}).Translate(raw)
}

func (a *mutableDetectAdapter) BlockedArgs() []string { return nil }

func driftDetectResult(name, version string) agentbridge.DetectResult {
	return agentbridge.DetectResult{Available: true, Version: version, Executable: name}
}
