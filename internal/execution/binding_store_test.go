package execution

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func fixtureBinding(task string) Binding {
	return Binding{WorkspaceID: "workspace", TaskID: task, Repository: filepath.Join(os.TempDir(), "repository"), ChangedFiles: []string{"result.txt"}, Checks: [][]string{{"git", "diff", "--check"}}}
}

func TestBindingStorePersistsScopedUpdatesAndDetachesRunningSnapshot(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bindings.json")
	store, err := OpenBindingStore(path)
	if err != nil {
		t.Fatal(err)
	}
	binding := fixtureBinding("task")
	if err := store.Replace("workspace", "task", &binding); err != nil {
		t.Fatal(err)
	}
	prior := store.Snapshot()
	binding.Checks[0][0] = "changed-by-caller"
	if store.Snapshot()[0].Checks[0][0] != "git" {
		t.Fatal("caller modified stored verification authority")
	}
	next := fixtureBinding("task")
	next.Repository = filepath.Join(os.TempDir(), "new-repository")
	if err := store.Replace("workspace", "task", &next); err != nil {
		t.Fatal(err)
	}
	if prior[0].Repository == store.Snapshot()[0].Repository {
		t.Fatal("running snapshot changed with future configuration")
	}
	reopened, err := OpenBindingStore(path)
	if err != nil || reopened.Snapshot()[0].Repository != next.Repository {
		t.Fatal("configuration was not durably replaced")
	}
	info, _ := os.Stat(path)
	if os.PathSeparator != '\\' && info.Mode().Perm() != 0o600 {
		t.Fatal("binding file must be owner-only")
	}
	if err := store.Replace("other-workspace", "task", &next); err == nil {
		t.Fatal("cross-workspace replacement accepted")
	}
	invalid := next
	invalid.ChangedFiles = []string{"../escape"}
	if err := store.Replace("workspace", "task", &invalid); err == nil || store.Snapshot()[0].ChangedFiles[0] != "result.txt" {
		t.Fatal("invalid replacement changed accepted authority")
	}
}

func TestBindingStoreConcurrentTasksDoNotLoseUpdates(t *testing.T) {
	store, err := OpenBindingStore(filepath.Join(t.TempDir(), "bindings.json"))
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := range 12 {
		wg.Go(func() {
			binding := fixtureBinding(fmt.Sprint(i))
			if err := store.Replace(binding.WorkspaceID, binding.TaskID, &binding); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	if len(store.Snapshot()) != 12 {
		t.Fatal("concurrent task configuration lost updates")
	}
}

func TestBindingStoreRejectsSymlinkTarget(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bindings.json")
	target := filepath.Join(t.TempDir(), "target.json")
	if err := os.WriteFile(target, []byte("[]"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, path); err != nil {
		t.Skip("symlink unavailable on this host")
	}
	if _, err := OpenBindingStore(path); err == nil {
		t.Fatal("symlink binding target accepted")
	}
}

func TestRepositoryOnlyBindingDoesNotClaimVerification(t *testing.T) {
	binding := Binding{WorkspaceID: "w", TaskID: "t", Repository: t.TempDir()}
	if err := ValidateBindings([]Binding{binding}); err != nil {
		t.Fatal(err)
	}
	prepared := &Prepared{Binding: binding}
	if evidence := prepared.Verify(context.Background()); evidence.ErrorKind != "not_configured" {
		t.Fatal("missing checks were accepted as verified")
	}
	binding.ChangedFiles = []string{"out"}
	if ValidateBindings([]Binding{binding}) == nil {
		t.Fatal("half-configured verification accepted")
	}
}
