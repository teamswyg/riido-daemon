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
	if *cfg.runProduct {
		cmd += " -run-product"
		cmd += " -client-root " + shellQuote(*cfg.clientRoot)
		cmd += " -product-base-url " + shellQuote(*cfg.productBaseURL)
		cmd += " -product-riido-api-host " + shellQuote(*cfg.productRiidoHost)
		cmd += " -product-storage-state " + shellQuote(*cfg.productStorage)
		if *cfg.startClient {
			cmd += " -product-start-client"
		}
		if *cfg.productWorkspace != "" {
			cmd += " -product-workspace-id " + shellQuote(*cfg.productWorkspace)
		}
		if *cfg.productTeamID != "" {
			cmd += " -product-team-id " + shellQuote(*cfg.productTeamID)
		}
		cmd += productTaskCommandArgs(cfg)
	}
	if *cfg.productEvidence != "" {
		cmd += " -product-evidence " + shellQuote(*cfg.productEvidence)
	}
	return cmd
}

func productTaskCommandArgs(cfg config) string {
	var cmd string
	if *cfg.productTaskID != "" {
		cmd += " -product-task-id " + shellQuote(*cfg.productTaskID)
	}
	if *cfg.productAgentID1 != "" {
		cmd += " -product-agent-id-1 " + shellQuote(*cfg.productAgentID1)
	}
	if *cfg.productAgentID2 != "" {
		cmd += " -product-agent-id-2 " + shellQuote(*cfg.productAgentID2)
	}
	if *cfg.productComment != "" {
		cmd += " -product-comment-body " + shellQuote(*cfg.productComment)
	}
	if !*cfg.taskMutations {
		cmd += " -product-task-mutations=false"
	}
	if !*cfg.taskFixture {
		cmd += " -product-create-task-fixture=false"
	}
	return cmd
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func xmlEscape(b *bytes.Buffer, value string) {
	_ = xml.EscapeText(b, []byte(value))
}
