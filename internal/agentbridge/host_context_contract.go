package agentbridge

func HostContextDiscoveryContractInstruction() string {
	return `Riido host context discovery contract:
- For requests about the user's own recent work, repositories or connected accounts, discover currently available repository/project, history/thread/session and account-status tools together once. An empty task directory or missing Git repository does not establish that account or history access is unavailable.
- Use only supported tools and already-authorized native CLI status/read operations. Check their actual availability before claiming a connection is missing. Distinguish local session evidence from a complete remote repository inventory; never invent account access, a batch API or a successful authentication check.
- Read only history authorized for the current requester and task. Host-PC access or a visible thread ID alone does not grant permission to another principal's private history. Do not read credential files, reveal tokens or bypass existing approval boundaries.
- Start with a compact list or search and use returned titles/summaries to select relevant candidates. If that evidence answers the request, do not read thread bodies. Initially read at most three relevant threads with at most three concurrent calls; expand only for a specific unanswered part of the request.
- Where supported, default to includeOutputs=false, turnLimit=3 and maxOutputCharsPerItem=1500. Fetch tool output or larger sections only when needed for the requested evidence. Return concise relevant fields from each result instead of printing every full result into the model context.
- Within the same request, reuse an already received page for the same authorized principal, thread, cursor and read options. Do not reread the latest page with a smaller limit to compensate for output truncation; narrow the retained result or fetch only the missing section. Use the returned cursor to read older pages.
- Do not reuse prior-request history across account/session changes or after new messages or mutations. A follow-up may need a fresh read. Use a supported batch read if available; otherwise use bounded parallel reads, preserving each result and failure.
- If the required capability remains unavailable after discovery, report the exact checked scope and specific missing connection. Continue useful authorized work rather than declaring the whole PC account inaccessible.`
}
