package main

import (
	"net/http"
	"strings"
)

func authlayer(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path[len("/f/"):]
	if !strings.Contains(url, "/") {
		sendError(w, http.StatusForbidden)
		return
	}
	if strings.Index(url, "/") != 32 {
		sendError(w, http.StatusForbidden)
		return
	}

	var id int
	var ok bool
	if id, ok = keyUser[url[:32]]; !ok {
		sendError(w, http.StatusForbidden)
		return
	}

	log.Infof("access: github(%d) -> %s", id, url[32:])

	a := r.URL.Path[:len("/f/")+32+len("/")]
	handleindex(w, r, a, strings.TrimPrefix(r.URL.Path, a))
}
