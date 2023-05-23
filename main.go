package main

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/oauth2"

	gogithub "github.com/google/go-github/v52/github"

	"github.com/ibihim/github-plays/cmd"
	"github.com/ibihim/github-plays/github"
)

func main() {
	if err := cmd.RootCommand().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func app() error {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)

	tc := oauth2.NewClient(ctx, ts)
	client := gogithub.NewClient(tc)

	prm := github.NewPullRequestManager(client, "openshift", "cluster-kube-apiserver-operator", 1493)

	_, _, failures, err := prm.GetChecks(ctx)
	if err != nil {
		return fmt.Errorf("failed to get checks: %w", err)
	}

	if len(failures) > 0 {
		if err := prm.WriteComment(ctx, "/retest-required"); err != nil {
			return fmt.Errorf("failed to write comment: %w", err)
		}
	}

	return nil
}
