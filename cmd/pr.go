package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	gogithub "github.com/google/go-github/v52/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"

	"github.com/ibihim/github-plays/github"
)

func RootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gp",
		Short: "gp is a CLI tool to manage GitHub day to day tasks.",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			flag.CommandLine.VisitAll(func(f *flag.Flag) {
				fmt.Printf("Flag: --%s=%q\n", f.Name, f.Value)
			})
		},
	}

	cmd.AddCommand(PR())

	return cmd
}

func PR() *cobra.Command {
	o := Options{}

	cmd := &cobra.Command{
		Use:   "pr",
		Short: "pr is a command to manage pull requests.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// This is where you would actually execute the action of the command.
			c, err := Complete(o)
			if err != nil {
				return fmt.Errorf("failed to complete: %w", err)
			}

			if err := Validate(c); err != nil {
				return fmt.Errorf("failed to validate: %w", err)
			}

			return PRRun(c)
		},
	}

	cmd.Flags().StringVarP(&o.Owner, "owner", "o", "", "Repository owner")
	cmd.Flags().StringVarP(&o.Repo, "repo", "r", "", "Repository name")
	cmd.Flags().IntVarP(&o.Num, "number", "n", 0, "Pull request number")
	cmd.Flags().StringVarP(&o.URL, "url", "u", "", "Pull request URL")
	cmd.Flags().StringVarP(&o.Token, "token", "t", "", "GitHub token")

	return cmd
}

func PRRun(c *Config) error {
	ctx := context.Background()
	prm := github.NewPullRequestManager(c.Client, c.Owner, c.Repo, c.Num)

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

// Options is a struct that holds all the options for the command.
type Options struct {
	Owner string
	Repo  string
	Num   int
	URL   string
	Token string
}

// Config is a complete configuration for the app.
type Config struct {
	Client *gogithub.Client
	Owner  string
	Repo   string
	Num    int
}

func Validate(c *Config) error {
	errBuilder := []string{}

	if c.Client == nil {
		errBuilder = append(errBuilder, "--token or env:GITHUB_TOKEN is required")
	}

	if c.Owner == "" {
		errBuilder = append(errBuilder, "--owner or --url with owner in path is required")
	}

	if c.Repo == "" {
		errBuilder = append(errBuilder, "--repo or --url with a repository in path is required")
	}

	if c.Num == 0 {
		errBuilder = append(errBuilder, "--number or --url with a pull request number in path is required")
	}

	if len(errBuilder) > 0 {
		return errors.New(strings.Join(append([]string{"\n\t"}, errBuilder...), "\n\t"))
	}

	return nil
}

func Complete(o Options) (*Config, error) {
	c := Config{}

	if o.Token == "" {
		o.Token = os.Getenv("GITHUB_TOKEN")
	}

	if o.Token != "" {
		c.Client = gogithub.NewClient(
			oauth2.NewClient(
				context.Background(),
				oauth2.StaticTokenSource(
					&oauth2.Token{AccessToken: o.Token},
				),
			),
		)
	}

	if o.URL == "" {
		c.Owner = o.Owner
		c.Repo = o.Repo
		c.Num = o.Num

		return &c, nil
	}

	u, err := url.Parse(o.URL)
	if err != nil {
		return nil, err
	}

	owner := 0
	repo := 1
	num := 3
	parts := strings.Split(path.Clean(u.Path), "/")

	if len(parts) > 0 && parts[0] == "" {
		parts = parts[1:]
	}

	if len(parts) < 4 || parts[2] != "pull" {
		return nil, fmt.Errorf(
			"invalid pull request URL: parts=%q -> owner=parts[%d], repo=parts[%d], num=parts[%d]",
			parts, owner, repo, num,
		)
	}

	c.Owner, c.Repo = parts[owner], parts[repo]
	c.Num, err = strconv.Atoi(parts[num])
	if err != nil {
		return nil, err
	}

	return &c, nil
}
