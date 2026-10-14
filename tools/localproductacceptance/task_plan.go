package main

import (
	"strings"
	"time"
)

type taskAgentPair struct {
	First  taskAgentCandidate
	Second taskAgentCandidate
}

type taskMutationPlan struct {
	TaskID       string
	TaskIDSource string
	Pair         taskAgentPair
	Candidates   []taskAgentCandidate
	CommentBody  string
}

func taskMutationPlanFor(
	cfg config,
	payload map[string]any,
	taskID string,
	source string,
	agents agentFixture,
) (taskMutationPlan, bool) {
	candidates := taskFlowAgentCandidates(cfg, payload, agents)
	if len(candidates) < 2 {
		return taskMutationPlan{}, false
	}
	return taskMutationPlan{
		TaskID:       taskID,
		TaskIDSource: source,
		Pair:         taskAgentPair{First: candidates[0], Second: candidates[1]},
		Candidates:   candidates,
		CommentBody:  taskCommentBody(cfg, taskID),
	}, true
}

func taskFlowTaskID(cfg config, discovery map[string]any) (string, string) {
	if taskID := strings.TrimSpace(*cfg.taskID); taskID != "" {
		return taskID, "configured"
	}
	if taskID := firstAssignedProfileTaskID(discovery); taskID != "" {
		return taskID, "assigned-agent-profiles"
	}
	return "local-qa-" + time.Now().UTC().Format("20060102T150405Z"), "generated"
}

func taskFlowAgentCandidates(cfg config, payload map[string]any, agents agentFixture) []taskAgentCandidate {
	if *cfg.firstAgentID != "" && *cfg.secondAgentID != "" {
		return []taskAgentCandidate{
			{AgentID: *cfg.firstAgentID},
			{AgentID: *cfg.secondAgentID},
		}
	}
	if len(agents.Candidates) >= 2 {
		return append([]taskAgentCandidate(nil), agents.Candidates...)
	}
	return prioritizeTaskAgentCandidates(taskAgentCandidates(payload))
}

func taskCommentBody(cfg config, taskID string) string {
	if body := strings.TrimSpace(*cfg.commentBody); body != "" {
		return body
	}
	return "local QA thread message for " + taskID
}
