package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/cliinstallation"
)

func TestDoctorPathUsesInstalledCommand(t *testing.T) {
	for _, tc := range []struct {
		name, marker, command string
	}{
		{"production", "", "riido"},
		{"staging", "staging\n", "riido-staging"},
		{"unknown marker", "other\n", "riido"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			exe := filepath.Join(dir, "riido")
			if err := os.WriteFile(exe, nil, 0o700); err != nil {
				t.Fatal(err)
			}
			if tc.marker != "" {
				if err := os.WriteFile(filepath.Join(dir, cliinstallation.ChannelMarkerName), []byte(tc.marker), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			for _, outcome := range []string{"same", "missing", "shadowed"} {
				t.Run(outcome, func(t *testing.T) {
					c := contextIO{
						executable: func() (string, error) { return exe, nil },
						getenv:     func(string) string { return "staging" },
						lookPath: func(name string) (string, error) {
							if name != tc.command {
								t.Errorf("looked up %q; want installed command %q", name, tc.command)
							}
							switch outcome {
							case "missing":
								return "", exec.ErrNotFound
							case "shadowed":
								return filepath.Join(dir, "other", name), nil
							default:
								return exe, nil
							}
						},
					}
					check := c.doctorPath()
					if outcome == "same" {
						if check.Status != doctorOK {
							t.Fatalf("matching install must pass: %+v", check)
						}
					} else if check.Status != doctorWarn || !strings.Contains(check.Detail, tc.command) {
						t.Fatalf("missing/shadowed command must warn with correct name: %+v", check)
					}
				})
			}
		})
	}
}
