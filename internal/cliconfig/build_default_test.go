package cliconfig

import "testing"

func TestBuildDefaultProfileSelectsChannelOnlyWithoutSavedChoice(t *testing.T) {
	original := buildDefaultProfile
	t.Cleanup(func() { buildDefaultProfile = original })
	none := func(string) string { return "" }
	for build, want := range map[string]string{"": "production", "staging": "staging", "local": "production", "bogus": "production", "production": "production"} {
		buildDefaultProfile = build
		resolved, err := Resolve(Default(), "", false, none)
		if err != nil || resolved.Profile != want || resolved.APIOrigin != DefaultAPIOrigin(want) {
			t.Fatalf("build %q: profile=%q origin=%q err=%v", build, resolved.Profile, resolved.APIOrigin, err)
		}
	}
	buildDefaultProfile = "staging"
	saved := Default()
	saved.ActiveProfile = "production"
	if resolved, _ := Resolve(saved, "", false, none); resolved.Profile != "production" {
		t.Fatal("saved profile overridden by the build default")
	}
	if resolved, _ := Resolve(Default(), "production", true, none); resolved.APIOrigin != "https://monolith.riido.io" {
		t.Fatal("--profile production overridden by the build default")
	}
	if resolved, _ := Resolve(Default(), "", false, func(k string) string {
		return map[string]string{"RIIDO_CLI_PROFILE": "production"}[k]
	}); resolved.Profile != "production" {
		t.Fatal("RIIDO_CLI_PROFILE overridden by the build default")
	}
}
