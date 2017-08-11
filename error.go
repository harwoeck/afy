package main

import (
	"fmt"
	"net/http"
)

func sendError(w http.ResponseWriter, code int) {
	w.WriteHeader(code)
	switch code {
	case http.StatusNotFound:
		fmt.Fprintf(w, "<h1>Not found</h1>")
	case http.StatusForbidden:
		fmt.Fprintf(w, "<h1>Forbidden</h1>")
	case http.StatusInternalServerError:
		fmt.Fprintf(w, "<h1>InternalServerError</h1>")
	}
}
