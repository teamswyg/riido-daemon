package supervisor

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

// Keep partial clones, including a PR's initialized but unfetched .git, out of
// the provider's working directory. Only publish a successful checkout.
func (a *Actor) prepareAssignmentCheckout(ctx context.Context, dir string, req *bridge.TaskRequest) error {
	url, err := assignmentCloneURL(req.Worktree)
	if err != nil || url == "" {
		return err
	}
	info, err := os.Lstat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return errors.New("assignment workspace must be a directory")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	if len(entries) != 0 {
		return errors.New("fresh assignment workspace is not empty")
	}

	stage, err := os.MkdirTemp(filepath.Dir(dir), ".checkout-")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(stage) }()
	var pull string
	err = materializeAssignmentWorktree(ctx, stage, req.Worktree, func(number string) { pull = number })
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if err != nil {
		var diagnostic *assignmentGitCloneFailure
		var message string
		switch {
		case errors.As(err, &diagnostic):
			message = diagnostic.reason
		case errors.Is(err, errAssignmentGitUnavailable):
			message = "Git is not installed on the daemon device."
		default:
			return err
		}
		if err := os.RemoveAll(stage); err != nil {
			return fmt.Errorf("discard incomplete checkout: %w", err)
		}
		req.AssignmentCheckoutUnavailable = true
		appendUnavailableCheckoutPrompt(req, message)
		a.reportCheckoutProgress(ctx, req, "저장소를 내려받지 못해 격리된 작업 공간에서 계속합니다.")
		return nil
	}
	// Remove only the empty destination so the rename also works on Windows.
	if err := os.Remove(dir); err != nil {
		return err
	}
	if err := os.Rename(stage, dir); err != nil {
		return err
	}
	if pull != "" {
		req.Prompt += "\n\nThe worktree has been checked out at the exact head of linked pull request #" + pull + ". Use this checkout as the starting point."
		a.reportCheckoutProgress(ctx, req, "Using linked pull request #"+pull+" for this execution.")
	}
	return nil
}

func appendUnavailableCheckoutPrompt(req *bridge.TaskRequest, reason string) {
	note := "\n\nRiido could not prepare the requested local checkout. " + reason + " Continue the task in the isolated workspace using the available authorized tools. For PR work, use the connected GitHub tools if available. If access is needed, ask the user in this conversation. Report precisely which files and checks you could verify."
	req.Prompt += note
	if req.ProviderNeutralPrompt != "" {
		req.ProviderNeutralPrompt += note
	}
}

func (a *Actor) reportCheckoutProgress(ctx context.Context, req *bridge.TaskRequest, message string) {
	if a.cfg.Reporter != nil {
		_ = a.cfg.Reporter.ReportEvent(controlplane.ContextWithTaskReport(ctx, reportContextFor(req)), req.ID, agentbridge.Event{Kind: agentbridge.EventProgress, Text: message})
	}
}
