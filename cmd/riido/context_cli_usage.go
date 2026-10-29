package main

import (
	"fmt"
	"io"
)

// printContextUsage lists the work-context commands shared by root and
// direct internal-command help; internal dispatch is documented separately.
func printContextUsage(w io.Writer) {
	fmt.Fprintln(w, "  # Riido work context (agents and people; add --json for the envelope)")
	printWorkUsage(w)
	printContextGroupUsage(w)
	printAgentsUsage(w)
	printMeetingUsage(w)
	printAuthUsage(w)
	fmt.Fprintln(w, "  riido whoami [--json]")
	printDoctorUsage(w)
	fmt.Fprintln(w, "  riido capabilities [--json]")
}

func printWorkUsage(w io.Writer) {
	fmt.Fprintln(w, "  riido work <get|tree|relations|comments|replies|activity|search> ...")
	fmt.Fprintln(w, "    riido work get|tree|relations <RIID-1234|ID> [--json]")
	fmt.Fprintln(w, "    riido work get <ID> --team-id <TEAM-ID> [--json]   (standalone user API)")
	fmt.Fprintln(w, "    riido work comments <RIID-1234|ID> [--cursor C] [--json]   (agent snapshot; fixed 20-item pages)")
	fmt.Fprintln(w, "    riido work comments <ID> --team-id <TEAM-ID> [--page N] [--limit N] [--json]   (user API; defaults 1/20, limit <= 100)")
	fmt.Fprintln(w, "    riido work replies <ID> --team-id <TEAM-ID> --comment-id <COMMENT-ID> [--json]   (user API; no pagination; agent replies unsupported)")
	fmt.Fprintln(w, "    riido work activity <RIID-1234|ID>   (unsupported, exit 8)")
	fmt.Fprintln(w, "    riido work search --query TEXT   (unsupported, exit 8)")
}

func printContextGroupUsage(w io.Writer) {
	fmt.Fprintln(w, "  riido context get [--work <RIID-1234|ID>] [--budget CHARS] [--json]   (agent runs default to the assigned task)")
	fmt.Fprintln(w, "  riido context changes --since SNAPSHOT [--json]   (unsupported, exit 8)")
}

func printAgentsUsage(w io.Writer) {
	fmt.Fprintln(w, "  riido agents list [--cursor C] [--json]   (live: agents the run's requester can see, with work status and connection)")
	fmt.Fprintln(w, "  riido agents show <AGENT-ID> [--json]")
}

func printAuthUsage(w io.Writer) {
	fmt.Fprintln(w, "  riido auth <login|logout|status> [--json]")
	fmt.Fprintln(w, "    riido auth login [--browser]   (browser PKCE; signed credential helper required)")
	fmt.Fprintln(w, "    --desktop and --with-token are unsupported (exit 8); secrets are never accepted as arguments")
}

func printMeetingUsage(w io.Writer) {
	fmt.Fprintln(w, "  riido meeting list [--work <RIID-1234|ID>] [--budget CHARS] [--json]   (snapshot metadata; defaults to the assigned task)")
	fmt.Fprintln(w, "  riido meeting get <MEETING-ID> [--budget CHARS] [--json]   (captured document body)")
}

func printDoctorUsage(w io.Writer) {
	fmt.Fprintln(w, "  riido doctor [--check-api] [--no-input] [--timeout DURATION] [--json]   (--check-api verifies user Whoami; agent API check is unsupported)")
}
