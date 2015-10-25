package ci

import (
	"fmt"
	"os"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

var (
	integrationName  = "ben-ci"
	rebaseDescripton = "Check if rebased"
)

// Linter makes git based checks on the different events.
type Linter struct {
	client *github.Client
}

// NewLinter returns a new instance of Linter with initialized gh client.
func NewLinter() *Linter {
	l := Linter{}
	l.client = newGithubClient()
	return &l
}

// Lint checks if the PR was rebased againts master.
func (l *Linter) Lint(event github.PullRequestEvent) error {
	status, err := l.rebasedStatus(event.PullRequest)
	if err != nil {
		return err
	}

	_, _, err = l.client.Repositories.CreateStatus(*event.Repo.Owner.Login, *event.Repo.Name, *event.PullRequest.Head.SHA, &status)
	return err
}

func (l *Linter) rebasedStatus(pr *github.PullRequest) (github.RepoStatus, error) {
	base := *pr.Base.Label
	head := *pr.Head.Label
	owner := *pr.Base.Repo.Owner.Login
	repo := *pr.Base.Repo.Name

	comp, _, err := l.client.Repositories.CompareCommits(owner, repo, base, head)
	if err != nil {
		return github.RepoStatus{}, fmt.Errorf("Error during comparing commits: %s", err)
	}

	rebased := *comp.BehindBy == 0

	var status github.RepoStatus
	state := map[bool]string{true: "success", false: "failure"}[rebased]
	status.State = &state
	status.Description = &rebaseDescripton
	status.Context = &integrationName

	return status, nil
}

func newGithubClient() *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GH_ACCESS_TOKEN")},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	return github.NewClient(tc)
}
