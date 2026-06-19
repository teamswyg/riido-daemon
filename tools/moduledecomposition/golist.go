package main

import (
	"bytes"
	"encoding/json"
	"os/exec"
)

func loadPackages(repo string) (map[string]packageInfo, []checkResult) {
	cmd := exec.Command("go", "list", "-json", "./...")
	cmd.Dir = repo
	out, err := cmd.Output()
	if err != nil {
		return nil, []checkResult{{Name: "go-list", Pass: false, Detail: err.Error()}}
	}
	packages, decodeErr := decodePackages(out)
	result := checkResult{Name: "go-list", Pass: decodeErr == nil}
	if decodeErr != nil {
		result.Detail = decodeErr.Error()
	}
	return packages, []checkResult{result}
}

func decodePackages(out []byte) (map[string]packageInfo, error) {
	dec := json.NewDecoder(bytes.NewReader(out))
	packages := map[string]packageInfo{}
	for dec.More() {
		var pkg packageInfo
		if err := dec.Decode(&pkg); err != nil {
			return packages, err
		}
		packages[pkg.ImportPath] = pkg
	}
	return packages, nil
}
