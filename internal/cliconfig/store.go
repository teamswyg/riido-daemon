package cliconfig

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/teamswyg/riido-daemon/internal/lock"
	"github.com/teamswyg/riido-daemon/pkg/util/fileutil"
)

const maxConfigBytes = 64 << 10

type Store struct{ Dir string }

func (s Store) Load() (Config, error) {
	if s.Dir == "" {
		return Default(), nil
	}
	if err := s.checkDirectory(); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Default(), nil
		}
		return Config{}, ErrUnavailable
	}
	root, err := os.OpenRoot(s.Dir)
	if err != nil {
		return Config{}, ErrUnavailable
	}
	defer root.Close()
	info, err := root.Lstat("config.json")
	if errors.Is(err, os.ErrNotExist) {
		return Default(), nil
	}
	if err != nil || !info.Mode().IsRegular() {
		return Config{}, ErrInvalid
	}
	f, err := root.Open("config.json")
	if err != nil {
		return Config{}, ErrUnavailable
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, maxConfigBytes+1))
	if err != nil {
		return Config{}, ErrUnavailable
	}
	if len(raw) > maxConfigBytes {
		return Config{}, ErrInvalid
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var c Config
	if decoder.Decode(&c) != nil || decoder.Decode(new(any)) != io.EOF || !c.Valid() {
		return Config{}, ErrInvalid
	}
	return c, nil
}

// Update serializes read/modify/replace. A malformed file is never reset.
func (s Store) Update(ctx context.Context, change func(*Config) error) (result error) {
	if s.Dir == "" || change == nil {
		return ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := os.MkdirAll(s.Dir, 0o700); err != nil {
		return ErrUnavailable
	}
	if err := s.checkDirectory(); err != nil {
		return ErrUnavailable
	}
	if info, err := os.Lstat(filepath.Join(s.Dir, "config.lock")); err == nil && !info.Mode().IsRegular() || err != nil && !errors.Is(err, os.ErrNotExist) {
		return ErrUnavailable
	}
	guard, err := lock.AcquireFile(ctx, filepath.Join(s.Dir, "config.lock"))
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return ErrUnavailable
	}
	defer func() {
		if err := guard.Release(); err != nil && result == nil {
			result = ErrUnavailable
		}
	}()
	c, err := s.Load()
	if err != nil {
		return err
	}
	if err = change(&c); err != nil {
		return err
	}
	if !c.Valid() {
		return ErrInvalid
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	if err = fileutil.WriteJSONAtomic(filepath.Join(s.Dir, "config.json"), c); err != nil {
		return ErrUnavailable
	}
	return nil
}

func (s Store) checkDirectory() error {
	info, err := os.Lstat(s.Dir)
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return ErrUnavailable
	}
	return nil
}
