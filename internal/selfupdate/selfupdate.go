// Package selfupdate replaces an installer-managed CLI with the latest package
// of its release channel (RIID-6021). It uses the same CDN assets and checks as
// scripts/install-riido-daemon.sh and leaves the current install untouched on
// any download, checksum, archive or helper failure.
package selfupdate

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"github.com/teamswyg/riido-daemon/internal/cliinstallation"
)

// DefaultBaseURL is the release CDN root. Channels map to the same active feeds
// as the desktop app and scripts/install-riido-daemon.sh: latest -> safe,
// staging -> staging-safe. The legacy latest/staging prefixes are frozen.
const DefaultBaseURL = "https://cdn.riido.io/releases"

const (
	ChannelStaging = "staging"
	ChannelLatest  = "latest"
	maxSumsBytes   = 64 << 10
)

var (
	ErrUnsupported = errors.New("self update is not supported for this installation")
	ErrInvalid     = errors.New("release package failed verification")
)

// cdnPrefix is the active CDN feed of a public channel name (RIID-6036).
var cdnPrefix = map[string]string{ChannelLatest: "safe", ChannelStaging: "staging-safe"}

// IsReleaseVersion reports whether version has the vMAJOR.MINOR.PATCH core
// that riido update compares. Other published names are never auto-installed.
func IsReleaseVersion(version string) bool {
	_, ok := versionCore(version)
	return ok
}

// IsOlder reports whether candidate's vMAJOR.MINOR.PATCH core is lower than
// current's. Versions without that core never count as older.
func IsOlder(candidate, current string) bool {
	a, okA := versionCore(candidate)
	b, okB := versionCore(current)
	if !okA || !okB {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return a[i] < b[i]
		}
	}
	return false
}

func versionCore(version string) ([3]int, bool) {
	var core [3]int
	version, ok := strings.CutPrefix(version, "v")
	if !ok {
		return core, false
	}
	if i := strings.IndexAny(version, "-+"); i >= 0 {
		version = version[:i]
	}
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return core, false
	}
	for i, part := range parts {
		n, err := strconv.Atoi(part)
		if err != nil || n < 0 || part == "" || len(part) > 9 {
			return core, false
		}
		core[i] = n
	}
	return core, true
}

func ValidChannel(channel string) bool {
	return channel == ChannelStaging || channel == ChannelLatest
}

// Updater downloads one channel package. Verify checks the staged helper bundle
// before it can replace the installed one (securecredential.VerifyPackagedHelper
// in the CLI); tests inject a fake.
type Updater struct {
	BaseURL string
	Client  *http.Client
	GOOS    string
	GOARCH  string
	Verify  func(ctx context.Context, bundle string, architectures []string) error
}

// Staged is a verified package extracted next to the install root.
type Staged struct {
	Dir       string
	Version   string
	HasHelper bool
}

func (s *Staged) Cleanup() {
	if s != nil && s.Dir != "" {
		_ = os.RemoveAll(s.Dir)
	}
}

func (u Updater) assetName() (string, error) {
	if u.GOOS != "darwin" || (u.GOARCH != "arm64" && u.GOARCH != "amd64") {
		return "", ErrUnsupported
	}
	return "riido-daemon_" + u.GOOS + "_" + u.GOARCH + ".tar.gz", nil
}

func (u Updater) channelURL(channel, name string) (string, error) {
	base, err := url.Parse(u.BaseURL)
	if err != nil || base.Scheme != "https" || base.Host == "" || base.User != nil || base.RawQuery != "" || base.Fragment != "" || !ValidChannel(channel) {
		return "", ErrInvalid
	}
	return strings.TrimSuffix(base.String(), "/") + "/" + cdnPrefix[channel] + "/ai-agent/" + name, nil
}

// Stage downloads, checks and extracts the channel package into a new
// directory under root (same filesystem, so Install can rename atomically).
func (u Updater) Stage(ctx context.Context, channel, root string) (*Staged, error) {
	asset, err := u.assetName()
	if err != nil {
		return nil, err
	}
	dir, err := os.MkdirTemp(root, ".riido-update-")
	if err != nil {
		return nil, err
	}
	staged := &Staged{Dir: dir}
	fail := func(err error) (*Staged, error) {
		staged.Cleanup()
		return nil, err
	}
	sums, err := u.fetch(ctx, channel, "SHA256SUMS", maxSumsBytes)
	if err != nil {
		return fail(err)
	}
	want, err := checksumFor(sums, asset)
	if err != nil {
		return fail(err)
	}
	archive, err := u.fetch(ctx, channel, asset, cliinstallation.MaxArchiveBytes)
	if err != nil {
		return fail(err)
	}
	sum := sha256.Sum256(archive)
	if hex.EncodeToString(sum[:]) != want {
		return fail(fmt.Errorf("%w: checksum mismatch", ErrInvalid))
	}
	pkg := filepath.Join(dir, "package")
	if err = extract(archive, pkg); err != nil {
		return fail(err)
	}
	if err = u.check(ctx, staged, pkg); err != nil {
		return fail(err)
	}
	return staged, nil
}

