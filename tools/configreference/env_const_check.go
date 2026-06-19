package main

import (
	"fmt"
	"slices"
)

func checkEnvConstants(repo string, manifest Manifest) []CheckResult {
	constants, err := parseEnvConstants(repoPath(repo, manifest.DaemonEnvSource))
	if err != nil {
		return []CheckResult{{Name: "parse-daemon-env-consts", File: manifest.DaemonEnvSource, Detail: err.Error()}}
	}
	return append(checkManifestEnvVars(constants, manifest), checkCodeEnvConsts(constants, manifest)...)
}

func checkManifestEnvVars(constants map[string]string, manifest Manifest) []CheckResult {
	results := make([]CheckResult, 0, len(manifest.DaemonEnvVars))
	for _, envVar := range manifest.DaemonEnvVars {
		results = append(results, checkOneManifestEnv(constants, manifest, envVar))
	}
	return results
}

func checkOneManifestEnv(constants map[string]string, manifest Manifest, envVar EnvVar) CheckResult {
	result := CheckResult{Name: "manifest-env-" + envVar.Name, File: manifest.DaemonEnvSource, Pass: true}
	if !slices.Contains(mapValues(constants), envVar.Name) {
		result.Pass = false
		result.Detail = "manifest env not found in daemon env consts"
	}
	return result
}

func checkCodeEnvConsts(constants map[string]string, manifest Manifest) []CheckResult {
	manifestNames := manifestEnvNames(manifest)
	results := make([]CheckResult, 0, len(constants))
	for constName, envName := range constants {
		pass := slices.Contains(manifestNames, envName)
		results = append(results, CheckResult{
			Name: "code-env-" + envName,
			File: manifest.DaemonEnvSource,
			Pass: pass, Detail: missingDetail(pass, constName),
		})
	}
	return results
}

func missingDetail(pass bool, constName string) string {
	if pass {
		return ""
	}
	return fmt.Sprintf("code const %s is missing from manifest", constName)
}
