package main

import "testing"

func TestRunAgainstRepositoryEvidence(t *testing.T) {
	err := run(options{Repo: "../..", Manifest: defaultManifest})
	if err != nil {
		t.Fatal(err)
	}
}

func TestManifestEnvNamesIncludeDaemonID(t *testing.T) {
	manifest := Manifest{DaemonEnvVars: []EnvVar{{Name: "RIIDO_DAEMON_ID"}}}
	names := manifestEnvNames(manifest)
	if len(names) != 1 || names[0] != "RIIDO_DAEMON_ID" {
		t.Fatalf("names = %v", names)
	}
}
