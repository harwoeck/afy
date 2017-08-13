package main

import (
	"bufio"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func githubLogin(w http.ResponseWriter, r *http.Request) {
	session, err := cookies.Get(r, "_oauthState")
	if err != nil && err.Error() != "securecookie: the value is not valid" {
		log.Error(err.Error())
		sendError(w, http.StatusInternalServerError)
		return
	}

	stateString := keyProvider(32)

	session.Values["github"] = stateString
	err = session.Save(r, w)
	if err != nil {
		log.Error(err.Error())
		sendError(w, http.StatusInternalServerError)
		return
	}

	url := configGithub.AuthCodeURL(stateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func githubCallback(w http.ResponseWriter, r *http.Request) {
	session, err := cookies.Get(r, "_oauthState")
	if err != nil {
		log.Error(err.Error())
		sendError(w, http.StatusInternalServerError)
		return
	}

	state := session.Values["github"]
	var expectedState string
	var ok bool
	if expectedState, ok = state.(string); !ok {
		log.Error("couldn't get expectedState from cookie _oauthState.github")
		sendError(w, http.StatusInternalServerError)
		return
	}

	actualState := r.FormValue("state")
	if actualState != expectedState {
		log.Warningf("couldn't authenticate user: actualState(%s) and expectedState(%s) don't match. login denied", actualState, expectedState)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	code := r.FormValue("code")
	token, err := configGithub.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Error(err.Error())
		sendError(w, http.StatusInternalServerError)
		return
	}

	httpClient := configGithub.Client(oauth2.NoContext, token)
	client := github.NewClient(httpClient)

	githubUser, _, err := client.Users.Get(oauth2.NoContext, "")
	if err != nil {
		log.Error(err.Error())
		sendError(w, http.StatusInternalServerError)
		return
	}

	var found bool
	f, err := os.Open("github.txt")
	if err != nil {
		log.Error(err.Error())
		sendError(w, http.StatusInternalServerError)
		return
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		cur := scanner.Text()
		if strings.Contains(cur, " ") {
			cur = strings.TrimSpace(strings.Split(cur, " ")[0])
		}
		itemID, err := strconv.Atoi(cur)
		if err != nil {
			log.Error(err.Error())
			sendError(w, http.StatusInternalServerError)
			return
		}
		if itemID == *githubUser.ID {
			found = true
			break
		}
	}
	if err := scanner.Err(); err != nil {
		log.Error(err.Error())
		sendError(w, http.StatusInternalServerError)
		return
	}
	if !found {
		log.Infof("login: github(%d) -> not authorized", *githubUser.ID)
		sendError(w, http.StatusForbidden)
		return
	}

	key := keyProvider(32)

	users[key] = *githubUser.ID
	log.Infof("login: github(%d) -> %s", *githubUser.ID, key)

	http.Redirect(w, r, "/f/"+key+"/", http.StatusTemporaryRedirect)
}
