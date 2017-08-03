package main

import (
	"io/ioutil"
	"mime"
	"net/http"
	"strings"
)

func static(w http.ResponseWriter, r *http.Request, abs string) {
	ext := abs[strings.LastIndex(abs, "."):]
	mimeType := mime.TypeByExtension(ext)
	w.Header().Set("Content-Type", mimeType)

	content, err := ioutil.ReadFile(abs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(content)
}
