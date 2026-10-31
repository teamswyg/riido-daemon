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
	printResourceUsage(w)
	printRepoUsage(w)
	printThreadUsage(w)
	printContextGroupUsage(w)
	printAgentsUsage(w)
	printMeetingUsage(w)
	printAuthUsage(w)
	fmt.Fprintln(w, "  riido whoami [--json]")
	printDoctorUsage(w)
	fmt.Fprintln(w, "  riido capabilities [--json]")
}

func printWorkUsage(w io.Writer) {
	fmt.Fprintln(w, "  riido work search --workspace <WORKSPACE-ID> --query <TEXT> [--cursor C] [--limit N] [--json]   (user live authorized title/key/body search; agent unsupported)")
	fmt.Fprintln(w, "  riido work tree <ID> --team-id <TEAM-ID> [--json]   (user live one-hop hierarchy; up to 10 selected children; no pagination)")
	fmt.Fprintln(w, "  riido work <get|body|tree|relations|comments|replies|activity|search> ...")
	fmt.Fprintln(w, "    riido work get|tree|relations <RIID-1234|ID> [--json]")
	fmt.Fprintln(w, "    riido work get <ID> --team-id <TEAM-ID> [--json]   (standalone user API)")
	fmt.Fprintln(w, "    riido work relations <ID> --team-id <TEAM-ID> [--json]   (user live direct relations; no pagination; agent keeps snapshot reads)")
	fmt.Fprintln(w, "    riido work body <ID> --team-id <TEAM-ID> --version <VERSION> [--kind component|backlog|intake] [--part document|description|document_alternate|metadata|comments_page] [--offset N] [--limit BYTES] [--json]")
	fmt.Fprintln(w, "      comments_page: copy contentPage.selector with [--comment-id ID] [--page N] [--comments-limit N]; keep it unchanged while reading byte pages")
	fmt.Fprintln(w, "    riido work comments <RIID-1234|ID> [--cursor C] [--json]   (agent snapshot; --cursor is for legacy 20-item pages; new captured parts use riido context part)")
	fmt.Fprintln(w, "    riido work comments <ID> --team-id <TEAM-ID> [--page N] [--limit N] [--json]   (user API; defaults 1/20, limit <= 100)")
	fmt.Fprintln(w, "    riido work replies <ID> --team-id <TEAM-ID> --comment-id <COMMENT-ID> [--json]   (user API; no pagination; agent replies unsupported)")
	fmt.Fprintln(w, "    riido work comments|replies <ID> --kind component|backlog|intake --team-id <TEAM-ID> [--comment-id <PARENT-ID>] [--page N] [--limit N] [--version <VERSION>] [--json]   (user read-scope API; replies need parent ID, page > 1 needs original version)")
	fmt.Fprintln(w, "    riido work activity <ID> --team-id <TEAM-ID> [--page N] [--limit N] [--json]   (user live activity; defaults 1/20, limit <= 100; agent unsupported, exit 8)")
	fmt.Fprintln(w, "    riido work search --query TEXT   (agent mode unsupported, exit 8)")
}

func printContextGroupUsage(w io.Writer) {
	fmt.Fprintln(w, "  riido context get [--work <RIID-1234|ID>] [--budget CHARS] [--json]   (agent runs default to the assigned task)")
	fmt.Fprintln(w, "  riido context get --work <ID> --team-id <TEAM-ID> [--json]   (user API; direct context within one relation hop)")
	fmt.Fprintln(w, "  riido context part <SOURCE-ID> --cursor <CURSOR> [--budget CHARS] [--json]   (agent snapshot; one call returns consecutive captured pages of the part, about 48 KiB by default; continue with next_cursor)")
	fmt.Fprintln(w, "  riido context changes --since SNAPSHOT [--json]   (unsupported, exit 8)")
}

func printAgentsUsage(w io.Writer) {
	fmt.Fprintln(w, "  riido agents list [--cursor C] [--json]   (live: agents the run's requester can see, with work status and connection)")
	fmt.Fprintln(w, "  riido agents show <AGENT-ID> [--json]")
}

func printAuthUsage(w io.Writer) {
	fmt.Fprintln(w, "  riido auth <login|link|repair|logout|status> [--json]")
	fmt.Fprintln(w, "    riido auth login [--browser]   (browser PKCE; signed credential helper required)")
	fmt.Fprintln(w, "    riido auth link               (explicit device approval; resumes the original protected request)")
	fmt.Fprintln(w, "    riido auth link --no-input    (receipt recovery only; fails if protected storage requires OS confirmation)")
	fmt.Fprintln(w, "    riido auth repair            (existing key + fresh browser consent; no OAuth login or grant restoration)")
	fmt.Fprintln(w, "    riido auth repair --restart  (explicit new consent after a terminal request; active requests are preserved)")
	fmt.Fprintln(w, "    riido auth repair --no-input (existing receipt recovery only; OS protection still applies)")
	fmt.Fprintln(w, "    --desktop and --with-token are unsupported (exit 8); secrets are never accepted as arguments")
}

func printMeetingUsage(w io.Writer) {
	fmt.Fprintln(w, "  riido meeting list [--work <RIID-1234|ID>] [--budget CHARS] [--json]   (snapshot metadata; defaults to the assigned task)")
	fmt.Fprintln(w, "  riido meeting get <MEETING-ID> [--budget CHARS] [--json]   (captured document body)")
	fmt.Fprintln(w, "  riido meeting get <MEETING-ID> --team-id <TEAM-ID> [--cursor C] [--limit N] [--json]   (user API; details, body and board; default 100 cards, maximum 200)")
	fmt.Fprintln(w, "  riido meeting comments <MEETING-ID> --team-id <TEAM-ID> [--comment-id PARENT-ID] [--page N] [--limit N] [--version VERSION] [--json]   (user API; default 20, maximum 100; page > 1 requires original version)")
	fmt.Fprintln(w, "  riido meeting body <MEETING-ID> --team-id <TEAM-ID> --version <VERSION> [--part document|description|document_alternate|metadata|comments_page|board_page] [--offset N] [--limit BYTES] [--json]")
	fmt.Fprintln(w, "    comments_page: copy contentPage.selector with [--comment-id PARENT-ID] [--page N] [--comments-limit N]; keep it unchanged while reading byte pages")
	fmt.Fprintln(w, "    board_page: copy contentPage.selector with [--board-cursor CURSOR] [--board-limit N]; keep it unchanged while reading byte pages")
}

func printDoctorUsage(w io.Writer) {
	fmt.Fprintln(w, "  riido doctor [--check-api] [--no-input] [--timeout DURATION] [--json]   (--check-api verifies user Whoami; agent API check is unsupported)")
}
