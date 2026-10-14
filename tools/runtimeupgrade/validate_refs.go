package main

import "fmt"

func validateReferences(m Manifest) []problem {
	known, problems := knownSourceChecks(m.SourceChecks)
	for _, group := range ruleGroups(m) {
		for _, rule := range group.rules {
			problems = append(problems, validateRule(group.section, rule, known)...)
		}
	}
	return problems
}

func knownSourceChecks(checks []SourceCheck) (map[string]bool, []problem) {
	known := map[string]bool{}
	var problems []problem
	for _, check := range checks {
		if known[check.Name] {
			problems = append(problems, problem{fmt.Sprintf("duplicate source check %q", check.Name)})
		}
		known[check.Name] = true
	}
	return known, problems
}

func validateRule(section string, rule Rule, known map[string]bool) []problem {
	switch rule.Status {
	case "implemented":
		return validateImplementedRule(section, rule, known)
	case "reserved":
		return validateReservedRule(section, rule)
	default:
		return []problem{{fmt.Sprintf("%s %q has unknown status %q", section, ruleLabel(rule), rule.Status)}}
	}
}
