package main

import (
	"strings"

	"github.com/teamswyg/riido-daemon/pkg/util/textutil"
)

func loadDaemonSettingsFromEnvWithHome(getenv func(string) string, hostname, userHome func() (string, error)) (daemonSettings, error) {
	identity := loadDaemonIdentityEnv(getenv, hostname)
	workspace, err := loadDaemonWorkspaceEnv(getenv, userHome)
	if err != nil {
		return daemonSettings{}, err
	}
	policyEnv, err := loadDaemonPolicyBundleEnv(getenv)
	if err != nil {
		return daemonSettings{}, err
	}
	controlPlane, err := loadDaemonControlPlaneEnv(getenv)
	if err != nil {
		return daemonSettings{}, err
	}
	intervals, err := loadDaemonIntervalEnv(getenv)
	if err != nil {
		return daemonSettings{}, err
	}
	pprofAddr, err := parseDaemonPprofAddr(getenv(envDaemonPprofAddr), identity.profile)
	if err != nil {
		return daemonSettings{}, err
	}

	return daemonSettings{
		DaemonID:             defaultDaemonID(getenv(envDaemonID), controlPlane.deviceID),
		DaemonVersion:        textutil.Default(getenv(envDaemonVersion), "riido-agentd v0.0.0"),
		Profile:              identity.profile,
		ServerURL:            strings.TrimSpace(getenv(envServerURL)),
		DeviceID:             controlPlane.deviceID,
		DeviceSecret:         controlPlane.deviceSecret,
		DeviceName:           identity.deviceName,
		RuntimeOwner:         identity.owner,
		WorkdirRoot:          workspace.root,
		PolicyBundle:         policyEnv.version,
		PolicyBundlePath:     policyEnv.path,
		PolicyBundleDoc:      policyEnv.doc,
		TaskQueueDir:         controlPlane.taskQueueDir,
		TaskReportDir:        controlPlane.taskReportDir,
		TaskDBSourcePath:     controlPlane.taskDBSourcePath,
		SaaSURL:              controlPlane.saaSURL,
		PollEvery:            intervals.pollEvery,
		IdlePollEvery:        intervals.idlePollEvery,
		HeartbeatEvery:       intervals.heartbeatEvery,
		WorkdirRetention:     workspace.retention,
		WorkdirCleanupEvery:  workspace.cleanupEvery,
		PprofAddr:            pprofAddr,
		WorkspaceCount:       workspace.count,
		RuntimeMaxConcurrent: workspace.runtimeMaxConcurrent,
		RuntimeAgents:        parseRuntimeAgents(getenv(envRuntimeAgents)),
	}, nil
}
