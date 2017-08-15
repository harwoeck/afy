package main

import (
	"net/http"

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

	authorized := false
	// If no authentication method is defined set authorized to true and continue
	if !config.ACP.Whitelist && !config.ACP.Org {
		authorized = true
	} else {
		// Query whitelisted persons
		if config.ACP.Whitelist {
			for _, entry := range whitelist {
				if entry == *githubUser.ID {
					authorized = true
					break
				}
			}
		}
		// Check if the the user is a member of a specific organization to enable
		// SSO functionality
		if config.ACP.Org {
			orgs, _, err := client.Organizations.List(oauth2.NoContext, "", nil)
			if err != nil {
				log.Warningf("login: github(%d).Orgs.List returned: %s", *githubUser.ID, err.Error())
			} else {
				for _, item := range orgs {
					if *item.Login == config.ACP.OrgName {
						authorized = true
						break
					}
				}
			}
		}
	}
	// Send forbidden pages to unauthorized users and log their github id
	if !authorized {
		log.Infof("login: github(%d) -> not authorized", *githubUser.ID)
		sendError(w, http.StatusForbidden)
		return
	}

	// If user is already logged in delete the old key
	if k, exists := userKey[*githubUser.ID]; exists {
		delete(keyUser, k)
		delete(userKey, *githubUser.ID)
	}

	key := keyProvider(32)
	keyUser[key] = *githubUser.ID
	userKey[*githubUser.ID] = key
	log.Infof("login: github(%d) -> %s", *githubUser.ID, key)

	http.Redirect(w, r, "/f/"+key+"/", http.StatusTemporaryRedirect)
}
