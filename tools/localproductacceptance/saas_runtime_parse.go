package main

func preparedRuntimesFromDevices(payload map[string]any, deviceIDs []string) []preparedRuntime {
	allowed := stringSet(deviceIDs)
	devices, _ := payload["devices"].([]any)
	var out []preparedRuntime
	for _, item := range devices {
		device, _ := item.(map[string]any)
		deviceID := stringValue(device["device_id"])
		if _, ok := allowed[deviceID]; !ok {
			continue
		}
		out = append(out, preparedRuntimesForDevice(deviceID, device)...)
	}
	return out
}

func preparedRuntimesForDevice(deviceID string, device map[string]any) []preparedRuntime {
	raw, _ := device["runtimes"].([]any)
	out := make([]preparedRuntime, 0, len(raw))
	for _, item := range raw {
		runtime, _ := item.(map[string]any)
		if !runtimeIsOnline(runtime) {
			continue
		}
		out = append(out, preparedRuntime{
			DeviceID:        deviceID,
			RuntimeID:       stringValue(runtime["runtime_id"]),
			Kind:            stringValue(runtime["kind"]),
			ProviderVersion: stringValue(runtime["provider_version"]),
		})
	}
	return out
}

func runtimeIsOnline(runtime map[string]any) bool {
	return stringValue(runtime["availability"]) == "online" &&
		stringValue(runtime["detection_state"]) == "detected"
}
