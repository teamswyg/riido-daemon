package main

import "os"

func maybeWriteDoc(opts options, manifest Manifest, rendered string) error {
	if !opts.WriteDoc {
		return nil
	}
	return writeText(repoPath(opts.Repo, manifest.GeneratedDoc), rendered)
}

func checkDoc(opts options, manifest Manifest, rendered string) []problem {
	if !opts.CheckDoc {
		return nil
	}
	data, err := os.ReadFile(repoPath(opts.Repo, manifest.GeneratedDoc))
	if err != nil {
		return []problem{{Message: err.Error()}}
	}
	if string(data) != rendered {
		return []problem{{Message: "generated doc drift: run go run ./tools/configreference -write-doc"}}
	}
	return nil
}
