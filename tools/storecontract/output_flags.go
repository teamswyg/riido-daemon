package main

import "errors"

func selectedOutputPath(outPath, evidenceOutPath string) (string, error) {
	if outPath != "" && evidenceOutPath != "" && outPath != evidenceOutPath {
		return "", errors.New("-out and -evidence-out must match when both are set")
	}
	if evidenceOutPath != "" {
		return evidenceOutPath, nil
	}
	return outPath, nil
}
