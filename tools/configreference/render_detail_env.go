package main

import (
	"fmt"
	"strings"
)

func renderEnvTable(b *strings.Builder, manifest Manifest, names []string) {
	b.WriteString("| Variable | Consumer | Default | Meaning |\n| --- | --- | --- | --- |\n")
	for _, name := range names {
		row, ok := envVarByName(manifest, name)
		if !ok {
			continue
		}
		fmt.Fprintf(b, "| `%s` | %s | %s | %s |\n", row.Name, row.Consumer, row.Default, row.Meaning)
	}
	b.WriteByte('\n')
}

func envVarByName(manifest Manifest, name string) (EnvVar, bool) {
	for _, row := range manifest.DaemonEnvVars {
		if row.Name == name {
			return row, true
		}
	}
	return EnvVar{}, false
}
