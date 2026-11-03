package scripts_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallRiidoDaemonInstallsHelperPackageBesideBinary(t *testing.T) {
	fixture := newInstallFixture(t)
	writePackageAssets(t, fixture, helperPackage("v-test"))
	installDir := t.TempDir()
	out, err := runInstaller(t, fixture, installDir, "RIIDO_DAEMON_VERSION=v-test")
	if err != nil {
		t.Fatalf("install failed: %v\n%s", err, out)
	}
	if got := mustRead(t, filepath.Join(installDir, "VERSION")); got != "v-test\n" {
		t.Fatalf("VERSION = %q", got)
	}
	if got := mustRead(t, filepath.Join(installDir, "RiidoCLIAuthHelper.app/Contents/Info.plist")); got != "new helper\n" {
		t.Fatalf("helper Info.plist = %q", got)
	}
	if _, err := os.Stat(filepath.Join(installDir, "CLI_AUTH_HELPER.json")); err != nil {
		t.Fatalf("helper manifest missing: %v", err)
	}
	for _, want := range []string{"riido auth login", "PATH에 추가하세요", ".zshrc"} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
}

func TestInstallRiidoDaemonReplacesOldHelperAndRemovesStaleOne(t *testing.T) {
	fixture := newInstallFixture(t)
	installDir := t.TempDir()
	oldHelper := filepath.Join(installDir, "RiidoCLIAuthHelper.app/Contents")
	if err := os.MkdirAll(oldHelper, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(oldHelper, "old-only"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(installDir, "CLI_AUTH_HELPER.json"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	writePackageAssets(t, fixture, helperPackage("v-test"))
	if out, err := runInstaller(t, fixture, installDir, "RIIDO_DAEMON_VERSION=v-test"); err != nil {
		t.Fatalf("install failed: %v\n%s", err, out)
	}
	if _, err := os.Stat(filepath.Join(oldHelper, "old-only")); !os.IsNotExist(err) {
		t.Fatalf("old helper file survived replacement: %v", err)
	}
	// A legacy package (no helper) must not leave the previous helper behind.
	writePackageAssets(t, fixture, []packageEntry{{"riido", "new daemon\n"}, {"VERSION", "v-legacy\n"}})
	out, err := runInstaller(t, fixture, installDir, "RIIDO_DAEMON_VERSION=v-legacy")
	if err != nil {
		t.Fatalf("legacy install failed: %v\n%s", err, out)
	}
	for _, name := range []string{"RiidoCLIAuthHelper.app", "CLI_AUTH_HELPER.json"} {
		if _, err := os.Lstat(filepath.Join(installDir, name)); !os.IsNotExist(err) {
			t.Fatalf("%s survived a legacy install: %v", name, err)
		}
	}
	if strings.Contains(out, "riido auth login") {
		t.Fatalf("legacy install suggested login without a helper:\n%s", out)
	}
}

func TestInstallRiidoDaemonRejectsIncompleteHelperBeforeTouchingBinary(t *testing.T) {
	fixture := newInstallFixture(t)
	installDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(installDir, "riido"), []byte("old daemon\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	writePackageAssets(t, fixture, []packageEntry{{"riido", "new daemon\n"}, {"VERSION", "v-test\n"}, {"CLI_AUTH_HELPER.json", "{}"}})
	out, err := runInstaller(t, fixture, installDir, "RIIDO_DAEMON_VERSION=v-test")
	if err == nil || !strings.Contains(out, "incomplete RiidoCLIAuthHelper.app") {
		t.Fatalf("incomplete helper accepted: err=%v\n%s", err, out)
	}
	if got := mustRead(t, filepath.Join(installDir, "riido")); got != "old daemon\n" {
		t.Fatalf("existing binary changed: %q", got)
	}
}

func TestInstallRiidoDaemonStagingChannelUsesCDNPackageVersion(t *testing.T) {
	fixture := newInstallFixture(t)
	writePackageAssets(t, fixture, helperPackage("v0.0.103-staging.cli"))
	installDir := t.TempDir()
	out, err := runInstaller(t, fixture, installDir, "RIIDO_DAEMON_CHANNEL=staging")
	if err != nil {
		t.Fatalf("staging install failed: %v\n%s", err, out)
	}
	for _, want := range []string{"riido-daemon version: v0.0.103-staging.cli", "riido-daemon channel: staging"} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
	log := mustRead(t, filepath.Join(fixture.assetDir, "curl.log"))
	if !strings.Contains(log, "https://cdn.riido.io/releases/staging/ai-agent/riido-daemon_darwin_arm64.tar.gz") || strings.Contains(log, "api.github.com") {
		t.Fatalf("staging channel used the wrong source:\n%s", log)
	}
}

func TestInstallRiidoDaemonRejectsInvalidChannelInput(t *testing.T) {
	fixture := newInstallFixture(t)
	for name, env := range map[string][]string{
		"unknown channel":     {"RIIDO_DAEMON_CHANNEL=nightly"},
		"channel and version": {"RIIDO_DAEMON_CHANNEL=staging", "RIIDO_DAEMON_VERSION=v1"},
		"plain http CDN":      {"RIIDO_DAEMON_CHANNEL=staging", "RIIDO_DAEMON_CDN_BASE=http://cdn.example"},
	} {
		installDir := t.TempDir()
		if out, err := runInstaller(t, fixture, installDir, env...); err == nil {
			t.Fatalf("%s: accepted\n%s", name, out)
		}
		if _, err := os.Stat(filepath.Join(installDir, "riido")); !os.IsNotExist(err) {
			t.Fatalf("%s: installed a binary", name)
		}
	}
}

func TestInstallRiidoDaemonWarnsWhenAnotherRiidoShadowsInstall(t *testing.T) {
	fixture := newInstallFixture(t)
	installDir := t.TempDir()
	shadowDir := t.TempDir()
	writeExecutable(t, filepath.Join(shadowDir, "riido"), "#!/bin/sh\n")
	path := fixture.binDir + string(os.PathListSeparator) + shadowDir + string(os.PathListSeparator) + installDir + string(os.PathListSeparator) + os.Getenv("PATH")
	out, err := runInstaller(t, fixture, installDir, "RIIDO_DAEMON_VERSION=v-test", "PATH="+path)
	if err != nil {
		t.Fatalf("install failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "먼저 찾는 riido는 "+filepath.Join(shadowDir, "riido")) || strings.Contains(out, "PATH에 추가하세요") {
		t.Fatalf("shadow warning missing:\n%s", out)
	}
}
