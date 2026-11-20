package selfupdate

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// machO returns a minimal thin 64-bit Mach-O header for the CheckArchitecture rule.
func machO(cpu uint32) []byte {
	raw := make([]byte, 64)
	binary.LittleEndian.PutUint32(raw, 0xfeedfacf)
	binary.LittleEndian.PutUint32(raw[4:], cpu)
	binary.LittleEndian.PutUint32(raw[12:], 2)
	return raw
}

type entry struct {
	name string
	body []byte
	mode int64
	typ  byte
}

func packageEntries(version string, helper bool) []entry {
	entries := []entry{{name: "riido", body: machO(0x0100000c), mode: 0o755}, {name: "VERSION", body: []byte(version + "\n"), mode: 0o644}}
	if helper {
		info := `<plist><dict><key>CFBundleIdentifier</key><string>io.riido.cli-auth-helper</string><key>CFBundleExecutable</key><string>riido-cli-auth-helper</string><key>RiidoCLIReleaseVersion</key><string>` + version + `</string></dict></plist>`
		entries = append(entries,
			entry{name: "CLI_AUTH_HELPER.json", body: []byte(`{"schema_version":1,"release_version":"` + version + `","helper_bundle":"RiidoCLIAuthHelper.app"}`), mode: 0o644},
			entry{name: "RiidoCLIAuthHelper.app/", typ: tar.TypeDir, mode: 0o755},
			entry{name: "RiidoCLIAuthHelper.app/Contents/Info.plist", body: []byte(info), mode: 0o644},
			entry{name: "RiidoCLIAuthHelper.app/Contents/embedded.provisionprofile", body: []byte("profile"), mode: 0o644},
			entry{name: "RiidoCLIAuthHelper.app/Contents/MacOS/riido-cli-auth-helper", body: machO(0x0100000c), mode: 0o755},
		)
	}
	return entries
}