func (u Updater) fetch(ctx context.Context, channel, name string, limit int64) ([]byte, error) {
	target, err := u.channelURL(channel, name)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}
	client := u.Client
	if client == nil {
		client = http.DefaultClient
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK || res.Request == nil || res.Request.URL.Scheme != "https" {
		return nil, fmt.Errorf("%w: download %s failed (%d)", ErrInvalid, name, res.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > limit {
		return nil, fmt.Errorf("%w: %s too large", ErrInvalid, name)
	}
	return body, nil
}

func checksumFor(sums []byte, asset string) (string, error) {
	scanner := bufio.NewScanner(bytes.NewReader(sums))
	found := ""
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 2 && fields[1] == asset {
			if found != "" {
				return "", fmt.Errorf("%w: duplicate checksum", ErrInvalid)
			}
			found = strings.ToLower(fields[0])
		}
	}
	if len(found) != 64 {
		return "", fmt.Errorf("%w: checksum for %s is missing", ErrInvalid, asset)
	}
	if _, err := hex.DecodeString(found); err != nil {
		return "", fmt.Errorf("%w: invalid checksum", ErrInvalid)
	}
	return found, nil
}

// allowedTop mirrors the release package layout checked by the CDN publisher.
var allowedTop = map[string]bool{"riido": true, "VERSION": true, "LICENSE": true, "NOTICE.md": true, cliinstallation.ManifestName: true, cliinstallation.BundleName: true}

func extract(archive []byte, dest string) error {
	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalid, err)
	}
	defer gz.Close()
	limited := &io.LimitedReader{R: gz, N: cliinstallation.MaxArchiveBytes + 1}
	reader := tar.NewReader(limited)
	if err = os.MkdirAll(dest, 0o700); err != nil {
		return err
	}
	seen := map[string]bool{}
	for count := 0; ; count++ {
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("%w: %w", ErrInvalid, err)
		}
		if count > 8192 {
			return fmt.Errorf("%w: too many files", ErrInvalid)
		}
		name, err := safeName(header.Name)
		if err != nil {
			return err
		}
		if name == "." {
			continue
		}
		top := strings.SplitN(name, "/", 2)[0]
		folded := strings.ToLower(name)
		if !allowedTop[top] || seen[folded] || (top != cliinstallation.BundleName && strings.Contains(name, "/")) {
			return fmt.Errorf("%w: unexpected archive content", ErrInvalid)
		}
		seen[folded] = true
		if header.Mode&0o6000 != 0 || header.Size < 0 || header.Size > cliinstallation.MaxBundleBytes {
			return fmt.Errorf("%w: unsafe archive entry", ErrInvalid)
		}
		target := filepath.Join(dest, filepath.FromSlash(name))
		switch header.Typeflag {
		case tar.TypeDir:
			if err = os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err = os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			mode := os.FileMode(0o644)
			if header.Mode&0o111 != 0 {
				mode = 0o755
			}
			file, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
			if err != nil {
				return err
			}
			_, copyErr := io.Copy(file, io.LimitReader(reader, header.Size))
			closeErr := file.Close()
			if copyErr != nil || closeErr != nil {
				return errors.Join(copyErr, closeErr)
			}
		default:
			return fmt.Errorf("%w: archive contains a link or special file", ErrInvalid)
		}
	}
	if limited.N <= 0 {
		return fmt.Errorf("%w: archive exceeds size limit", ErrInvalid)
	}
	return nil
}

func safeName(name string) (string, error) {
	name = strings.TrimSuffix(strings.TrimPrefix(name, "./"), "/")
	if name == "" || name == "." {
		return ".", nil
	}
	if strings.HasPrefix(name, "/") || strings.ContainsAny(name, "\\:") || path.Clean(name) != name || name == ".." || strings.HasPrefix(name, "../") {
		return "", fmt.Errorf("%w: unsafe archive path", ErrInvalid)
	}
	for _, r := range name {
		if unicode.IsControl(r) {
			return "", fmt.Errorf("%w: unsafe archive path", ErrInvalid)
		}
	}
	return name, nil
}

