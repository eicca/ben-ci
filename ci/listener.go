package ci

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// ListenAndServe runs new http server for ben-ci.
func ListenAndServe() {
	http.HandleFunc("/gh_hook", githubHook)

	port := ":" + os.Getenv("LISTENER_PORT")
	fmt.Printf("CI listens for hooks on http://127.0.0.1%s\n", port)
	fmt.Println(http.ListenAndServe(port, nil))
}

func githubHook(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	// TODO change to push event
	// then iterate over all open pull requests
	event := github.PullRequestEvent{}
	if err := decoder.Decode(&event); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	if err := lint(event); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func lint(event github.PullRequestEvent) error {
	client := newGithubClient()
	ok, err := rebased(client, event.PullRequest)
	if err != nil {
		return err
	}

	// TODO clean the mess with RepoStatus
	var status github.RepoStatus
	var res string
	if ok {
		res = "success"
	} else {
		res = "failure"
	}
	desc := "Check if rebased"
	context := "ben-ci"
	status.State = &res
	status.Description = &desc
	status.Context = &context

	_, _, err = client.Repositories.CreateStatus(*event.Repo.Owner.Login, *event.Repo.Name, *event.PullRequest.Head.SHA, &status)
	return err
}

// TODO take info from event maybe
func rebased(client *github.Client, pr *github.PullRequest) (bool, error) {
	base := *pr.Base.Label
	head := *pr.Head.Label
	owner := *pr.Base.Repo.Owner.Login
	repo := *pr.Base.Repo.Name

	comp, _, err := client.Repositories.CompareCommits(owner, repo, base, head)
	if err != nil {
		return false, fmt.Errorf("Error during comparing commits: %s", err)
	}

	return *comp.BehindBy == 0, nil
}

func newGithubClient() *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GH_ACCESS_TOKEN")},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	return github.NewClient(tc)
}
