package main

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/cliinstallation"
	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/securecredential"
)

func (c contextIO) doctorInstallation(env contextcli.Environment) []doctorCheck {
	exe, err := c.executable()
	if err != nil {
		return []doctorCheck{{Name: "binary_target", Status: doctorWarn, Detail: "실행 파일의 실제 경로를 확인하지 못했습니다"}}
	}
	resolved, err := filepath.EvalSymlinks(exe)
	if err != nil {
		return []doctorCheck{{Name: "binary_target", Status: doctorWarn, Detail: "실행 파일의 symlink 대상이 없거나 접근할 수 없습니다"}}
	}
	checks := []doctorCheck{{Name: "binary_target", Status: doctorOK, Detail: resolved}}
	if env.Mode == contextcli.ModeAgent {
		return append(checks, doctorCheck{Name: "helper_bundle", Status: doctorSkip, Detail: "에이전트 모드에서는 사용자 인증 도우미를 사용하지 않습니다"})
	}
	root := filepath.Dir(resolved)
	rawVersion, versionErr := doctorReadRegular(filepath.Join(root, "VERSION"), 4096)
	version := strings.TrimSpace(string(rawVersion))
	versionOK := versionErr == nil && cliinstallation.ValidateVersion(version) == nil && c.cliVersion() == "riido-daemon "+version
	if versionOK {
		checks = append(checks, doctorCheck{Name: "package_version", Status: doctorOK, Detail: version})
	} else {
		checks = append(checks, doctorCheck{Name: "package_version", Status: doctorWarn, Detail: "인접 VERSION이 없거나 실행 중인 CLI 버전과 일치하지 않습니다"})
	}
	manifestPath := filepath.Join(root, cliinstallation.ManifestName)
	bundle := filepath.Join(root, cliinstallation.BundleName)
	_, manifestErr := os.Lstat(manifestPath)
	bundleInfo, bundleErr := os.Lstat(bundle)
	if errors.Is(manifestErr, os.ErrNotExist) && errors.Is(bundleErr, os.ErrNotExist) {
		return append(checks, doctorCheck{Name: "helper_bundle", Status: doctorWarn, Detail: "인증 도우미가 없는 legacy 설치입니다. 설치 확인은 로그인 성공을 뜻하지 않습니다"})
	}
	fail := func(detail string) []doctorCheck {
		return append(checks, doctorCheck{Name: "helper_bundle", Status: doctorFail, Detail: detail})
	}
	if manifestErr != nil || bundleErr != nil || !bundleInfo.IsDir() || bundleInfo.Mode()&os.ModeSymlink != 0 {
		return fail("helper와 manifest가 모두 있는 정상 설치가 아닙니다")
	}
	if !versionOK {
		return fail("helper를 포함한 설치의 VERSION과 실행 중인 CLI 버전이 일치하지 않습니다")
	}
	manifest, err := doctorReadRegular(manifestPath, 8192)
	if err != nil || cliinstallation.CheckManifest(manifest, version) != nil {
		return fail("helper manifest 형식·경로·버전이 올바르지 않습니다")
	}
	info, err := doctorReadRegular(filepath.Join(bundle, "Contents/Info.plist"), 1<<20)
	if err != nil || cliinstallation.CheckInfo(info, version) != nil {
		return fail("helper Info.plist의 식별자·실행 파일·버전이 올바르지 않습니다")
	}
	profile, err := doctorReadRegular(filepath.Join(bundle, "Contents/embedded.provisionprofile"), 4<<20)
	if err != nil || len(profile) == 0 {
		return fail("helper provisioning profile이 없거나 올바른 파일이 아닙니다")
	}
	helperPath := filepath.Join(bundle, "Contents/MacOS", cliinstallation.ExecutableName)
	helperInfo, err := os.Lstat(helperPath)
	if err != nil || helperInfo.Mode().Perm()&0o111 == 0 {
		return fail("helper 실행 파일이 없거나 실행 권한이 없습니다")
	}
	executable, err := doctorReadRegular(helperPath, cliinstallation.MaxBundleBytes)
	if err != nil {
		return fail("helper 실행 파일을 확인하지 못했습니다")
	}
	archs, err := cliinstallation.CheckArchitecture(executable, runtime.GOARCH)
	if err != nil {
		return fail("helper가 현재 CLI 아키텍처와 맞지 않습니다")
	}
	checks = append(checks, doctorCheck{Name: "helper_bundle", Status: doctorOK, Detail: bundle + " (" + version + ")"})
	verify := c.verifyHelperSignature
	if verify == nil {
		verify = securecredential.VerifyPackagedHelper
	}
	ctx, cancel := context.WithTimeout(context.Background(), doctorProbeTimeout)
	defer cancel()
	if err = verify(ctx, bundle, archs); err != nil {
		status := doctorFail
		detail := "helper 정적 서명 검증에 실패했습니다. 자격에는 접근하지 않았습니다"
		if runtime.GOOS != "darwin" {
			status = doctorUnsupported
			detail = "현재 OS에서는 macOS helper 서명을 검증할 수 없습니다"
		}
		return append(checks, doctorCheck{Name: "helper_signature", Status: status, Detail: detail})
	}
	return append(checks, doctorCheck{Name: "helper_signature", Status: doctorOK, Detail: "정적 서명만 확인했습니다. 로그인·사용자 확인·API 성공은 확인하지 않았습니다"})
}

func doctorReadRegular(name string, limit int64) ([]byte, error) {
	info, err := os.Lstat(name)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > limit {
		return nil, errors.New("invalid diagnostic file")
	}
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	raw, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil || int64(len(raw)) > limit {
		return nil, errors.New("diagnostic file exceeds size limit")
	}
	return raw, nil
}
