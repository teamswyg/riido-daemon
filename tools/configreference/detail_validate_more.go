package main

import "fmt"

func validateTable(path string, table *DetailTable) []problem {
	if table == nil {
		return []problem{{Message: "missing detail table in " + path}}
	}
	if len(table.Headers) == 0 || len(table.Rows) == 0 {
		return []problem{{Message: "empty detail table in " + path}}
	}
	for _, row := range table.Rows {
		if len(row) != len(table.Headers) {
			return []problem{{Message: "detail table row width mismatch in " + path}}
		}
	}
	return nil
}

func validateEnvNames(manifest Manifest, path string, names []string) []problem {
	if len(names) == 0 {
		return []problem{{Message: "empty env_table in " + path}}
	}
	var problems []problem
	for _, name := range names {
		if _, ok := envVarByName(manifest, name); !ok {
			problems = append(problems, problem{Message: fmt.Sprintf("unknown env %s in %s", name, path)})
		}
	}
	return problems
}
