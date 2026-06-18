package supervisor

import (
	"sort"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
)

func reportContextFor(req *bridge.TaskRequest) controlplane.TaskReportContext {
	report, _ := controlplane.TaskReportContextFromMetadata(req.Metadata)
	return report
}

func findCapability(caps []runtimeactor.Capability, provider string) (runtimeactor.Capability, bool) {
	for _, capView := range caps {
		if capView.Provider == provider {
			return capView, true
		}
	}
	return runtimeactor.Capability{}, false
}

func runtimeTaskIDs(tasks []runtimeactor.TaskStatus) []string {
	ids := make([]string, 0, len(tasks))
	for _, task := range tasks {
		if task.TaskID != "" {
			ids = append(ids, task.TaskID)
		}
	}
	sort.Strings(ids)
	return ids
}
