package main

import (
	"fmt"
	"io"
)

// printContextUsage lists the work-context commands. printUsage includes it
// so `riido --help` keeps every declared command group in one place.
func printContextUsage(w io.Writer) {
	fmt.Fprintln(w, "  # Riido work context (agents and people; add --json for the envelope)")
	printWorkUsage(w)
	printContextGroupUsage(w)
	printAuthUsage(w)
	fmt.Fprintln(w, "  riido whoami [--json]")
	fmt.Fprintln(w, "  riido doctor [--check-api] [--json]")
	fmt.Fprintln(w, "  riido capabilities [--json]")
}

func printWorkUsage(w io.Writer) {
	fmt.Fprintln(w, "  riido work <get|tree|relations|comments|activity|search> ...")
	fmt.Fprintln(w, "    riido work get|tree|relations|activity <RIID-1234|ID> [--json]")
	fmt.Fprintln(w, "    riido work comments <RIID-1234|ID> [--cursor C] [--limit N] [--json]")
	fmt.Fprintln(w, "    riido work search --query TEXT [--limit N] [--json]")
}

func printContextGroupUsage(w io.Writer) {
	fmt.Fprintln(w, "  riido context get [--work <RIID-1234|ID>] [--budget CHARS] [--json]   (agent runs default to the assigned task)")
	fmt.Fprintln(w, "  riido context changes --since SNAPSHOT [--json]")
}

func printAuthUsage(w io.Writer) {
	fmt.Fprintln(w, "  riido auth <login|logout|status> [--json]")
	fmt.Fprintln(w, "    riido auth login --desktop|--browser|--with-token   (--with-token reads stdin; secrets are never accepted as arguments)")
}
