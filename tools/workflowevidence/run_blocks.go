package main

import "strings"

func workflowRunBlocks(text string) []string {
	lines := strings.Split(text, "\n")
	var blocks []string
	for i := 0; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		trimmed = strings.TrimPrefix(trimmed, "- ")
		value, ok := strings.CutPrefix(trimmed, "run:")
		if !ok {
			continue
		}
		value = strings.TrimSpace(value)
		if value != "|" && value != ">" {
			blocks = append(blocks, value)
			continue
		}
		block, next := workflowRunBlock(lines, i+1, leadingSpaces(lines[i]))
		blocks = append(blocks, block)
		i = next - 1
	}
	return blocks
}

func workflowRunBlock(lines []string, start, runIndent int) (string, int) {
	var body []string
	i := start
	for ; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "" {
			continue
		}
		if leadingSpaces(lines[i]) <= runIndent {
			break
		}
		body = append(body, strings.TrimSpace(lines[i]))
	}
	return strings.Join(body, "\n"), i
}
