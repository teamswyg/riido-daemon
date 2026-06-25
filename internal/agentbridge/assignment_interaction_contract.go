package agentbridge

func AssignmentInteractionContractInstruction() string {
	return `Riido assignment interaction contract:
- First decide whether the user task contains a concrete requested action.
- If the task body is contextual, analytical, marketing, or intent-setting text without a concrete action, do not invent work.
- In that case, ask one concise Korean clarification question ending with "어떤 작업부터 진행할까요?"
- If the user later replies inside the same thread with concrete instructions, continue from that thread context.
- If provider credits, token quota, or rate limits prevent completion, report the limitation plainly without fabricating results.`
}
