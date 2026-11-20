package cliconfig

// buildDefaultProfile is set at link time for channel builds, for example
// -X github.com/teamswyg/riido-daemon/internal/cliconfig.buildDefaultProfile=staging
// in the staging-only package. An unset or unknown value keeps production.
var buildDefaultProfile string

// installedDefaultProfile is "staging" when the running binary was installed
// as the staging test channel (RIID-6036): the same package then defaults to
// staging without a staging-only build.
var installedDefaultProfile string

// UseInstalledChannel records the install channel of the running binary
// (cliinstallation.InstalledChannel). Only "staging" changes the default.
func UseInstalledChannel(channel string) {
	if channel == "staging" {
		installedDefaultProfile = "staging"
	} else {
		installedDefaultProfile = ""
	}
}

// DefaultProfile is used when no saved configuration selects a profile. A
// saved activeProfile, RIIDO_CLI_PROFILE and --profile still take precedence.
func DefaultProfile() string {
	if buildDefaultProfile == "staging" || installedDefaultProfile == "staging" {
		return "staging"
	}
	return "production"
}

// DefaultAPIOrigin is the fixed API origin of a named profile ("" for local,
// which always needs an explicit api_origin).
func DefaultAPIOrigin(profile string) string {
	switch profile {
	case "production":
		return "https://monolith.riido.io"
	case "staging":
		return "https://monolith.api.riido.io"
	}
	return ""
}
