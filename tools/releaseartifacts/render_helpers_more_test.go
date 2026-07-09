package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAssetAndInstallerRenderingHelpers(t *testing.T) {
	if got := assetName(target{GOOS: "windows", GOARCH: "amd64", Format: "zip"}); got != "riido-daemon_windows_amd64.zip" {
		t.Fatalf("zip asset=%q", got)
	}
	if got := assetName(target{GOOS: "darwin", GOARCH: "arm64", Format: "tar.gz"}); got != "riido-daemon_darwin_arm64.tar.gz" {
		t.Fatalf("tar asset=%q", got)
	}
	facts := installerFacts(manifest{Installer: installer{
		SupportedGOOS:     []string{"darwin"},
		SupportedGOARCH:   []string{"arm64"},
		DefaultInstallDir: "/usr/local/bin",
		InstallDirEnv:     "RIIDO_INSTALL_DIR",
	}})
	if len(facts) != 5 || !strings.Contains(facts[0], "`darwin`") {
		t.Fatalf("installer facts=%#v", facts)
	}
	flow := desktopFlow(manifest{DesktopMSIX: desktopMSIX{
		DownloadSource: "https://cdn.riido.io",
		StorageRoot:    "~/Library/Application Support/Riido",
	}})
	if len(flow) != 5 || !strings.Contains(flow[1], "https://cdn.riido.io") {
		t.Fatalf("desktop flow=%#v", flow)
	}
}

func mustReleaseWrite(t *testing.T, path string, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
