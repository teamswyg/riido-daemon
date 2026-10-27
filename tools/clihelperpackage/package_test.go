package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrepareBundleCopiesVerifiedPairAndExactVersion(t *testing.T) {
	src := helperFixture(t, "v1.2.3-qa", "arm64")
	out := t.TempDir()
	err := prepareBundle(context.Background(), src, out, "v1.2.3-qa", "arm64", fixtureSignature)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(out, "CLI_AUTH_HELPER.json"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "{\"schema_version\":1,\"release_version\":\"v1.2.3-qa\",\"helper_bundle\":\"RiidoCLIAuthHelper.app\"}\n" {
		t.Fatalf("manifest=%s", got)
	}
	data, err := os.ReadFile(filepath.Join(out, "RiidoCLIAuthHelper.app/Contents/embedded.provisionprofile"))
	if err != nil || string(data) != "fixture-profile" {
		t.Fatalf("profile lost: %v", err)
	}
}

func TestPrepareBundleRejectsInvalidInputs(t *testing.T) {
	cases := []struct {
		name          string
		mutate        func(string)
		version, arch string
	}{
		{"version mismatch", nil, "v9", "arm64"},
		{"wrong arch", nil, "v1.2.3", "amd64"},
		{"wrong identifier", func(p string) { replacePlist(t, p, "io.riido.cli-auth-helper", "io.other") }, "v1.2.3", "arm64"},
		{"wrong executable", func(p string) { replacePlist(t, p, "riido-cli-auth-helper</string>", "other</string>") }, "v1.2.3", "arm64"},
		{"missing profile", func(p string) {
			if err := os.Remove(filepath.Join(p, "Contents/embedded.provisionprofile")); err != nil {
				t.Fatal(err)
			}
		}, "v1.2.3", "arm64"},
		{"symlink", func(p string) {
			if err := os.Symlink("/tmp", filepath.Join(p, "Contents/escape")); err != nil {
				t.Fatal(err)
			}
		}, "v1.2.3", "arm64"},
		{"non executable", func(p string) {
			if err := os.Chmod(filepath.Join(p, "Contents/MacOS/riido-cli-auth-helper"), 0o644); err != nil {
				t.Fatal(err)
			}
		}, "v1.2.3", "arm64"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			src := helperFixture(t, "v1.2.3", "arm64")
			if tt.mutate != nil {
				tt.mutate(src)
			}
			out := t.TempDir()
			if err := prepareBundle(context.Background(), src, out, tt.version, tt.arch, fixtureSignature); err == nil {
				t.Fatal("invalid helper accepted")
			}
			if _, err := os.Stat(filepath.Join(out, "CLI_AUTH_HELPER.json")); !os.IsNotExist(err) {
				t.Fatal("manifest published before validation")
			}
		})
	}
}

func TestPrepareBundleRejectsSignatureFailure(t *testing.T) {
	src := helperFixture(t, "v1.2.3", "arm64")
	out := t.TempDir()
	fail := func(context.Context, string, []string) error { return errors.New("bad team or signature") }
	if err := prepareBundle(context.Background(), src, out, "v1.2.3", "arm64", fail); err == nil {
		t.Fatal("invalid signature accepted")
	}
	if _, err := os.Stat(filepath.Join(out, "CLI_AUTH_HELPER.json")); !os.IsNotExist(err) {
		t.Fatal("manifest published after signature failure")
	}
}

func fixtureSignature(_ context.Context, bundle string, archs []string) error {
	if filepath.Base(bundle) != "RiidoCLIAuthHelper.app" || len(archs) != 1 || archs[0] != "arm64" {
		return errors.New("wrong verification target")
	}
	return nil
}

func helperFixture(t *testing.T, version, arch string) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "RiidoCLIAuthHelper.app")
	if err := os.MkdirAll(filepath.Join(root, "Contents/MacOS"), 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string][]byte{"Contents/Info.plist": []byte(helperPlist(version)), "Contents/embedded.provisionprofile": []byte("fixture-profile"), "Contents/MacOS/riido-cli-auth-helper": machOFixture(arch)}
	for name, data := range files {
		mode := os.FileMode(0o644)
		if strings.Contains(name, "MacOS/") {
			mode = 0o755
		}
		if err := os.WriteFile(filepath.Join(root, name), data, mode); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func helperPlist(version string) string {
	return `<?xml version="1.0"?><plist version="1.0"><dict><key>CFBundleIdentifier</key><string>io.riido.cli-auth-helper</string><key>CFBundleExecutable</key><string>riido-cli-auth-helper</string><key>RiidoCLIReleaseVersion</key><string>` + version + `</string></dict></plist>`
}

func replacePlist(t *testing.T, bundle, old, replacement string) {
	t.Helper()
	p := filepath.Join(bundle, "Contents/Info.plist")
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(strings.ReplaceAll(string(data), old, replacement)), 0o644); err != nil {
		t.Fatal(err)
	}
}

