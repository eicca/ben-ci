package ci

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const (
	eventPush        = "push"
	eventPullRequest = "pull_request"
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

// ProcessHook performs specific linter logic depending on the event type.
func (l *Linter) ProcessHook(eventType string, body io.Reader) error {
	eventHandlers := map[string]func(*json.Decoder) error{
		eventPush:        l.processPushEvent,
		eventPullRequest: l.processPullRequestEvent,
	}

	decoder := json.NewDecoder(body)
	fn, ok := eventHandlers[eventType]
	if !ok {
		return fmt.Errorf("ci: unrecognized type of the event: %s", eventType)
	}

	return fn(decoder)
}

func (l *Linter) processPushEvent(decoder *json.Decoder) error {
	event := github.PushEvent{}
	if err := decoder.Decode(&event); err != nil {
		return err
	}

	// By default it returns all open PRs sorted by creation date.
	prs, _, err := l.client.PullRequests.List(*event.Repo.Owner.Name, *event.Repo.Name,
		&github.PullRequestListOptions{})
	if err != nil {
		return err
	}

	for _, pr := range prs {
		status, err := l.rebasedStatus(&pr)
		if err != nil {
			return err
		}

		_, _, err = l.client.Repositories.CreateStatus(*event.Repo.Owner.Name, *event.Repo.Name, *pr.Head.SHA, &status)
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *Linter) processPullRequestEvent(decoder *json.Decoder) error {
	event := github.PullRequestEvent{}
	if err := decoder.Decode(&event); err != nil {
		return err
	}

	// TODO here it's better to run linting in another goroutine
	// and post result or error to commit gh status.
	linter := NewLinter()
	return linter.lint(event)
}

func (l *Linter) lint(event github.PullRequestEvent) error {
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
