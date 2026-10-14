package main

func runtimeWaitObserved(status int, payload map[string]any, deviceIDs []string) map[string]any {
	runtimes := preparedRuntimesFromDevices(payload, deviceIDs)
	pair, ok := choosePreparedRuntimePair(runtimes)
	observed := map[string]any{
		"status_code":              status,
		"prepared_devices_count":   len(deviceIDs),
		"prepared_runtimes_count":  len(runtimes),
		"same_runtime_kind_pair":   ok,
		"provider_version_present": ok && pair[0].ProviderVersion != "" && pair[1].ProviderVersion != "",
	}
	if ok {
		observed["runtime_kind"] = pair[0].Kind
		observed["first_runtime_id"] = pair[0].RuntimeID
		observed["second_runtime_id"] = pair[1].RuntimeID
	}
	return observed
}

func stringSet(values []string) map[string]struct{} {
	out := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value != "" {
			out[value] = struct{}{}
		}
	}
	return out
}
