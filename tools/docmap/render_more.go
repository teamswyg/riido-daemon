package main

import (
	"fmt"
	"strings"
)

func renderLinks(docPath string, docs []string) string {
	links := make([]string, 0, len(docs))
	for _, doc := range docs {
		links = append(links, fmt.Sprintf("[`%s`](%s)", linkLabel(doc), linkFrom(docPath, doc)))
	}
	return strings.Join(links, ", ")
}

func renderRepos(b *strings.Builder, repos []repo) {
	b.WriteString("## Repo 간 책임 경계\n\n| Repo | 책임 |\n| --- | --- |\n")
	for _, item := range repos {
		fmt.Fprintf(b, "| `%s` | %s |\n", item.Repo, item.Responsibility)
	}
	b.WriteByte('\n')
}

func renderRules(b *strings.Builder, rules []string) {
	b.WriteString("## 작업 규칙\n\n")
	for _, rule := range rules {
		fmt.Fprintf(b, "- %s\n", rule)
	}
}
