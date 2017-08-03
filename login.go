package main

import (
	"net/http"
	"strings"
)

func login(w http.ResponseWriter, r *http.Request) {
	log.Info("In login")
	url := r.URL.Path[len("/f/"):]
	if !strings.Contains(url, "/") {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if strings.Index(url, "/") != 32 {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var id int
	var ok bool
	if id, ok = users[url[:32]]; !ok {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	log.Infof("access: github(%d) -> %s", id, url[32:])

	a := r.URL.Path[:len("/f/")+32+len("/")]
	handle(w, r, a, strings.TrimPrefix(r.URL.Path, a))
}
