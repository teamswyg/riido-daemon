package main

import "github.com/teamswyg/riido-contracts/assignment"

func buildFSMSnapshot() FSMSnapshot {
	fsm := assignment.GeneratedAssignmentFSM()
	return FSMSnapshot{
		Name:           fsm.Name(),
		TypeUnion:      string(fsm.TypeUnion()),
		States:         stateNames(fsm.States()),
		StartStates:    stateNames(fsm.StartStates()),
		EndStates:      stateNames(fsm.EndStates()),
		TerminalStates: stateNames(fsm.TerminalStates()),
		AgentActive:    activeStateNames(fsm.States()),
		Transitions:    transitionNames(fsm.Transitions()),
		Mermaid:        fsm.Mermaid(),
	}
}

func stateNames(codes []assignment.AssignmentStateCode) []string {
	out := make([]string, len(codes))
	for i, code := range codes {
		out[i] = code.String()
	}
	return out
}

func activeStateNames(codes []assignment.AssignmentStateCode) []string {
	var out []string
	for _, code := range codes {
		if code.IsAgentActive() {
			out = append(out, code.String())
		}
	}
	return out
}

func transitionNames(codes []assignment.AssignmentTransitionCode) []Transition {
	out := make([]Transition, len(codes))
	for i, code := range codes {
		out[i] = Transition{From: code.From.String(), To: code.To.String()}
	}
	return out
}
