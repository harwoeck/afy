package main

import "net/http"

func router(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.Redirect(w, r, "/auth/github/login", http.StatusTemporaryRedirect)
		return
	}
	sendError(w, http.StatusNotFound)
}
