package main

import (
	"io/ioutil"
	"mime"
	"net/http"
	"strings"
)

func handlestatic(w http.ResponseWriter, r *http.Request, abs string) {
	if strings.Contains(abs, ".") {
		ext := abs[strings.LastIndex(abs, "."):]
		mimeType := mime.TypeByExtension(ext)
		if mimeType != "" {
			w.Header().Set("Content-Type", mimeType)
		}
	}

	content, err := ioutil.ReadFile(abs)
	if err != nil {
		sendError(w, http.StatusInternalServerError)
		return
	}

	w.Write(content)
}