func tarball(t *testing.T, entries []entry) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for _, e := range entries {
		typ := e.typ
		if typ == 0 {
			typ = tar.TypeReg
		}
		if err := tw.WriteHeader(&tar.Header{Name: "./" + e.name, Mode: e.mode, Size: int64(len(e.body)), Typeflag: typ}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write(e.body); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func cdn(t *testing.T, channel string, archive []byte, sums string) (*httptest.Server, *int) {
	t.Helper()
	if sums == "" {
		sum := sha256.Sum256(archive)
		sums = hex.EncodeToString(sum[:]) + "  riido-daemon_darwin_arm64.tar.gz\n"
	}
	hits := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		switch r.URL.Path {
		case "/releases/" + cdnPrefix[channel] + "/ai-agent/SHA256SUMS":
			_, _ = w.Write([]byte(sums))
		case "/releases/" + cdnPrefix[channel] + "/ai-agent/riido-daemon_darwin_arm64.tar.gz":
			_, _ = w.Write(archive)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	return server, &hits
}

func updater(server *httptest.Server, verify func(context.Context, string, []string) error) Updater {
	return Updater{BaseURL: server.URL + "/releases", Client: server.Client(), GOOS: "darwin", GOARCH: "arm64", Verify: verify}
}

func installRoot(t *testing.T, version string) string {
	t.Helper()
	root := t.TempDir()
	for name, body := range map[string]string{"riido": "old-binary", "VERSION": version + "\n", "CLI_AUTH_HELPER.json": "old-manifest"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(root, "RiidoCLIAuthHelper.app", "Contents"), 0o755); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestIsReleaseVersion(t *testing.T) {
	for version, want := range map[string]bool{"v0.0.105-staging.riid5971": true, "v0.1.0": true, "production-riid6005-20261103": false, "dev": false, "v1.2": false} {
		if got := IsReleaseVersion(version); got != want {
			t.Errorf("%s: %v", version, got)
		}
	}
}

func TestStageAndInstallReplacesPackage(t *testing.T) {
	version := "v0.0.108-staging.riid6021"
	server, _ := cdn(t, ChannelStaging, tarball(t, packageEntries(version, true)), "")
	verified := ""
	u := updater(server, func(_ context.Context, bundle string, archs []string) error {
		verified = bundle
		if len(archs) != 1 || archs[0] != "arm64" {
			t.Errorf("archs %v", archs)
		}
		return nil
	})
	root := installRoot(t, "v0.0.105-staging.riid5971")
	staged, err := u.Stage(context.Background(), ChannelStaging, root)
	if err != nil {
		t.Fatal(err)
	}
	if staged.Version != version || !staged.HasHelper || !strings.HasPrefix(verified, root) {
		t.Fatalf("staged %+v verified %s", staged, verified)
	}
	if err = Install(staged, root); err != nil {
		t.Fatal(err)
	}
	staged.Cleanup()
	got, _ := os.ReadFile(filepath.Join(root, "VERSION"))
	binary, _ := os.ReadFile(filepath.Join(root, "riido"))
	manifest, _ := os.ReadFile(filepath.Join(root, "CLI_AUTH_HELPER.json"))
	if strings.TrimSpace(string(got)) != version || !bytes.Equal(binary, machO(0x0100000c)) || !strings.Contains(string(manifest), version) {
		t.Fatal("package not installed")
	}
	if _, err = os.Stat(filepath.Join(root, "RiidoCLIAuthHelper.app", "Contents", "MacOS", "riido-cli-auth-helper")); err != nil {
		t.Fatal("helper not installed", err)
	}
	leftovers, _ := filepath.Glob(filepath.Join(root, ".riido-update-*"))
	if len(leftovers) != 0 {
		t.Fatal("staging directory left behind")
	}
}

func TestStageFailuresLeaveInstallUntouched(t *testing.T) {
	version := "v0.0.108-staging.riid6021"
	good := tarball(t, packageEntries(version, true))
	badSum := strings.Repeat("0", 64) + "  riido-daemon_darwin_arm64.tar.gz\n"
	cases := map[string]struct {
		archive []byte
		sums    string
		verify  func(context.Context, string, []string) error
	}{
		"checksum mismatch":  {good, badSum, nil},
		"signature rejected": {good, "", func(context.Context, string, []string) error { return errors.New("bad signature") }},
		"symlink entry":      {tarball(t, append(packageEntries(version, false), entry{name: "LICENSE", typ: tar.TypeSymlink})), "", nil},
		"path traversal":     {tarball(t, append(packageEntries(version, false), entry{name: "../evil", body: []byte("x"), mode: 0o644})), "", nil},
		"unexpected file":    {tarball(t, append(packageEntries(version, false), entry{name: "payload.sh", body: []byte("x"), mode: 0o755})), "", nil},
		"helper version":     {tarball(t, append(packageEntries("v0.0.999", true)[2:], packageEntries(version, false)...)), "", nil},
		"wrong architecture": {tarball(t, []entry{{name: "riido", body: machO(0x01000007), mode: 0o755}, {name: "VERSION", body: []byte(version), mode: 0o644}}), "", nil},
		"not a release":      {tarball(t, packageEntries("production-riid6005-20261103", false)), "", nil},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server, _ := cdn(t, ChannelStaging, tc.archive, tc.sums)
			verify := tc.verify
			if verify == nil {
				verify = func(context.Context, string, []string) error { return nil }
			}
			root := installRoot(t, "v0.0.105-staging.riid5971")
			if _, err := updater(server, verify).Stage(context.Background(), ChannelStaging, root); !errors.Is(err, ErrInvalid) {
				t.Fatalf("err=%v", err)
			}
			binary, _ := os.ReadFile(filepath.Join(root, "riido"))
			leftovers, _ := filepath.Glob(filepath.Join(root, ".riido-update-*"))
			if string(binary) != "old-binary" || len(leftovers) != 0 {
				t.Fatal("failed update touched the install")
			}
		})
	}
}

func TestPackageWithoutHelperRemovesOldHelper(t *testing.T) {
	version := "v0.0.108"
	server, _ := cdn(t, ChannelLatest, tarball(t, packageEntries(version, false)), "")
	root := installRoot(t, "v0.0.102")
	staged, err := updater(server, nil).Stage(context.Background(), ChannelLatest, root)
	if err != nil || staged.HasHelper {
		t.Fatal(err)
	}
	defer staged.Cleanup()
	if err = Install(staged, root); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(filepath.Join(root, "RiidoCLIAuthHelper.app")); !os.IsNotExist(err) {
		t.Fatal("old helper left beside a helperless binary")
	}
}

func TestStageRejectsUnsupportedPlatformAndInsecureBase(t *testing.T) {
	root := t.TempDir()
	if _, err := (Updater{BaseURL: DefaultBaseURL, GOOS: "linux", GOARCH: "amd64"}).Stage(context.Background(), ChannelLatest, root); !errors.Is(err, ErrUnsupported) {
		t.Fatal(err)
	}
	if _, err := (Updater{BaseURL: "http://cdn.example/releases", GOOS: "darwin", GOARCH: "arm64"}).Stage(context.Background(), ChannelLatest, root); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
	if _, err := (Updater{BaseURL: DefaultBaseURL, GOOS: "darwin", GOARCH: "arm64"}).Stage(context.Background(), "nightly", root); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
}

func TestIsOlder(t *testing.T) {
	cases := []struct {
		candidate, current string
		want               bool
	}{
		{"v0.0.102", "v0.0.107-staging.riid6014", true},
		{"v0.0.108-staging.riid6021", "v0.0.107-staging.riid6014", false},
		{"v0.0.107-staging.riid6021", "v0.0.107-staging.riid6014", false},
		{"v0.1.0", "v0.0.999", false},
		{"production-riid6005-20261103", "v0.0.107", false},
		{"v0.0.1", "dev", false},
	}
	for _, tc := range cases {
		if got := IsOlder(tc.candidate, tc.current); got != tc.want {
			t.Errorf("IsOlder(%q, %q) = %v", tc.candidate, tc.current, got)
		}
	}
}

// RIID-6036: public channel names read the active feeds, never the frozen
// legacy latest/staging prefixes.
func TestChannelFeedsAreTheActiveCDNPrefixes(t *testing.T) {
	u := Updater{BaseURL: DefaultBaseURL}
	for channel, want := range map[string]string{
		ChannelLatest:  "https://cdn.riido.io/releases/safe/ai-agent/SHA256SUMS",
		ChannelStaging: "https://cdn.riido.io/releases/staging-safe/ai-agent/SHA256SUMS",
	} {
		if got, err := u.channelURL(channel, "SHA256SUMS"); err != nil || got != want {
			t.Fatalf("%s: %s %v", channel, got, err)
		}
	}
}
