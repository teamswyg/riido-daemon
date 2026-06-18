package scheduling

import "testing"

func TestNormalizeRequiredSurfaces(t *testing.T) {
	got := NormalizeRequiredSurfaces([]RequiredSurface{" MCP ", "mcp", "system_prompt", "worktree", ""})
	want := []RequiredSurface{SurfaceMCP, SurfaceSystemPrompt, SurfaceWorktree}
	if len(got) != len(want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %+v, want %+v", got, want)
		}
	}
}
