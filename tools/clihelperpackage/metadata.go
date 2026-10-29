package main

import "github.com/teamswyg/riido-daemon/internal/cliinstallation"

const (
	bundleName      = cliinstallation.BundleName
	executableName  = cliinstallation.ExecutableName
	manifestName    = cliinstallation.ManifestName
	maxBundleBytes  = cliinstallation.MaxBundleBytes
	maxArchiveBytes = cliinstallation.MaxArchiveBytes
)

type helperManifest = cliinstallation.Manifest

func parseInfo(raw []byte, version string) error { return cliinstallation.CheckInfo(raw, version) }
func checkManifest(raw []byte, version string) error {
	return cliinstallation.CheckManifest(raw, version)
}

func checkArchitecture(raw []byte, arch string) ([]string, error) {
	return cliinstallation.CheckArchitecture(raw, arch)
}
func validateVersion(version string) error { return cliinstallation.ValidateVersion(version) }
