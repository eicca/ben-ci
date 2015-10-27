package ci

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/go-github/github"
	"io"
)

const (
	eventPush        = "push"
	eventPullRequest = "pull_request"
)

var eventHandlers = map[string]func(*json.Decoder) error{
	eventPush:        processPushEvent,
	eventPullRequest: processPullRequestEvent,
}

// ListenAndServe runs new http server for ben-ci.
func ListenAndServe(port string) {
	http.HandleFunc("/gh_hook", ghHookHandler)

	fmt.Printf("CI listens for hooks on http://127.0.0.1%s\n", port)
	fmt.Println(http.ListenAndServe(port, nil))
}

func ghHookHandler(w http.ResponseWriter, r *http.Request) {
	eventType := r.Header.Get("X-Github-Event")
	err := processHook(eventType, r.Body)
	defer r.Body.Close()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func processHook(eventType string, body io.Reader) error {
	decoder := json.NewDecoder(body)

	fn, ok := eventHandlers[eventType]
	if !ok {
		return fmt.Errorf("ci: unrecognized type of the event: %s", eventType)
	}

	return fn(decoder)
}

func processPushEvent(decoder *json.Decoder) error {
	// TODO handle push event here and update (all open?) PRs.
	return nil
}

func processPullRequestEvent(decoder *json.Decoder) error {
	event := github.PullRequestEvent{}
	if err := decoder.Decode(&event); err != nil {
		return err
	}

	// TODO here it's better to run linting in another goroutine
	// and post result or error to commit gh status.
	linter := NewLinter()
	return linter.Lint(event)
}
