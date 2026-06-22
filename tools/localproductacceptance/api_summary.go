package main

import "net/url"

func summarizeBootstrap(payload map[string]any) map[string]any {
	agents := arrayLen(payload["agents"])
	devices := arrayLen(payload["devices"])
	return map[string]any{
		"agents_count":         agents,
		"devices_count":        devices,
		"workspace_id_present": stringPresent(payload["workspace_id"]),
	}
}

func summarizeDevices(payload map[string]any) map[string]any {
	devices, _ := payload["devices"].([]any)
	runtimeCount := 0
	providerVersionFieldPresent := false
	providerVersionPresent := false
	for _, device := range devices {
		record, _ := device.(map[string]any)
		runtimes, _ := record["runtimes"].([]any)
		runtimeCount += len(runtimes)
		for _, runtime := range runtimes {
			rt, _ := runtime.(map[string]any)
			_, hasProviderVersion := rt["provider_version"]
			providerVersionFieldPresent = providerVersionFieldPresent || hasProviderVersion
			providerVersionPresent = providerVersionPresent || stringPresent(rt["provider_version"])
		}
	}
	return map[string]any{
		"devices_count":                  len(devices),
		"runtimes_count":                 runtimeCount,
		"provider_version_field_present": providerVersionFieldPresent,
		"provider_version_value_present": providerVersionPresent,
	}
}

func summarizeUploadIntent(payload map[string]any) map[string]any {
	uploadURL, _ := payload["upload_url"].(string)
	thumbnailURL, _ := payload["profile_thumbnail_url"].(string)
	return map[string]any{
		"method":                   payload["method"],
		"form_fields_count":        arrayLen(payload["form_fields"]),
		"form_file_field":          payload["form_file_field"],
		"upload_host":              safeHost(uploadURL),
		"profile_thumbnail_host":   safeHost(thumbnailURL),
		"max_content_length_bytes": payload["max_content_length_bytes"],
	}
}

func arrayLen(value any) int {
	items, _ := value.([]any)
	return len(items)
}

func stringPresent(value any) bool {
	text, _ := value.(string)
	return text != ""
}

func safeHost(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return parsed.Host
}
