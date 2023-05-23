package github

import (
	"context"
	"fmt"
	"log"

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
	log.Printf("Want PR:       %s/%s#%d\n", prm.owner, prm.repo, prm.num)
	pr, _, err := prm.client.PullRequests.Get(ctx, prm.owner, prm.repo, prm.num)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get PR: %w", err)
	}

	log.Printf("Have HEAD ref: %s/%s@%s\n", prm.owner, prm.repo, pr.GetHead().GetSHA())
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

	log.Printf("Have %d success checks\n", len(success))
	log.Printf("Have %d pending checks\n", len(pending))
	log.Printf("Have %d failure checks\n", len(failure))

	return success, pending, failure, nil
}

// WriteComment writes a comment to a given PR.
func (prm *PullRequestManager) WriteComment(ctx context.Context, message string) error {
	comment := &github.IssueComment{
		Body: &message,
	}

	log.Printf("Want comment: %s", message)
	_, _, err := prm.client.Issues.CreateComment(ctx, prm.owner, prm.repo, prm.num, comment)
	if err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}

	log.Printf(", done\n")

	return nil
}
