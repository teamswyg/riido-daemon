package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
)

const releaseVersion = "v-local-qa"

func newInstallFixture() (installFixture, func(), error) {
	root, err := os.MkdirTemp("", "riido-release-acceptance-*")
	if err != nil {
		return installFixture{}, nil, fmt.Errorf("create fixture: %w", err)
	}
	cleanup := func() { _ = os.RemoveAll(root) }
	fixture := installFixture{
		assetDir:   filepath.Join(root, "assets"),
		binDir:     filepath.Join(root, "bin"),
		installDir: filepath.Join(root, "install"),
		marker:     filepath.Join(root, "install-marker"),
	}
	if err := writeFixtureFiles(fixture); err != nil {
		cleanup()
		return installFixture{}, nil, err
	}
	return fixture, cleanup, nil
}

func writeFixtureFiles(fixture installFixture) error {
	for _, dir := range []string{fixture.assetDir, fixture.binDir, fixture.installDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	if err := writeArchiveAndSums(fixture.assetDir); err != nil {
		return err
	}
	if err := writeExecutable(filepath.Join(fixture.binDir, "curl"), fakeCurlScript()); err != nil {
		return err
	}
	if err := writeExecutable(filepath.Join(fixture.binDir, "install"), fakeInstallScript()); err != nil {
		return err
	}
	return writeExecutable(filepath.Join(fixture.binDir, "uname"), fakeUnameScript())
}

func writeArchiveAndSums(assetDir string) error {
	asset := filepath.Join(assetDir, "riido-daemon_darwin_arm64.tar.gz")
	if err := writeDaemonArchive(asset); err != nil {
		return err
	}
	bytes, err := os.ReadFile(asset)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(bytes)
	body := fmt.Sprintf("%x  riido-daemon_darwin_arm64.tar.gz\n", sum)
	return os.WriteFile(filepath.Join(assetDir, "SHA256SUMS"), []byte(body), 0o644)
}
