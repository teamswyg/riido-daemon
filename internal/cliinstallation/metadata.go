// Package cliinstallation shares release metadata rules with local CLI diagnostics.
package cliinstallation

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strings"
)

const (
	BundleName      = "RiidoCLIAuthHelper.app"
	ExecutableName  = "riido-cli-auth-helper"
	ManifestName    = "CLI_AUTH_HELPER.json"
	MaxBundleBytes  = int64(64 << 20)
	MaxArchiveBytes = int64(512 << 20)
)

var releaseVersion = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._+\-]{0,127}$`)

type Manifest struct {
	SchemaVersion  int    `json:"schema_version"`
	ReleaseVersion string `json:"release_version"`
	HelperBundle   string `json:"helper_bundle"`
}
type plistValue struct {
	XMLName  xml.Name
	Text     string       `xml:",chardata"`
	Children []plistValue `xml:",any"`
}

func CheckInfo(raw []byte, version string) error {
	if len(raw) > 1<<20 {
		return errors.New("helper Info.plist too large")
	}
	var root plistValue
	dec := xml.NewDecoder(bytes.NewReader(raw))
	if err := dec.Decode(&root); err != nil {
		return errors.New("helper Info.plist must be XML")
	}
	if root.XMLName.Local != "plist" || len(root.Children) != 1 || root.Children[0].XMLName.Local != "dict" {
		return errors.New("invalid helper Info.plist")
	}
	values := map[string]plistValue{}
	nodes := root.Children[0].Children
	if len(nodes)%2 != 0 {
		return errors.New("invalid helper Info.plist dictionary")
	}
	for i := 0; i < len(nodes); i += 2 {
		key := nodes[i]
		if key.XMLName.Local != "key" {
			return errors.New("invalid helper Info.plist key")
		}
		if _, ok := values[key.Text]; ok {
			return errors.New("duplicate helper Info.plist key")
		}
		values[key.Text] = nodes[i+1]
	}
	for key, want := range map[string]string{"CFBundleIdentifier": "io.riido.cli-auth-helper", "CFBundleExecutable": ExecutableName, "RiidoCLIReleaseVersion": version} {
		v := values[key]
		if v.XMLName.Local != "string" || v.Text != want {
			return fmt.Errorf("helper %s mismatch", key)
		}
	}
	return nil
}

func CheckManifest(raw []byte, version string) error {
	if len(raw) > 8192 {
		return errors.New("helper manifest too large")
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	token, err := dec.Token()
	if err != nil || token != json.Delim('{') {
		return errors.New("invalid helper manifest")
	}
	values := map[string]json.RawMessage{}
	for dec.More() {
		keyToken, e := dec.Token()
		key, ok := keyToken.(string)
		if e != nil || !ok {
			return errors.New("invalid manifest field")
		}
		if _, ok = values[key]; ok {
			return errors.New("duplicate manifest field")
		}
		var value json.RawMessage
		if e = dec.Decode(&value); e != nil {
			return e
		}
		values[key] = value
	}
	if _, err = dec.Token(); err != nil {
		return err
	}
	if _, err = dec.Token(); err != io.EOF {
		return errors.New("trailing manifest data")
	}
	if len(values) != 3 {
		return errors.New("unknown or missing manifest field")
	}
	var m Manifest
	if json.Unmarshal(raw, &m) != nil || m.SchemaVersion != 1 || m.ReleaseVersion != version || m.HelperBundle != BundleName {
		return errors.New("helper manifest mismatch")
	}
	return nil
}

func machoArchitectures(raw []byte) ([]string, error) {
	if len(raw) < 32 {
		return nil, errors.New("invalid Mach-O executable")
	}
	magic := binary.BigEndian.Uint32(raw)
	if magic == 0xcafebabe || magic == 0xcafebabf {
		count := binary.BigEndian.Uint32(raw[4:])
		size := 20
		if magic == 0xcafebabf {
			size = 32
		}
		if count == 0 || count > 8 || uint64(8+int(count)*size) > uint64(len(raw)) {
			return nil, errors.New("invalid universal executable")
		}
		archs := []string{}
		seen := map[string]bool{}
		for i := 0; i < int(count); i++ {
			h := raw[8+i*size:]
			offset, n := uint64(binary.BigEndian.Uint32(h[8:])), uint64(binary.BigEndian.Uint32(h[12:]))
			if size == 32 {
				offset = binary.BigEndian.Uint64(h[8:])
				n = binary.BigEndian.Uint64(h[16:])
			}
			if offset > uint64(len(raw)) || n > uint64(len(raw))-offset || n < 32 {
				return nil, errors.New("invalid universal slice")
			}
			a, e := thinArchitecture(raw[offset : offset+n])
			if e != nil || seen[a] {
				return nil, errors.New("invalid or duplicate universal architecture")
			}
			cpu := binary.BigEndian.Uint32(h)
			if cpu != cpuForArch(a) {
				return nil, errors.New("universal architecture mismatch")
			}
			seen[a] = true
			archs = append(archs, a)
		}
		return archs, nil
	}
	arch, err := thinArchitecture(raw)
	if err != nil {
		return nil, err
	}
	return []string{arch}, nil
}

func thinArchitecture(raw []byte) (string, error) {
	if len(raw) < 32 || binary.LittleEndian.Uint32(raw) != 0xfeedfacf || binary.LittleEndian.Uint32(raw[12:]) != 2 {
		return "", errors.New("helper must be a 64-bit Mach-O executable")
	}
	switch binary.LittleEndian.Uint32(raw[4:]) {
	case 0x0100000c:
		return "arm64", nil
	case 0x01000007:
		return "x86_64", nil
	default:
		return "", errors.New("unsupported Mach-O architecture")
	}
}

func cpuForArch(arch string) uint32 {
	if arch == "arm64" {
		return 0x0100000c
	}
	return 0x01000007
}

func CheckArchitecture(raw []byte, arch string) ([]string, error) {
	archs, err := machoArchitectures(raw)
	if err != nil {
		return nil, err
	}
	target := arch
	if target == "amd64" {
		target = "x86_64"
	}
	if arch != "arm64" && arch != "amd64" {
		return nil, errors.New("unsupported target architecture")
	}
	if slices.Contains(archs, target) {
		return archs, nil
	}
	return nil, errors.New("helper architecture mismatch")
}

func ValidateVersion(version string) error {
	if !releaseVersion.MatchString(version) || strings.TrimSpace(version) != version {
		return errors.New("invalid release version")
	}
	return nil
}
