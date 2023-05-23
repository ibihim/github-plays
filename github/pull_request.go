package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v52/github"
)

// PullRequestManager manages a pull request.
type PullRequestManager struct {
	client *github.Client
	owner  string
	repo   string
	num    int
}

// NewPullRequestManager returns a new PullRequestManager.
func NewPullRequestManager(client *github.Client, owner, repo string, num int) *PullRequestManager {
	return &PullRequestManager{
		client: client,
		owner:  owner,
		repo:   repo,
		num:    num,
	}
}

// GetChecks returns the list of checks for a given ref.
// It returns the list of success, pending and failure checks.
func (prm *PullRequestManager) GetChecks(ctx context.Context) ([]string, []string, []string, error) {
	pr, _, err := prm.client.PullRequests.Get(ctx, prm.owner, prm.repo, prm.num)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get PR: %w", err)
	}

	rs, _, err := prm.client.Repositories.ListStatuses(ctx, prm.owner, prm.repo, pr.GetHead().GetSHA(), nil)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to list statuses: %w", err)
	}

	success := []string{}
	failure := []string{}
	pending := []string{}

	for _, r := range rs {
		if r.GetState() == "success" {
			success = append(success, r.GetContext())
		} else if r.GetState() == "failure" {
			failure = append(failure, r.GetContext())
		} else if r.GetState() == "pending" {
			pending = append(pending, r.GetContext())
		}
	}

	return success, pending, failure, nil
}

// WriteComment writes a comment to a given PR.
func (prm *PullRequestManager) WriteComment(ctx context.Context, message string) error {
	comment := &github.IssueComment{
		Body: &message,
	}

	_, _, err := prm.client.Issues.CreateComment(ctx, prm.owner, prm.repo, prm.num, comment)
	if err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}

	return nil
}