func machOFixture(arch string) []byte {
	data := make([]byte, 32)
	binary.LittleEndian.PutUint32(data, 0xfeedfacf)
	cpu := uint32(0x0100000c)
	if arch == "amd64" {
		cpu = 0x01000007
	}
	binary.LittleEndian.PutUint32(data[4:], cpu)
	binary.LittleEndian.PutUint32(data[12:], 2)
	return data
}

type archiveEntry struct {
	name string
	body []byte
	kind byte
	mode int64
	link string
}

func packageEntries(version string, modern bool) []archiveEntry {
	entries := []archiveEntry{{"./VERSION", []byte(version + "\n"), tar.TypeReg, 0o644, ""}, {"./riido", machOFixture("arm64"), tar.TypeReg, 0o755, ""}}
	if modern {
		entries = append(entries, archiveEntry{"CLI_AUTH_HELPER.json", []byte(`{"schema_version":1,"release_version":"` + version + `","helper_bundle":"RiidoCLIAuthHelper.app"}`), tar.TypeReg, 0o644, ""}, archiveEntry{"RiidoCLIAuthHelper.app/Contents/Info.plist", []byte(helperPlist(version)), tar.TypeReg, 0o644, ""}, archiveEntry{"RiidoCLIAuthHelper.app/Contents/MacOS/riido-cli-auth-helper", machOFixture("arm64"), tar.TypeReg, 0o755, ""}, archiveEntry{"RiidoCLIAuthHelper.app/Contents/embedded.provisionprofile", []byte("fixture-profile"), tar.TypeReg, 0o644, ""})
	}
	return entries
}

