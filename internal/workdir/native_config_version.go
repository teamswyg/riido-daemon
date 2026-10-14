package workdir

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// ComputeNativeConfigVersion returns the sha256-hex NativeConfigVersion
// defined by docs/20-domain/workspace.md §6 for the materialized native
// config tree.
func ComputeNativeConfigVersion(ws Workspace, input NativeConfigVersionInput) (string, error) {
	if err := validateNativeConfigVersionInput(ws, input); err != nil {
		return "", err
	}
	injected, err := injectedFileHashes(ws.NativeConfig)
	if err != nil {
		return "", err
	}
	if len(injected) == 0 {
		return "", errors.New("workdir: native-config has no injected files")
	}
	doc := nativeConfigVersionDoc{
		PolicyBundleVersion: input.PolicyBundleVersion,
		NativeConfigPlan: nativeConfigPlan{
			ProviderKind:  input.ProviderKind,
			ProtocolKind:  input.ProtocolKind,
			InjectedFiles: injected,
		},
		SchemaVersion: NativeConfigVersionSchemaVersion,
	}
	return hashNativeConfigVersionDoc(doc)
}

func validateNativeConfigVersionInput(ws Workspace, input NativeConfigVersionInput) error {
	switch {
	case strings.TrimSpace(ws.NativeConfig) == "":
		return errors.New("workdir: native-config dir is required")
	case strings.TrimSpace(input.PolicyBundleVersion) == "":
		return errors.New("workdir: policy bundle version is required")
	case strings.TrimSpace(input.ProviderKind) == "":
		return errors.New("workdir: provider kind is required")
	case strings.TrimSpace(input.ProtocolKind) == "":
		return errors.New("workdir: protocol kind is required")
	default:
		return nil
	}
}

func hashNativeConfigVersionDoc(doc nativeConfigVersionDoc) (string, error) {
	data, err := json.Marshal(doc)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum[:]), nil
}
