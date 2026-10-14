// Package execution owns explicit host-authorized repository bindings and checks.
package execution

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func Load(path string) ([]Binding, error) {
	if path == "" {
		return nil, nil
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var bindings []Binding
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&bindings); err != nil {
		return nil, err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return nil, errors.New("execution bindings: trailing JSON")
	}
	seen := map[string]bool{}
	for _, binding := range bindings {
		key := binding.WorkspaceID + "\x00" + binding.TaskID
		if binding.WorkspaceID == "" || binding.TaskID == "" || !filepath.IsAbs(binding.Repository) {
			return nil, errors.New("execution binding requires workspace_id, task_id and absolute repository")
		}
		if seen[key] {
			return nil, errors.New("execution binding repeats workspace/task")
		}
		seen[key] = true
		if len(binding.ChangedFiles) == 0 || len(binding.Checks) == 0 {
			return nil, errors.New("execution binding requires changed_files and checks")
		}
		if binding.TimeoutSeconds < 0 || binding.TimeoutSeconds > 600 {
			return nil, errors.New("execution binding timeout must be 0..600 seconds")
		}
		if binding.ContinuationLimit < 0 || binding.ContinuationLimit > 2 {
			return nil, errors.New("execution binding continuation_limit must be 0..2")
		}
		for _, path := range binding.ChangedFiles {
			if !filepath.IsLocal(path) || path == "." {
				return nil, fmt.Errorf("invalid changed file %q", path)
			}
		}
		for _, argv := range binding.Checks {
			if len(argv) == 0 || strings.TrimSpace(argv[0]) == "" {
				return nil, errors.New("execution check requires an executable")
			}
		}
	}
	return bindings, nil
}

func Prepare(ctx context.Context, binding Binding, workdir string) (*Prepared, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	dir := filepath.Join(workdir, "repository")
	cmd := exec.CommandContext(ctx, "git", "-c", "core.hooksPath=/dev/null", "clone", "--no-local", "--", binding.Repository, dir)
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("repository preparation failed: %w: %s", err, output)
	}
	cmd = exec.CommandContext(ctx, "git", "-C", dir, "remote", "remove", "origin")
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("remove clone remote: %w: %s", err, output)
	}
	prepared := &Prepared{Binding: binding, Directory: dir, before: map[string][32]byte{}, existed: map[string]bool{}}
	for _, path := range binding.ChangedFiles {
		body, err := readFile(dir, path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		prepared.before[path] = sha256.Sum256(body)
		prepared.existed[path] = true
	}
	return prepared, nil
}

func (p *Prepared) Verify(ctx context.Context) Evidence {
	evidence := Evidence{Directory: p.Directory}
	for _, path := range p.Binding.ChangedFiles {
		body, err := readFile(p.Directory, path)
		if err != nil {
			evidence.Error = fmt.Sprintf("required output %s: %v", path, err)
			evidence.ErrorKind = "invalid_output"
			if errors.Is(err, os.ErrNotExist) {
				evidence.ErrorKind = "missing_output"
			}
			return evidence
		}
		if p.existed[path] && p.before[path] == sha256.Sum256(body) {
			evidence.Error = "required output was not changed: " + path
			evidence.ErrorKind = "unchanged_output"
			return evidence
		}
		evidence.ChangedFiles = append(evidence.ChangedFiles, path)
	}
	timeout := time.Duration(p.Binding.TimeoutSeconds) * time.Second
	if timeout == 0 {
		timeout = 60 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	for _, argv := range p.Binding.Checks {
		cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
		cmd.Dir = p.Directory
		cmd.WaitDelay = 2 * time.Second
		output := &boundedOutput{}
		cmd.Stdout = output
		cmd.Stderr = output
		err := cmd.Run()
		evidence.Checks = append(evidence.Checks, CheckResult{Command: argv, Output: string(output.body), Passed: err == nil})
		if err != nil {
			evidence.Error = fmt.Sprintf("verification %s failed: %v: %s", strings.Join(argv, " "), err, output.body)
			evidence.ErrorKind = "check_failed"
			if ctx.Err() != nil {
				evidence.ErrorKind = "verification_timeout"
			}
			return evidence
		}
	}
	return evidence
}

func readFile(root, path string) ([]byte, error) {
	resolved, err := filepath.EvalSymlinks(filepath.Join(root, path))
	if err != nil {
		return nil, err
	}
	canonicalRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return nil, err
	}
	relative, err := filepath.Rel(canonicalRoot, resolved)
	if err != nil || !filepath.IsLocal(relative) {
		return nil, errors.New("required output escapes repository")
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > 8*1024*1024 {
		return nil, errors.New("required output must be a regular file under 8 MiB")
	}
	return os.ReadFile(resolved)
}

type boundedOutput struct{ body []byte }

func (b *boundedOutput) Write(p []byte) (int, error) {
	n := len(p)
	remaining := 8192 - len(b.body)
	if remaining > 0 {
		if len(p) > remaining {
			p = p[:remaining]
		}
		b.body = append(b.body, p...)
	}
	return n, nil
}
