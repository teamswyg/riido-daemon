package workdir

import (
	"strings"
	"testing"
)

func TestValidateNativeConfigVersionInputRejectsMissingFields(t *testing.T) {
	validWS := Workspace{NativeConfig: t.TempDir()}
	valid := NativeConfigVersionInput{
		PolicyBundleVersion: "policy",
		ProviderKind:        "codex",
		ProtocolKind:        "codex-app-server",
	}
	cases := []struct {
		name  string
		ws    Workspace
		input NativeConfigVersionInput
		want  string
	}{
		{"native config", Workspace{}, valid, "native-config dir"},
		{"policy", validWS, NativeConfigVersionInput{ProviderKind: "codex", ProtocolKind: "codex"}, "policy bundle"},
		{"provider", validWS, NativeConfigVersionInput{PolicyBundleVersion: "policy", ProtocolKind: "codex"}, "provider kind"},
		{"protocol", validWS, NativeConfigVersionInput{PolicyBundleVersion: "policy", ProviderKind: "codex"}, "protocol kind"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateNativeConfigVersionInput(tc.ws, tc.input)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestComputeNativeConfigVersionRejectsEmptyNativeConfigTree(t *testing.T) {
	ws := Workspace{NativeConfig: t.TempDir()}
	input := NativeConfigVersionInput{
		PolicyBundleVersion: "policy",
		ProviderKind:        "codex",
		ProtocolKind:        "codex-app-server",
	}
	_, err := ComputeNativeConfigVersion(ws, input)
	if err == nil || !strings.Contains(err.Error(), "no injected files") {
		t.Fatalf("error = %v, want no injected files", err)
	}
}
