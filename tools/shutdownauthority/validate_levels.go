package main

import "fmt"

func validateLevels(manifest Manifest, source levelSource) ([]problem, []LevelCheck) {
	var problems []problem
	checks := make([]LevelCheck, 0, len(manifest.Levels))
	for _, row := range manifest.Levels {
		check := LevelCheck{
			Const: row.Const, ExpectedName: row.Name,
			ActualName: source.Names[row.Const], ExpectedOrder: row.Order,
			ActualOrder: source.Order[row.Const],
		}
		check.Pass = check.ExpectedName == check.ActualName &&
			check.ExpectedOrder == check.ActualOrder
		if !check.Pass {
			problems = append(problems, problem{fmt.Sprintf("shutdown level drift: %s", row.Const)})
		}
		checks = append(checks, check)
	}
	return problems, checks
}
