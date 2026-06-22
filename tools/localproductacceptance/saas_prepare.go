package main

func prepareSaaSDaemons(cfg config, client apiClient) saasPrepareResult {
	if !*cfg.prepareDaemon || !*cfg.runMutations {
		return saasPrepareResult{}
	}
	result := saasPrepareResult{}
	build := buildLocalDaemonBinary(*cfg.daemonBinary)
	result.Scenarios = append(result.Scenarios, build)
	if build.Status != statusPassed {
		return result
	}
	for index := 1; index <= normalizedDaemonSlots(*cfg.daemonSlots); index++ {
		slot, scenarios := prepareSaaSDaemonSlot(cfg, client, index)
		result.Scenarios = append(result.Scenarios, scenarios...)
		if slot.Credential.DeviceID != "" {
			result.DeviceIDs = append(result.DeviceIDs, slot.Credential.DeviceID)
		}
		if startScenarioPassed(scenarios) {
			result.Slots = append(result.Slots, slot)
		}
	}
	return result
}

func waitForPreparedSaaSRuntimes(cfg config, client apiClient, prep saasPrepareResult) saasPrepareResult {
	if len(prep.DeviceIDs) == 0 {
		return prep
	}
	sc, payload := waitForDeviceRuntimeSnapshot(cfg, client, prep.DeviceIDs)
	prep.Scenarios = append(prep.Scenarios, sc)
	prep.Runtimes = preparedRuntimesFromDevices(payload, prep.DeviceIDs)
	return prep
}

func normalizedDaemonSlots(value int) int {
	if value < 2 {
		return 2
	}
	return value
}

func startScenarioPassed(scenarios []scenario) bool {
	return len(scenarios) > 1 && scenarios[1].Status == statusPassed
}
