package detectutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMacOSAppResourceDirsCoverChatGPTAndLegacyCodex(t *testing.T) {
	home := t.TempDir()
	dirs := macOSAppResourceDirs(home)
	wants := []string{
		"/Applications/ChatGPT.app/Contents/Resources",
		"/Applications/Codex.app/Contents/Resources",
		filepath.Join(home, "Applications", "ChatGPT.app", "Contents", "Resources"),
		filepath.Join(home, "Applications", "Codex.app", "Contents", "Resources"),
	}
	for _, want := range wants {
		if !containsDir(dirs, want) {
			t.Fatalf("macOS app resource dir %q missing from %v", want, dirs)
		}
	}
}

func TestResolveExecutableFindsChatGPTBundledCodex(t *testing.T) {
	resourceDir := filepath.Join(t.TempDir(), "ChatGPT.app", "Contents", "Resources")
	if err := os.MkdirAll(resourceDir, 0o755); err != nil {
		t.Fatal(err)
	}
	want := writeExecutable(t, filepath.Join(resourceDir, "codex"), "codex-cli test")
	t.Setenv("PATH", t.TempDir())
	overrideAugmentedSearchDirs(t, resourceDir)
	got, ok := ResolveExecutable("codex", "")
	if !ok || got != want {
		t.Fatalf("bundled codex = %q ok=%v, want %q", got, ok, want)
	}
}
