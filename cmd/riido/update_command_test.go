package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// updateTestCDN serves a helperless darwin/arm64 package of version.
func updateTestCDN(t *testing.T, version string) (*httptest.Server, *int) {
	t.Helper()
	mach := make([]byte, 64)
	binary.LittleEndian.PutUint32(mach, 0xfeedfacf)
	binary.LittleEndian.PutUint32(mach[4:], 0x0100000c)
	binary.LittleEndian.PutUint32(mach[12:], 2)
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for _, f := range []struct {
		name string
		body []byte
		mode int64
	}{{"riido", mach, 0o755}, {"VERSION", []byte(version + "\n"), 0o644}} {
		if err := tw.WriteHeader(&tar.Header{Name: "./" + f.name, Mode: f.mode, Size: int64(len(f.body)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		_, _ = tw.Write(f.body)
	}
	_ = tw.Close()
	_ = gz.Close()
	archive := buf.Bytes()
	sum := sha256.Sum256(archive)
	hits := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		switch {
		case strings.HasSuffix(r.URL.Path, "/staging/ai-agent/SHA256SUMS"):
			_, _ = w.Write([]byte(hex.EncodeToString(sum[:]) + "  riido-daemon_darwin_arm64.tar.gz\n"))
		case strings.HasSuffix(r.URL.Path, "/staging/ai-agent/riido-daemon_darwin_arm64.tar.gz"):
			_, _ = w.Write(archive)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	return server, &hits
}

func updateHarness(t *testing.T, vars map[string]string, current string) (*contextHarness, string, *int) {
	t.Helper()
	old := binaryVersion
	binaryVersion = current
	t.Cleanup(func() { binaryVersion = old })
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "riido"), []byte("old-binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "VERSION"), []byte(current+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	server, hits := updateTestCDN(t, "v0.0.108-staging.riid6021")
	h := newContextHarness(t, vars)
	h.io.executable = func() (string, error) { return filepath.Join(root, "riido"), nil }
	h.io.updateBaseURL = server.URL + "/releases"
	h.io.updateClient = server.Client()
	h.io.updateGOOS, h.io.updateGOARCH = "darwin", "arm64"
	h.io.verifyHelperSignature = func(context.Context, string, []string) error { return nil }
	return h, root, hits
}

func decodeUpdate(t *testing.T, raw []byte) updateData {
	t.Helper()
	var env struct {
		Data struct {
			Update updateData `json:"update"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatal(err, string(raw))
	}
	return env.Data.Update
}

func TestRunUpdateCheckThenInstall(t *testing.T) {
	h, root, _ := updateHarness(t, nil, "v0.0.105-staging.riid5971")
	if err := runUpdate(h.io, []string{"--check", "--json"}); err != nil {
		t.Fatal(err, h.stdout.String())
	}
	if got := decodeUpdate(t, h.stdout.Bytes()); got.LatestVersion != "v0.0.108-staging.riid6021" || !got.UpdateNeeded || got.Updated {
		t.Fatalf("check output %+v", got)
	}
	if body, _ := os.ReadFile(filepath.Join(root, "riido")); string(body) != "old-binary" {
		t.Fatal("--check replaced the binary")
	}
	h.stdout.Reset()
	if err := runUpdate(h.io, []string{"--json"}); err != nil {
		t.Fatal(err, h.stdout.String())
	}
	if got := decodeUpdate(t, h.stdout.Bytes()); !got.Updated {
		t.Fatalf("update output %+v", got)
	}
	if version, _ := os.ReadFile(filepath.Join(root, "VERSION")); strings.TrimSpace(string(version)) != "v0.0.108-staging.riid6021" {
		t.Fatal("VERSION not replaced")
	}
}

func TestRunUpdateAlreadyLatest(t *testing.T) {
	h, _, _ := updateHarness(t, nil, "v0.0.108-staging.riid6021")
	if err := runUpdate(h.io, nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(h.stdout.String(), "이미 최신 버전입니다") {
		t.Fatalf("output %q", h.stdout.String())
	}
}

func TestRunUpdateRefusesUnmanagedInstalls(t *testing.T) {
	cases := map[string]func(h *contextHarness, root string){
		"agent mode": func(h *contextHarness, _ string) {
			h.io.getenv = func(k string) string {
				if k == "RIIDO_RUN_TOKEN" {
					return "run-token"
				}
				return ""
			}
		},
		"app bundle": func(h *contextHarness, root string) {
			dir := filepath.Join(root, "Riido.app", "Contents", "MacOS")
			_ = os.MkdirAll(dir, 0o755)
			_ = os.WriteFile(filepath.Join(dir, "riido"), []byte("x"), 0o755)
			h.io.executable = func() (string, error) { return filepath.Join(dir, "riido"), nil }
		},
		"version mismatch": func(_ *contextHarness, root string) {
			_ = os.WriteFile(filepath.Join(root, "VERSION"), []byte("v0.0.1\n"), 0o644)
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			h, root, hits := updateHarness(t, nil, "v0.0.105-staging.riid5971")
			mutate(h, root)
			if err := runUpdate(h.io, []string{"--json"}); err == nil {
				t.Fatal("update accepted an unmanaged install")
			}
			if !strings.Contains(h.stdout.String(), `"code": "UNSUPPORTED"`) || *hits != 0 {
				t.Fatalf("output %s hits %d", h.stdout.String(), *hits)
			}
			if body, _ := os.ReadFile(filepath.Join(root, "riido")); string(body) != "old-binary" {
				t.Fatal("refused update touched the install")
			}
		})
	}
}

func TestRunUpdateRejectsUnknownChannel(t *testing.T) {
	h, _, hits := updateHarness(t, nil, "v0.0.105-staging.riid5971")
	if err := runUpdate(h.io, []string{"--channel", "nightly", "--json"}); err == nil || *hits != 0 {
		t.Fatal("unknown channel accepted")
	}
}
