package execution

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

// BindingStore serializes host configuration updates. Readers get detached
// copies so an in-flight run never observes a later UI configuration change.
type BindingStore struct {
	mu       sync.RWMutex
	path     string
	bindings []Binding
}

func OpenBindingStore(path string) (*BindingStore, error) {
	if !filepath.IsAbs(path) {
		return nil, errors.New("execution binding store requires an absolute path")
	}
	if info, err := os.Lstat(path); err == nil && !info.Mode().IsRegular() {
		return nil, errors.New("execution binding target must be a regular file")
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	bindings, err := Load(path)
	if errors.Is(err, os.ErrNotExist) {
		err = nil
	}
	if err != nil {
		return nil, err
	}
	return &BindingStore{path: path, bindings: bindings}, nil
}

func copyBindings(in []Binding) []Binding {
	out := make([]Binding, len(in))
	for i, b := range in {
		out[i] = b
		out[i].ChangedFiles = append([]string(nil), b.ChangedFiles...)
		out[i].Checks = make([][]string, len(b.Checks))
		for j, argv := range b.Checks {
			out[i].Checks[j] = append([]string(nil), argv...)
		}
	}
	return out
}

func (s *BindingStore) Snapshot() []Binding {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return copyBindings(s.bindings)
}

// Replace updates one exact workspace/task binding without losing concurrent
// changes to other tasks. Removing a missing binding is idempotent.
func (s *BindingStore) Replace(workspaceID, taskID string, binding *Binding) error {
	if workspaceID == "" || taskID == "" {
		return errors.New("workspace/task scope is required")
	}
	if binding != nil && (binding.WorkspaceID != workspaceID || binding.TaskID != taskID) {
		return errors.New("execution binding scope mismatch")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	next := []Binding{}
	for _, current := range s.bindings {
		if current.WorkspaceID != workspaceID || current.TaskID != taskID {
			next = append(next, current)
		}
	}
	if binding != nil {
		next = append(next, *binding)
	}
	next = copyBindings(next)
	if err := ValidateBindings(next); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}
	if info, err := os.Lstat(s.path); err == nil && (!info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0) {
		return errors.New("execution binding target must be a regular file")
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	body, err := json.MarshalIndent(next, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(s.path), ".execution-bindings-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(append(body, '\n')); err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	if err := os.Rename(f.Name(), s.path); err != nil {
		return err
	}
	s.bindings = next
	return nil
}
