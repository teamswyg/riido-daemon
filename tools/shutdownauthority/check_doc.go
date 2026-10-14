package main

import "os"

func checkDoc(opts options, manifest Manifest, rendered string) []problem {
	if !opts.CheckDoc {
		return nil
	}
	path, err := cleanRepoPath(opts.Repo, manifest.GeneratedDoc)
	if err != nil {
		return []problem{{err.Error()}}
	}
	actual, err := os.ReadFile(path)
	if err != nil {
		return []problem{{err.Error()}}
	}
	if string(actual) != rendered {
		return []problem{{"generated doc drift: run go run ./tools/shutdownauthority -write-doc"}}
	}
	return nil
}
