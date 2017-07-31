package main

import (
	"net/http"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func githubLogin(w http.ResponseWriter, r *http.Request) {
	session, err := cookies.Get(r, cookieOAuthState)
	if err != nil {
		log.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	stateString, err := keyProvider(32, 10)
	if err != nil {
		log.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	session.Values["github"] = stateString
	err = session.Save(r, w)
	if err != nil {
		log.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	url := configGithub.AuthCodeURL(stateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func githubCallback(w http.ResponseWriter, r *http.Request) {
	session, err := cookies.Get(r, cookieOAuthState)
	if err != nil {
		log.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	state := session.Values["github"]
	var expectedState string
	var ok bool
	if expectedState, ok = state.(string); !ok {
		log.Errorf("couldn't get expectedState from cookie %s.github", cookieOAuthState)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	actualState := r.FormValue("state")
	if actualState != expectedState {
		log.Error("couldn't authenticate user: actualState and expectedState don't match. login denied")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	code := r.FormValue("code")
	token, err := configGithub.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	httpClient := configGithub.Client(oauth2.NoContext, token)
	client := github.NewClient(httpClient)

	// Receive user from github client
	githubUser, _, err := client.Users.Get(oauth2.NoContext, "")
	if err != nil {
		log.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	key, err := keyProvider(32, 10)
	if err != nil {
		log.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	users[key] = *githubUser.ID
	log.Info("login: github(%d) -> %s", *githubUser.ID, key)

	http.Redirect(w, r, "/fs/"+key+"/", http.StatusTemporaryRedirect)
}
