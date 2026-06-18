package main

import "time"

type daemonIntervalEnv struct {
	pollEvery      time.Duration
	idlePollEvery  time.Duration
	heartbeatEvery time.Duration
}

func loadDaemonIntervalEnv(getenv func(string) string) (daemonIntervalEnv, error) {
	pollEvery, err := parseOptionalPositiveDurationSeconds(getenv(envDaemonPollIntervalSeconds), envDaemonPollIntervalSeconds, time.Second)
	if err != nil {
		return daemonIntervalEnv{}, err
	}
	idlePollEvery, err := parseOptionalPositiveDurationSeconds(getenv(envDaemonIdlePollIntervalSeconds), envDaemonIdlePollIntervalSeconds, 5*time.Second)
	if err != nil {
		return daemonIntervalEnv{}, err
	}
	if idlePollEvery < pollEvery {
		return daemonIntervalEnv{}, daemonErrorf(ErrDaemonConfig, "settings.validate-intervals", "%s must be greater than or equal to %s", envDaemonIdlePollIntervalSeconds, envDaemonPollIntervalSeconds)
	}
	heartbeatEvery, err := parseOptionalPositiveDurationSeconds(getenv(envDaemonHeartbeatSeconds), envDaemonHeartbeatSeconds, 5*time.Second)
	if err != nil {
		return daemonIntervalEnv{}, err
	}
	return daemonIntervalEnv{
		pollEvery:      pollEvery,
		idlePollEvery:  idlePollEvery,
		heartbeatEvery: heartbeatEvery,
	}, nil
}
