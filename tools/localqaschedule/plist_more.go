package main

import (
	"bytes"
	"fmt"
)

func plistArgs(b *bytes.Buffer, command string) {
	b.WriteString("<key>ProgramArguments</key><array>\n")
	b.WriteString("<string>/bin/zsh</string>\n")
	b.WriteString("<string>-lc</string>\n")
	b.WriteString("<string>")
	xmlEscape(b, command)
	b.WriteString("</string>\n")
	b.WriteString("</array>\n")
}

func plistCalendar(b *bytes.Buffer, hour, minute int) {
	b.WriteString("<key>StartCalendarInterval</key><dict>\n")
	fmt.Fprintf(b, "<key>Hour</key><integer>%d</integer>\n", hour)
	fmt.Fprintf(b, "<key>Minute</key><integer>%d</integer>\n", minute)
	b.WriteString("</dict>\n")
}
