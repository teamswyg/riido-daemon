package main

import "fmt"

func validateImplementedRule(section string, rule Rule, known map[string]bool) []problem {
	if len(rule.SourceChecks) == 0 {
		return []problem{{fmt.Sprintf("%s %q is implemented without source checks", section, ruleLabel(rule))}}
	}
	var problems []problem
	for _, ref := range rule.SourceChecks {
		if !known[ref] {
			problems = append(problems, problem{fmt.Sprintf("%s %q references unknown source check %q", section, ruleLabel(rule), ref)})
		}
	}
	return problems
}

func validateReservedRule(section string, rule Rule) []problem {
	if rule.RequiredEvidence == "" {
		return []problem{{fmt.Sprintf("%s %q is reserved without required_evidence", section, ruleLabel(rule))}}
	}
	if len(rule.SourceChecks) > 0 {
		return []problem{{fmt.Sprintf("%s %q is reserved but carries implemented source checks", section, ruleLabel(rule))}}
	}
	return nil
}
