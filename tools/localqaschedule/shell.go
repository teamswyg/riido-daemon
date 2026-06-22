package main

import (
	"bytes"
	"encoding/xml"
	"strings"
)

const launchdPath = "/opt/homebrew/bin:/usr/local/bin:/usr/local/go/bin:/usr/bin:/bin:/usr/sbin:/sbin"

func localQACommand(cfg config, paths schedulePaths) string {
	cmd := "cd " + shellQuote(paths.repo) + " && PATH=" + shellQuote(launchdPath)
	cmd += " go run ./tools/localqarunner"
	if *cfg.s3Prefix != "" {
		cmd += " -s3-prefix " + shellQuote(*cfg.s3Prefix)
	}
	return cmd
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func xmlEscape(b *bytes.Buffer, value string) {
	_ = xml.EscapeText(b, []byte(value))
}