func (u Updater) check(ctx context.Context, staged *Staged, pkg string) error {
	invalid := func(why string) error { return fmt.Errorf("%w: %s", ErrInvalid, why) }
	rawVersion, err := os.ReadFile(filepath.Join(pkg, "VERSION"))
	if err != nil {
		return invalid("package is missing VERSION")
	}
	version := strings.TrimSpace(strings.SplitN(string(rawVersion), "\n", 2)[0])
	if cliinstallation.ValidateVersion(version) != nil || !IsReleaseVersion(version) {
		return invalid("invalid VERSION")
	}
	staged.Version = version
	binary, err := os.ReadFile(filepath.Join(pkg, "riido"))
	if err != nil {
		return invalid("package is missing riido")
	}
	if _, err = cliinstallation.CheckArchitecture(binary, u.GOARCH); err != nil {
		return invalid("riido does not match this Mac")
	}
	manifestPath := filepath.Join(pkg, cliinstallation.ManifestName)
	bundle := filepath.Join(pkg, cliinstallation.BundleName)
	_, manifestErr := os.Stat(manifestPath)
	_, bundleErr := os.Stat(bundle)
	if errors.Is(manifestErr, os.ErrNotExist) && errors.Is(bundleErr, os.ErrNotExist) {
		return nil // a helperless package (non-signed build) installs without a helper
	}
	if manifestErr != nil || bundleErr != nil {
		return invalid("incomplete helper")
	}
	manifest, err := os.ReadFile(manifestPath)
	if err != nil || cliinstallation.CheckManifest(manifest, version) != nil {
		return invalid("helper manifest")
	}
	info, err := os.ReadFile(filepath.Join(bundle, "Contents", "Info.plist"))
	if err != nil || cliinstallation.CheckInfo(info, version) != nil {
		return invalid("helper Info.plist")
	}
	if profile, err := os.ReadFile(filepath.Join(bundle, "Contents", "embedded.provisionprofile")); err != nil || len(profile) == 0 {
		return invalid("helper provisioning profile")
	}
	helper, err := os.ReadFile(filepath.Join(bundle, "Contents", "MacOS", cliinstallation.ExecutableName))
	if err != nil {
		return invalid("helper executable")
	}
	archs, err := cliinstallation.CheckArchitecture(helper, u.GOARCH)
	if err != nil {
		return invalid("helper does not match this Mac")
	}
	if u.Verify == nil {
		return invalid("helper signature cannot be verified")
	}
	abs, err := filepath.Abs(bundle)
	if err != nil {
		return err
	}
	if err = u.Verify(ctx, abs, archs); err != nil {
		return invalid("helper signature")
	}
	staged.HasHelper = true
	return nil
}

// Install swaps the verified package into root: helper and metadata first,
// the riido binary last, each by rename within the same directory, so a failed
// step leaves a usable binary. The running process keeps its open image.
func Install(staged *Staged, root string) error {
	if staged == nil || staged.Version == "" {
		return ErrInvalid
	}
	pkg := filepath.Join(staged.Dir, "package")
	bundle := filepath.Join(root, cliinstallation.BundleName)
	manifest := filepath.Join(root, cliinstallation.ManifestName)
	if staged.HasHelper {
		old := bundle + ".old"
		_ = os.RemoveAll(old)
		if _, err := os.Lstat(bundle); err == nil {
			if err = os.Rename(bundle, old); err != nil {
				return err
			}
		}
		if err := os.Rename(filepath.Join(pkg, cliinstallation.BundleName), bundle); err != nil {
			_ = os.Rename(old, bundle)
			return err
		}
		_ = os.RemoveAll(old)
		if err := os.Rename(filepath.Join(pkg, cliinstallation.ManifestName), manifest); err != nil {
			return err
		}
	} else {
		// A package without the helper must not leave an older helper beside
		// the new binary; the CLI rejects a helper from another version.
		_ = os.RemoveAll(bundle)
		_ = os.Remove(manifest)
	}
	if err := os.Rename(filepath.Join(pkg, "VERSION"), filepath.Join(root, "VERSION")); err != nil {
		return err
	}
	return os.Rename(filepath.Join(pkg, "riido"), filepath.Join(root, "riido"))
}
