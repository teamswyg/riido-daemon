package main

type ruleGroup struct {
	section string
	rules   []Rule
}

func ruleGroups(m Manifest) []ruleGroup {
	return []ruleGroup{
		{"inputs", m.Inputs},
		{"flow", m.Flow},
		{"policies", m.Policies},
		{"native_config", m.NativeConfig},
	}
}

func ruleLabel(rule Rule) string {
	if rule.Name != "" {
		return rule.Name
	}
	return rule.Step
}

func collectReserved(m Manifest) []ReservedRule {
	var out []ReservedRule
	for _, group := range ruleGroups(m) {
		for _, rule := range group.rules {
			if rule.Status == "reserved" {
				out = append(out, ReservedRule{
					Section: group.section, Name: ruleLabel(rule),
					RequiredEvidence: rule.RequiredEvidence, DecisionRefs: rule.DecisionRefs,
				})
			}
		}
	}
	return out
}
