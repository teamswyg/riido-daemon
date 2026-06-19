package main

type BehaviorEvidence struct {
	Name string `json:"name"`
	OK   bool   `json:"ok"`
}

func validateBehaviors(m Manifest) ([]problem, []BehaviorEvidence) {
	var problems []problem
	out := make([]BehaviorEvidence, 0, len(manifestBehaviors(m)))
	for _, name := range manifestBehaviors(m) {
		err := runBehavior(name)
		if err != nil {
			problems = append(problems, problem{Message: err.Error()})
		}
		out = append(out, BehaviorEvidence{Name: name, OK: err == nil})
	}
	return problems, out
}

func runBehavior(name string) error {
	switch name {
	case "path_order":
		return verifyPathOrder()
	case "override_only":
		return verifyOverrideOnly()
	case "override_fail_closed":
		return verifyOverrideFailClosed()
	case "spawn_launch_path":
		return verifySpawnLaunchPath()
	case "spawn_explicit_path":
		return verifySpawnExplicitPath()
	default:
		return behaviorError("unknown behavior " + name)
	}
}
