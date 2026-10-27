package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"unicode"
)

type packageEntry struct {
	data      []byte
	mode      int64
	directory bool
}

func walkArchive(archive string, consume func(string, *tar.Header, io.Reader) error) error {
	file, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer file.Close()
	gz, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gz.Close()
	limited := &io.LimitedReader{R: gz, N: maxArchiveBytes + 1}
	reader := tar.NewReader(limited)
	names := map[string]bool{}
	count := 0
	for {
		header, e := reader.Next()
		if errors.Is(e, io.EOF) {
			break
		}
		if e != nil {
			return e
		}
		count++
		if count > 8192 {
			return errors.New("archive contains too many files")
		}
		name, e := safeArchiveName(header.Name)
		if e != nil {
			return e
		}
		foldedName := strings.ToLower(name)
		if _, ok := names[foldedName]; ok {
			return errors.New("duplicate archive path")
		}
		names[foldedName] = header.Typeflag == tar.TypeDir
		if header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeDir {
			return errors.New("archive contains link or special file")
		}
		if header.Mode&0o6000 != 0 || header.Size < 0 || header.Size > maxArchiveBytes {
			return errors.New("archive has unsafe permissions or size")
		}
		if e = consume(name, header, reader); e != nil {
			return e
		}
	}
	// Tar end markers may be followed by zero padding, never another hidden
	// payload that an extractor could interpret differently.
	padding := make([]byte, 32<<10)
	for {
		n, readErr := limited.Read(padding)
		for _, b := range padding[:n] {
			if b != 0 {
				return errors.New("uninspected trailing archive payload")
			}
		}
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return readErr
		}
	}
	if limited.N <= 0 {
		return errors.New("archive exceeds size limit")
	}
	for name := range names {
		for parent := path.Dir(name); parent != "."; parent = path.Dir(parent) {
			if isDir, exists := names[parent]; exists && !isDir {
				return errors.New("archive has conflicting file and directory")
			}
		}
	}
	return nil
}

func safeArchiveName(name string) (string, error) {
	name = strings.TrimPrefix(name, "./")
	name = strings.TrimSuffix(name, "/")
	if name == "" || name == "." {
		return ".", nil
	}
	if strings.HasPrefix(name, "/") || strings.ContainsAny(name, "\\:") || path.Clean(name) != name || name == ".." || strings.HasPrefix(name, "../") {
		return "", errors.New("unsafe archive path")
	}
	for _, r := range name {
		if unicode.IsControl(r) {
			return "", errors.New("unsafe archive path")
		}
	}
	return name, nil
}

func verifyArchive(archive, version, arch string) error {
	if err := validateVersion(version); err != nil {
		return err
	}
	entries := map[string]packageEntry{}
	hasBundle := false
	required := map[string]bool{"VERSION": true, manifestName: true, "riido": true, bundleName + "/Contents/Info.plist": true, bundleName + "/Contents/MacOS/" + executableName: true, bundleName + "/Contents/embedded.provisionprofile": true}
	err := walkArchive(archive, func(name string, h *tar.Header, r io.Reader) error {
		top := strings.SplitN(name, "/", 2)[0]
		switch top {
		case ".", "riido", "LICENSE", "NOTICE.md", "VERSION", manifestName:
			if strings.Contains(name, "/") {
				return errors.New("unexpected nested archive file")
			}
		case bundleName:
			hasBundle = true
		default:
			return errors.New("unexpected archive content")
		}
		value := packageEntry{mode: h.Mode, directory: h.Typeflag == tar.TypeDir}
		if required[name] && !value.directory {
			if h.Size > maxBundleBytes {
				return errors.New("package metadata or executable too large")
			}
			body, e := io.ReadAll(io.LimitReader(r, maxBundleBytes+1))
			if e != nil {
				return e
			}
			value.data = body
		}
		entries[name] = value
		return nil
	})
	if err != nil {
		return err
	}
	versionEntry := entries["VERSION"]
	if versionEntry.directory || strings.TrimSpace(string(versionEntry.data)) != version {
		return errors.New("archive VERSION mismatch")
	}
	manifest, hasManifest := entries[manifestName]
	if hasBundle != hasManifest {
		return errors.New("helper bundle and manifest must both be present")
	}
	if !hasBundle {
		return nil
	}
	if manifest.directory {
		return errors.New("helper manifest must be a file")
	}
	if err = checkManifest(manifest.data, version); err != nil {
		return err
	}
	info := entries[bundleName+"/Contents/Info.plist"]
	if info.directory {
		return errors.New("helper Info.plist must be a file")
	}
	if err = parseInfo(info.data, version); err != nil {
		return err
	}
	profile := entries[bundleName+"/Contents/embedded.provisionprofile"]
	if profile.directory || len(profile.data) == 0 {
		return errors.New("helper provisioning profile missing")
	}
	for _, name := range []string{"riido", bundleName + "/Contents/MacOS/" + executableName} {
		entry := entries[name]
		if entry.directory || entry.mode&0o111 == 0 {
			return errors.New("package executable missing or not executable")
		}
		if _, err = checkArchitecture(entry.data, arch); err != nil {
			return err
		}
	}
	return nil
}

func recheckArchive(ctx context.Context, archive, version, arch string, verify signatureVerifier) error {
	out, err := os.MkdirTemp("", "riido-helper-roundtrip-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(out)
	err = walkArchive(archive, func(name string, h *tar.Header, r io.Reader) error {
		if name != bundleName && !strings.HasPrefix(name, bundleName+"/") {
			return nil
		}
		dest := filepath.Join(out, filepath.FromSlash(name))
		if h.Typeflag == tar.TypeDir {
			return os.MkdirAll(dest, 0o755)
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		file, err := os.OpenFile(dest, os.O_CREATE|os.O_EXCL|os.O_WRONLY, os.FileMode(h.Mode).Perm())
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(file, r)
		closeErr := file.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	})
	if err != nil {
		return err
	}
	return validateSignedBundle(ctx, filepath.Join(out, bundleName), version, arch, verify)
}
