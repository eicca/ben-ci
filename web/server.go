package web

import (
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github"
)

var (
	oauthConf = &oauth2.Config{
		ClientID:     os.Getenv("GH_CLIENT_ID"),
		ClientSecret: os.Getenv("GH_CLIENT_SECRET"),
		Scopes:       []string{"user:email", "repo"},
		Endpoint:     githuboauth.Endpoint,
	}
	oauthStateString = "7vYviTJFu6FH75khN0QM"
)

// ListenAndServe runs new http server for ben-ci.
func ListenAndServe() {
	http.HandleFunc("/", index)
	http.HandleFunc("/login", githubLogin)
	http.HandleFunc("/gh_oauth_cb", githubCallback)

	port := ":" + os.Getenv("SERVER_PORT")
	fmt.Printf("Started running on http://127.0.0.1%s\n", port)
	fmt.Println(http.ListenAndServe(port, nil))
}

var htmlIndex = `<html><body>
Login in with <a href="/login">GitHub</a>
</body></html>
`

func index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(htmlIndex))
}

func githubLogin(w http.ResponseWriter, r *http.Request) {
	url := oauthConf.AuthCodeURL(oauthStateString, oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func githubCallback(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	if state != oauthStateString {
		msg := fmt.Sprintf("invalid oauth state, expected '%s', got '%s'\n", oauthStateString, state)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(msg))
		return
	}

	code := r.FormValue("code")
	token, err := oauthConf.Exchange(oauth2.NoContext, code)
	if err != nil {
		msg := fmt.Sprintf("oauthConf.Exchange() failed with '%s'\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(msg))
		return
	}

	if !token.Valid() {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("%+v", token)))
		return
	}

	msg := fmt.Sprintf("Auth was successful. Your token: %s", token.AccessToken)
	w.Write([]byte(msg))
}
