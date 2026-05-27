package repoidentity

import "testing"

func TestIdentity(t *testing.T) {
	if Name != "riido-daemon" {
		t.Fatalf("Name = %q", Name)
	}
	if ModulePath != "github.com/teamswyg/riido-daemon" {
		t.Fatalf("ModulePath = %q", ModulePath)
	}
	if Boundary == "" {
		t.Fatal("Boundary is empty")
	}
}
