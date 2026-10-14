package main

func checkBehaviors(repo string, manifest Manifest) []CheckResult {
	return []CheckResult{
		checkHelpOutput(repo, manifest),
		checkBridgeProviders(repo, manifest),
	}
}

func appendFailedProblems(problems []problem, prefix string, results []CheckResult) []problem {
	return append(problems, resultProblems(results, prefix)...)
}
