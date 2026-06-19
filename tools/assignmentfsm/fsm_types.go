package main

type FSMSnapshot struct {
	Name           string       `json:"name"`
	TypeUnion      string       `json:"type_union"`
	States         []string     `json:"states"`
	StartStates    []string     `json:"start_states"`
	EndStates      []string     `json:"end_states"`
	TerminalStates []string     `json:"terminal_states"`
	AgentActive    []string     `json:"agent_active_states"`
	Transitions    []Transition `json:"transitions"`
	Mermaid        string       `json:"mermaid"`
}

type Transition struct {
	From string `json:"from"`
	To   string `json:"to"`
}
