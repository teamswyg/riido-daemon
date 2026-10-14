package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
)

func renderPlist(cfg config, paths schedulePaths) string {
	var b bytes.Buffer
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "`)
	b.WriteString(`http://www.apple.com/DTDs/PropertyList-1.0.dtd">` + "\n")
	b.WriteString(`<plist version="1.0"><dict>` + "\n")
	plistString(&b, "Label", *cfg.label)
	plistArgs(&b, localQACommand(cfg, paths))
	plistCalendar(&b, *cfg.hour, *cfg.minute)
	plistBool(&b, "RunAtLoad", *cfg.runAtLoad)
	plistString(&b, "StandardOutPath", paths.stdout)
	plistString(&b, "StandardErrorPath", paths.stderr)
	b.WriteString("</dict></plist>\n")
	return b.String()
}

func plistString(b *bytes.Buffer, key, value string) {
	fmt.Fprintf(b, "<key>%s</key><string>", key)
	_ = xml.EscapeText(b, []byte(value))
	b.WriteString("</string>\n")
}

func plistBool(b *bytes.Buffer, key string, value bool) {
	fmt.Fprintf(b, "<key>%s</key>", key)
	if value {
		b.WriteString("<true/>\n")
		return
	}
	b.WriteString("<false/>\n")
}
