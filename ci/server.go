package ci

import (
	"fmt"
	"net/http"
)

var linter = NewLinter()

// ListenAndServe runs new http server for ben-ci.
func ListenAndServe(port string) {
	http.HandleFunc("/gh_hook", ghHookHandler)

	fmt.Printf("CI listens for hooks on http://127.0.0.1%s\n", port)
	fmt.Println(http.ListenAndServe(port, nil))
}

func ghHookHandler(w http.ResponseWriter, r *http.Request) {
	eventType := r.Header.Get("X-Github-Event")
	err := linter.ProcessHook(eventType, r.Body)
	defer r.Body.Close()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
}
