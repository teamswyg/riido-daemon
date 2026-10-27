package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type signatureVerifier func(context.Context, string, []string) error

func prepareBundle(ctx context.Context, source, out, version, arch string, verify signatureVerifier) error {
	if err := validateVersion(version); err != nil {
		return err
	}
	if !filepath.IsAbs(source) || filepath.Clean(source) != source || filepath.Base(source) != bundleName {
		return errors.New("helper input must be an absolute RiidoCLIAuthHelper.app path")
	}
	destination := filepath.Join(out, bundleName)
	if _, err := os.Lstat(destination); !os.IsNotExist(err) {
		return errors.New("helper output already exists")
	}
	if _, err := os.Lstat(filepath.Join(out, manifestName)); !os.IsNotExist(err) {
		return errors.New("helper manifest already exists")
	}
	if err := copyBundle(source, destination); err != nil {
		return err
	}
	if err := validateSignedBundle(ctx, destination, version, arch, verify); err != nil {
		return err
	}
	data, err := json.Marshal(helperManifest{1, version, bundleName})
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(out, manifestName), append(data, '\n'), 0o644)
}

func copyBundle(source, destination string) error {
	info, err := os.Lstat(source)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return errors.New("helper bundle must be a real directory")
	}
	root, err := os.OpenRoot(source)
	if err != nil {
		return err
	}
	defer root.Close()
	var total int64
	entries := 0
	return filepath.WalkDir(source, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		entries++
		if entries > 4096 {
			return errors.New("helper contains too many entries")
		}
		rel, err := filepath.Rel(source, current)
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 || (!info.IsDir() && !info.Mode().IsRegular()) || info.Mode()&(os.ModeSetuid|os.ModeSetgid) != 0 {
			return errors.New("helper contains unsafe file type or permissions")
		}
		dest := filepath.Join(destination, rel)
		if info.IsDir() {
			return os.Mkdir(dest, 0o755)
		}
		total += info.Size()
		if info.Size() < 0 || total > maxBundleBytes {
			return errors.New("helper bundle exceeds size limit")
		}
		input, err := root.Open(rel)
		if err != nil {
			return err
		}
		defer input.Close()
		output, err := os.OpenFile(dest, os.O_CREATE|os.O_EXCL|os.O_WRONLY, info.Mode().Perm())
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(output, io.LimitReader(input, info.Size()+1))
		closeErr := output.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	})
}

func validateSignedBundle(ctx context.Context, bundle, version, arch string, verify signatureVerifier) error {
	plist, err := os.ReadFile(filepath.Join(bundle, "Contents/Info.plist"))
	if err != nil {
		return err
	}
	if err = parseInfo(plist, version); err != nil {
		return err
	}
	profile, err := os.Lstat(filepath.Join(bundle, "Contents/embedded.provisionprofile"))
	if err != nil || !profile.Mode().IsRegular() || profile.Size() == 0 {
		return errors.New("helper provisioning profile missing")
	}
	executable := filepath.Join(bundle, "Contents/MacOS", executableName)
	info, err := os.Lstat(executable)
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0o111 == 0 {
		return errors.New("helper executable missing or not executable")
	}
	raw, err := os.ReadFile(executable)
	if err != nil {
		return err
	}
	archs, err := checkArchitecture(raw, arch)
	if err != nil {
		return err
	}
	if verify == nil {
		return errors.New("helper signature verifier unavailable")
	}
	if err = verify(ctx, bundle, archs); err != nil {
		return fmt.Errorf("helper signature verification failed: %w", err)
	}
	return nil
}
