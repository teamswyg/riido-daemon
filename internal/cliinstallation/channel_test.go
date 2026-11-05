package cliinstallation

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstalledChannelFollowsMarkerBesideResolvedBinary(t *testing.T) {
	root := t.TempDir()
	libexec := filepath.Join(root, "libexec")
	bin := filepath.Join(root, "bin")
	for _, dir := range []string{libexec, bin} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	binary := filepath.Join(libexec, "riido")
	if err := os.WriteFile(binary, []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(bin, "riido-staging")
	if err := os.Symlink(binary, link); err != nil {
		t.Fatal(err)
	}
	if got := InstalledChannel(link); got != "" {
		t.Fatalf("no marker = %q", got)
	}
	marker := filepath.Join(libexec, ChannelMarkerName)
	if err := os.WriteFile(marker, []byte("staging\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := InstalledChannel(link); got != ChannelStaging {
		t.Fatalf("marker via link = %q", got)
	}
	for _, body := range []string{"production\n", "staging-extra\n"} {
		if err := os.WriteFile(marker, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		if got := InstalledChannel(binary); got != "" {
			t.Fatalf("marker %q = %q", body, got)
		}
	}
	if err := os.Remove(marker); err != nil {
		t.Fatal(err)
	}
	other := filepath.Join(root, "marker")
	if err := os.WriteFile(other, []byte("staging\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(other, marker); err != nil {
		t.Fatal(err)
	}
	if got := InstalledChannelAt(libexec); got != "" {
		t.Fatalf("linked marker accepted: %q", got)
	}
	if ConfigName(ChannelStaging) != "riido-cli-staging" || ConfigName("") != "riido-cli" || CommandName(ChannelStaging) != "riido-staging" || CommandName("") != "riido" {
		t.Fatal("channel names changed")
	}
}
