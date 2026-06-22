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
	CommentBody  string
}

func taskMutationPlanFor(
	cfg config,
	payload map[string]any,
	taskID string,
	source string,
) (taskMutationPlan, bool) {
	pair, ok := taskFlowAgentPair(cfg, payload)
	if !ok {
		return taskMutationPlan{}, false
	}
	return taskMutationPlan{
		TaskID:       taskID,
		TaskIDSource: source,
		Pair:         pair,
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

func taskFlowAgentPair(cfg config, payload map[string]any) (taskAgentPair, bool) {
	if *cfg.firstAgentID != "" && *cfg.secondAgentID != "" {
		return taskAgentPair{
			First:  taskAgentCandidate{AgentID: *cfg.firstAgentID},
			Second: taskAgentCandidate{AgentID: *cfg.secondAgentID},
		}, true
	}
	return chooseTaskAgentPair(taskAgentCandidates(payload))
}

func taskCommentBody(cfg config, taskID string) string {
	if body := strings.TrimSpace(*cfg.commentBody); body != "" {
		return body
	}
	return "local QA thread message for " + taskID
}