func writeArchive(t *testing.T, entries []archiveEntry) string {
	t.Helper()
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	tw := tar.NewWriter(gz)
	for _, entry := range entries {
		if err := tw.WriteHeader(&tar.Header{Name: entry.name, Mode: entry.mode, Size: int64(len(entry.body)), Typeflag: entry.kind, Linkname: entry.link}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write(entry.body); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(t.TempDir(), "asset.tar.gz")
	if err := os.WriteFile(p, b.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestVerifyArchiveLegacyAndModern(t *testing.T) {
	for _, modern := range []bool{false, true} {
		if err := verifyArchive(writeArchive(t, packageEntries("v1.2.3", modern)), "v1.2.3", "arm64"); err != nil {
			t.Fatal(err)
		}
	}
}

func TestVerifyArchiveRejectsMalformedPairAndUnsafePaths(t *testing.T) {
	cases := map[string]func([]archiveEntry) []archiveEntry{
		"missing manifest": func(e []archiveEntry) []archiveEntry { return append(e[:2], e[3:]...) },
		"missing bundle":   func(e []archiveEntry) []archiveEntry { return e[:3] },
		"version mismatch": func(e []archiveEntry) []archiveEntry { e[3].body = []byte(helperPlist("v2")); return e },
		"wrong manifest path": func(e []archiveEntry) []archiveEntry {
			e[2].body = bytes.ReplaceAll(e[2].body, []byte("RiidoCLIAuthHelper.app"), []byte("../Other.app"))
			return e
		},
		"duplicate manifest key": func(e []archiveEntry) []archiveEntry {
			e[2].body = bytes.ReplaceAll(e[2].body, []byte(`"schema_version":1`), []byte(`"schema_version":2,"schema_version":1`))
			return e
		},
		"wrong schema": func(e []archiveEntry) []archiveEntry {
			e[2].body = bytes.ReplaceAll(e[2].body, []byte(`"schema_version":1`), []byte(`"schema_version":2`))
			return e
		},
		"duplicate path": func(e []archiveEntry) []archiveEntry { return append(e, e[2]) },
		"case collision": func(e []archiveEntry) []archiveEntry {
			return append(e, archiveEntry{"RiidoCLIAuthHelper.app/Contents/info.plist", []byte("other"), tar.TypeReg, 0o644, ""})
		},
		"parent escape": func(e []archiveEntry) []archiveEntry {
			return append(e, archiveEntry{"../escape", []byte("no"), tar.TypeReg, 0o644, ""})
		},
		"absolute path": func(e []archiveEntry) []archiveEntry {
			return append(e, archiveEntry{"/escape", []byte("no"), tar.TypeReg, 0o644, ""})
		},
		"symbolic link": func(e []archiveEntry) []archiveEntry {
			return append(e, archiveEntry{"RiidoCLIAuthHelper.app/Contents/link", nil, tar.TypeSymlink, 0o777, "/tmp"})
		},
		"special file": func(e []archiveEntry) []archiveEntry {
			return append(e, archiveEntry{"RiidoCLIAuthHelper.app/Contents/pipe", nil, tar.TypeFifo, 0o644, ""})
		},
		"wrong binary arch": func(e []archiveEntry) []archiveEntry { e[4].body = machOFixture("amd64"); return e },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			if err := verifyArchive(writeArchive(t, mutate(packageEntries("v1.2.3", true))), "v1.2.3", "arm64"); err == nil {
				t.Fatal("unsafe archive accepted")
			}
		})
	}
}

func TestPrepareBundlePreservesEmptySignedResource(t *testing.T) {
	src := helperFixture(t, "v1.2.3", "arm64")
	if err := os.WriteFile(filepath.Join(src, "Contents/empty-resource"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	out := t.TempDir()
	if err := prepareBundle(context.Background(), src, out, "v1.2.3", "arm64", fixtureSignature); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(filepath.Join(out, "RiidoCLIAuthHelper.app/Contents/empty-resource")); err != nil || len(data) != 0 {
		t.Fatalf("empty resource lost: %v", err)
	}
}

func TestArchiveRoundtripRechecksCopiedHelper(t *testing.T) {
	archive := writeArchive(t, packageEntries("v1.2.3", true))
	if err := recheckArchive(context.Background(), archive, "v1.2.3", "arm64", fixtureSignature); err != nil {
		t.Fatal(err)
	}
	if err := recheckArchive(context.Background(), archive, "v1.2.3", "arm64", func(context.Context, string, []string) error { return errors.New("invalid copied signature") }); err == nil {
		t.Fatal("archive published without rechecking signature")
	}
}

func TestUniversal64ChecksAllArchitectures(t *testing.T) {
	raw := make([]byte, 136)
	binary.BigEndian.PutUint32(raw, 0xcafebabf)
	binary.BigEndian.PutUint32(raw[4:], 2)
	for i, cpu := range []uint32{0x0100000c, 0x01000007} {
		h := raw[8+i*32:]
		binary.BigEndian.PutUint32(h, cpu)
		binary.BigEndian.PutUint64(h[8:], uint64(72+i*32))
		binary.BigEndian.PutUint64(h[16:], 32)
	}
	copy(raw[72:], machOFixture("arm64"))
	copy(raw[104:], machOFixture("amd64"))
	archs, err := checkArchitecture(raw, "amd64")
	if err != nil || strings.Join(archs, ",") != "arm64,x86_64" {
		t.Fatalf("architectures=%v err=%v", archs, err)
	}
	binary.BigEndian.PutUint64(raw[48:], 9999)
	if _, err = checkArchitecture(raw, "arm64"); err == nil {
		t.Fatal("invalid universal offset accepted")
	}
}

func TestArchiveRejectsTrailingPayloadAfterTarEnd(t *testing.T) {
	archive := writeArchive(t, packageEntries("v1.2.3", false))
	input, err := os.Open(archive)
	if err != nil {
		t.Fatal(err)
	}
	gz, err := gzip.NewReader(input)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := io.ReadAll(gz)
	if err != nil {
		t.Fatal(err)
	}
	if err = gz.Close(); err != nil {
		t.Fatal(err)
	}
	if err = input.Close(); err != nil {
		t.Fatal(err)
	}
	var body bytes.Buffer
	writer := gzip.NewWriter(&body)
	if _, err = writer.Write(append(raw, []byte("hidden nonzero tar payload")...)); err != nil {
		t.Fatal(err)
	}
	if err = writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(archive, body.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	if err = verifyArchive(archive, "v1.2.3", "arm64"); err == nil {
		t.Fatal("uninspected trailing archive payload accepted")
	}
}
