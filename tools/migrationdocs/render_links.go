package main

import (
	"fmt"
	"strings"
)

func writeLinks(b *strings.Builder, links []link) {
	for _, link := range links {
		fmt.Fprintf(b, "- [%s](%s)\n", link.Title, link.Path)
	}
}
