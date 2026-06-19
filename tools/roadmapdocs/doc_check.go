package main

import "os"

func maybeWriteDoc(opts options, m manifest, body string) error {
	if !opts.WriteDoc {
		return nil
	}
	return writeText(repoPath(opts.Repo, m.GeneratedDoc), body)
}

func checkDoc(opts options, m manifest, body string) []string {
	if !opts.CheckDoc {
		return nil
	}
	data, err := os.ReadFile(repoPath(opts.Repo, m.GeneratedDoc))
	if err != nil {
		return []string{err.Error()}
	}
	if string(data) != body {
		return []string{"generated doc drift: run go run ./tools/roadmapdocs -write-doc"}
	}
	return nil
}
