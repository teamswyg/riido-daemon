package main

import (
	"net/http"
	"time"
)

func waitForDeviceRuntimeSnapshot(cfg config, client apiClient, deviceIDs []string) (scenario, map[string]any) {
	base := workspaceBase(*cfg.workspaceID)
	var payload map[string]any
	var statusCode int
	var err error
	deadline := time.Now().Add(45 * time.Second)
	for time.Now().Before(deadline) {
		payload, statusCode, err = client.call(http.MethodGet, base+"/devices", nil)
		if err == nil && preparedRuntimesReady(payload, deviceIDs) {
			return passedRuntimeWaitScenario(base, statusCode, payload, deviceIDs), payload
		}
		time.Sleep(time.Second)
	}
	return failedRuntimeWaitScenario(base, statusCode, err, payload, deviceIDs), payload
}

func preparedRuntimesReady(payload map[string]any, deviceIDs []string) bool {
	runtimes := preparedRuntimesFromDevices(payload, deviceIDs)
	pair, ok := choosePreparedRuntimePair(runtimes)
	return ok && pair[0].ProviderVersion != "" && pair[1].ProviderVersion != ""
}

func passedRuntimeWaitScenario(base string, status int, payload map[string]any, deviceIDs []string) scenario {
	return scenario{
		ID:       "local.saas.runtime_snapshot.wait",
		Status:   statusPassed,
		Method:   http.MethodGet,
		Endpoint: base + "/devices",
		Observed: runtimeWaitObserved(status, payload, deviceIDs),
	}
}

func failedRuntimeWaitScenario(base string, status int, err error, payload map[string]any, deviceIDs []string) scenario {
	sc := passedRuntimeWaitScenario(base, status, payload, deviceIDs)
	sc.Status = statusFailed
	sc.FailureSummary = "prepared SaaS daemons did not publish two same-provider runtimes"
	if err != nil {
		sc.FailureSummary = err.Error()
	}
	return sc
}
