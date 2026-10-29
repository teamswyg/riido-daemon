package childenv

import (
	"slices"
	"testing"
)

func TestScrubRemovesDaemonSecretsCaseInsensitively(t *testing.T) {
	got := Scrub([]string{"PATH=/bin", "RIIDO_DEVICE_SECRET=a", "riido_device_secret=b", "Riido_Token=c", "RIIDO_SAAS_TOKEN=d", "RIIDO_DEVICE_ID=e"})
	want := []string{"PATH=/bin", "RIIDO_DEVICE_ID=e"}
	if !slices.Equal(got, want) {
		t.Fatalf("Scrub = %v, want %v", got, want)
	}
}

func TestEnvironNeverContainsDeviceSecret(t *testing.T) {
	t.Setenv("RIIDO_DEVICE_SECRET", "rdev-secret")
	t.Setenv("RIIDO_DEVICE_ID", "device-1")
	env := Environ()
	if slices.Contains(env, "RIIDO_DEVICE_SECRET=rdev-secret") {
		t.Fatal("Environ leaked RIIDO_DEVICE_SECRET")
	}
	if !slices.Contains(env, "RIIDO_DEVICE_ID=device-1") {
		t.Fatal("Environ dropped a non-secret variable")
	}
}
