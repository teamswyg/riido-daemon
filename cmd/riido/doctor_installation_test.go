package main

import (
	"context"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
)

func TestDoctorReportsResolvedExecutableAndPathConflict(t *testing.T) {
	root := t.TempDir()
	resolvedBinary := filepath.Join(root, "generation", "riido")
	if err := os.MkdirAll(filepath.Dir(resolvedBinary), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(resolvedBinary, []byte("fixture executable"), 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "riido")
	if err := os.Symlink(resolvedBinary, link); err != nil {
		t.Fatal(err)
	}
	conflict := filepath.Join(root, "other-riido")
	if err := os.WriteFile(conflict, []byte("other"), 0o755); err != nil {
		t.Fatal(err)
	}
	h := newContextHarness(t, nil)
	h.io.executable = func() (string, error) { return link, nil }
	h.io.lookPath = func(string) (string, error) { return conflict, nil }
	if err := runDoctor(h.io, []string{"--json"}); err != nil {
		t.Fatal(err)
	}
	checks := doctorChecks(t, decodeEnvelope(t, h.stdout.Bytes()))
	if checks["binary_target"] != "ok" || checks["path"] != "warn" || !strings.Contains(h.stdout.String(), resolvedBinary) {
		t.Fatalf("missing resolved target/conflict: %s", h.stdout.String())
	}
}

func TestDoctorRejectsPartialOrMismatchedModernPackage(t *testing.T) {
	for _, scenario := range []string{"manifest-only", "bundle-only", "wrong-version"} {
		t.Run(scenario, func(t *testing.T) {
			root := t.TempDir()
			exe := filepath.Join(root, "riido")
			if err := os.WriteFile(exe, []byte("fixture"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, "VERSION"), []byte("v1.2.3\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			if scenario != "bundle-only" {
				if err := os.WriteFile(filepath.Join(root, "CLI_AUTH_HELPER.json"), []byte(`{"schema_version":1,"release_version":"v9","helper_bundle":"RiidoCLIAuthHelper.app"}`), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			if scenario != "manifest-only" {
				if err := os.Mkdir(filepath.Join(root, "RiidoCLIAuthHelper.app"), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			h := newContextHarness(t, nil)
			h.io.executable = func() (string, error) { return exe, nil }
			h.io.cliVersion = func() string { return "riido-daemon v1.2.3" }
			if got := exitCodeOf(t, runDoctor(h.io, []string{"--json"})); got != 6 {
				t.Fatalf("malformed pair accepted: exit=%d %s", got, h.stdout.String())
			}
			data := decodeEnvelope(t, h.stdout.Bytes()).Data.(map[string]any)
			if data["api_verified"] != false || doctorChecks(t, decodeEnvelope(t, h.stdout.Bytes()))["helper_bundle"] != "fail" {
				t.Fatal("missing package failure")
			}
		})
	}
}

func doctorModernFixture(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	bundle := filepath.Join(root, "RiidoCLIAuthHelper.app")
	if err := os.MkdirAll(filepath.Join(bundle, "Contents/MacOS"), 0o755); err != nil {
		t.Fatal(err)
	}
	macho := make([]byte, 32)
	binary.LittleEndian.PutUint32(macho, 0xfeedfacf)
	cpu := uint32(0x01000007)
	if runtime.GOARCH == "arm64" {
		cpu = 0x0100000c
	}
	binary.LittleEndian.PutUint32(macho[4:], cpu)
	binary.LittleEndian.PutUint32(macho[12:], 2)
	for name, data := range map[string][]byte{
		"riido":                []byte("fixture CLI"),
		"VERSION":              []byte("v1.2.3\n"),
		"CLI_AUTH_HELPER.json": []byte(`{"schema_version":1,"release_version":"v1.2.3","helper_bundle":"RiidoCLIAuthHelper.app"}`),
		"RiidoCLIAuthHelper.app/Contents/Info.plist":                  []byte(`<plist><dict><key>CFBundleIdentifier</key><string>io.riido.cli-auth-helper</string><key>CFBundleExecutable</key><string>riido-cli-auth-helper</string><key>RiidoCLIReleaseVersion</key><string>v1.2.3</string></dict></plist>`),
		"RiidoCLIAuthHelper.app/Contents/embedded.provisionprofile":   []byte("fixture profile, not signed"),
		"RiidoCLIAuthHelper.app/Contents/MacOS/riido-cli-auth-helper": macho,
	} {
		if err := os.WriteFile(filepath.Join(root, name), data, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return filepath.Join(root, "riido"), bundle
}

func TestDoctorStaticSignatureDoesNotClaimAuthentication(t *testing.T) {
	for _, rejected := range []bool{false, true} {
		t.Run(map[bool]string{false: "valid", true: "rejected"}[rejected], func(t *testing.T) {
			exe, bundle := doctorModernFixture(t)
			calls := 0
			h := newContextHarness(t, nil)
			h.io.executable = func() (string, error) { return exe, nil }
			h.io.cliVersion = func() string { return "riido-daemon v1.2.3" }
			h.io.verifyHelperSignature = func(_ context.Context, actual string, archs []string) error {
				calls++
				if !sameFile(actual, bundle) || len(archs) != 1 {
					t.Fatal("wrong signature target")
				}
				if rejected {
					return errors.New("private codesign diagnostic")
				}
				return nil
			}
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) {
				t.Fatal("static check accessed user session")
				return nil, nil
			}
			err := runDoctor(h.io, []string{"--json"})
			if rejected {
				want := 6
				if runtime.GOOS != "darwin" {
					want = 8
				}
				if got := exitCodeOf(t, err); got != want {
					t.Fatalf("exit=%d", got)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			data := decodeEnvelope(t, h.stdout.Bytes()).Data.(map[string]any)
			if calls != 1 || data["api_verified"] != false || strings.Contains(h.stdout.String(), "private codesign diagnostic") {
				t.Fatalf("invalid static result: %s", h.stdout.String())
			}
			if !rejected && doctorChecks(t, decodeEnvelope(t, h.stdout.Bytes()))["helper_signature"] != "ok" {
				t.Fatal("missing static signature result")
			}
		})
	}
}
