package policy

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func LoadPolicyBundleFile(path string) (PolicyBundle, error) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return PolicyBundle{}, errors.New("policy: bundle path is required")
	}
	data, err := os.ReadFile(trimmed)
	if err != nil {
		return PolicyBundle{}, fmt.Errorf("policy: load bundle %s: %w", trimmed, err)
	}
	bundle, err := ParsePolicyBundleJSON(data)
	if err != nil {
		return PolicyBundle{}, fmt.Errorf("policy: load bundle %s: %w", trimmed, err)
	}
	return bundle, nil
}

func ParsePolicyBundleJSON(data []byte) (PolicyBundle, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var bundle PolicyBundle
	if err := dec.Decode(&bundle); err != nil {
		return PolicyBundle{}, fmt.Errorf("parse policy bundle: %w", err)
	}
	var extra any
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		return PolicyBundle{}, errors.New("parse policy bundle: trailing JSON value")
	}
	if err := bundle.Validate(); err != nil {
		return PolicyBundle{}, err
	}
	return bundle, nil
}
