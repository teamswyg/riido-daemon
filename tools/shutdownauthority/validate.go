package main

func validate(repo string, manifest Manifest) ([]problem, []LevelCheck, []TimeoutCheck, []ConsumerCheck) {
	levelSrc, err := readSource(repo, manifest.Sources.Levels)
	if err != nil {
		return []problem{{err.Error()}}, nil, nil, nil
	}
	timeoutSrc, err := readSource(repo, manifest.Sources.Timeouts)
	if err != nil {
		return []problem{{err.Error()}}, nil, nil, nil
	}
	levelProblems, levels := validateLevels(manifest, parseLevels(levelSrc))
	timeoutProblems, timeouts := validateTimeouts(manifest, parseTimeouts(timeoutSrc))
	consumerProblems, consumers := validateConsumers(repo, manifest)
	problems := levelProblems
	problems = append(problems, timeoutProblems...)
	problems = append(problems, consumerProblems...)
	return problems, levels, timeouts, consumers
}
