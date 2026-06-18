package main

import (
	"strings"

	"github.com/teamswyg/riido-daemon/internal/policy"
)

type daemonPolicyBundleEnv struct {
	version string
	path    string
	doc     policy.PolicyBundle
}

func loadDaemonPolicyBundleEnv(getenv func(string) string) (daemonPolicyBundleEnv, error) {
	version := strings.TrimSpace(getenv(envPolicyBundle))
	path := strings.TrimSpace(getenv(envPolicyBundlePath))
	doc := policy.DefaultLocalPolicyBundle()
	if path != "" {
		bundle, err := policy.LoadPolicyBundleFile(path)
		if err != nil {
			return daemonPolicyBundleEnv{}, daemonWrapf(ErrDaemonConfig, "settings.load-policy-bundle", err, "load policy bundle file")
		}
		if version != "" && version != bundle.Version {
			return daemonPolicyBundleEnv{}, daemonErrorf(ErrDaemonConfig, "settings.validate-policy-bundle", "%s=%q does not match %s version %q", envPolicyBundle, version, envPolicyBundlePath, bundle.Version)
		}
		doc = bundle
		version = bundle.Version
	}
	if version == "" {
		version = policy.DefaultLocalPolicyBundleVersion
	} else if path == "" {
		doc.Version = version
	}
	return daemonPolicyBundleEnv{version: version, path: path, doc: doc}, nil
}
