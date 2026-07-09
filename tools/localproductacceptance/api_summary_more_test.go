package main

import "testing"

func TestSummarizeBootstrapAndDevices(t *testing.T) {
	bootstrap := summarizeBootstrap(map[string]any{
		"workspace_id": "ws_1",
		"agents":       []any{"a", "b"},
		"devices":      []any{"d"},
	})
	if bootstrap["agents_count"] != 2 || bootstrap["devices_count"] != 1 {
		t.Fatalf("unexpected bootstrap summary: %#v", bootstrap)
	}
	if bootstrap["workspace_id_present"] != true {
		t.Fatalf("workspace id should be present: %#v", bootstrap)
	}
	devices := summarizeDevices(map[string]any{"devices": []any{
		map[string]any{"runtimes": []any{
			map[string]any{"provider_version": "1.2.3"},
			map[string]any{"provider_version": ""},
		}},
		map[string]any{"runtimes": []any{map[string]any{"kind": "codex"}}},
	}})
	if devices["devices_count"] != 2 || devices["runtimes_count"] != 3 {
		t.Fatalf("unexpected devices summary: %#v", devices)
	}
	if devices["provider_version_field_present"] != true {
		t.Fatalf("provider_version field should be observed: %#v", devices)
	}
	if devices["provider_version_value_present"] != true {
		t.Fatalf("provider_version value should be observed: %#v", devices)
	}
}

func TestSummarizeUploadIntentAndHelpers(t *testing.T) {
	summary := summarizeUploadIntent(map[string]any{
		"method":                   "POST",
		"form_fields":              []any{"key", "policy"},
		"form_file_field":          "file",
		"upload_url":               "https://bucket.s3.amazonaws.com/",
		"profile_thumbnail_url":    "https://cdn.riido.io/a.png",
		"max_content_length_bytes": float64(128),
	})
	if summary["upload_host"] != "bucket.s3.amazonaws.com" {
		t.Fatalf("unexpected upload host: %#v", summary)
	}
	if summary["profile_thumbnail_host"] != "cdn.riido.io" {
		t.Fatalf("unexpected thumbnail host: %#v", summary)
	}
	if summary["form_fields_count"] != 2 || summary["form_file_field"] != "file" {
		t.Fatalf("unexpected form summary: %#v", summary)
	}
	if arrayLen("not-array") != 0 || stringPresent(3) {
		t.Fatal("helper coercion should be conservative")
	}
	if safeHost("://broken") != "" {
		t.Fatal("invalid URL host should be empty")
	}
}
