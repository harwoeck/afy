package main

import (
	"encoding/base64"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
)

func base64Must(str string) []byte {
	buffer, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		log.Fatal("One of the specified auth/crypt keys of your cookie store is invalid")
	}
	return buffer
}

func recoveryHandler(handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rval := recover(); rval != nil {
				debug.PrintStack()
				sendError(w, http.StatusInternalServerError)
			}
		}()
		handler(w, r)
	}
}

func assertFSDirPtr(dir *string) {
	abs, err := filepath.Abs(*dir)
	if err != nil {
		log.Fatal(err.Error())
	}
	fi, err := os.Stat(abs)
	if err != nil {
		log.Fatal(err.Error())
	}
	if !fi.IsDir() {
		log.Fatalf("%s isn't a directory", *dir)
	}
	*dir = abs
}

func assertFSDir(dir string) string {
	assertFSDirPtr(&dir)
	return dir
}

func assertFSFilePtr(file *string) {
	abs, err := filepath.Abs(*file)
	if err != nil {
		log.Fatalf(err.Error())
	}
	_, err = os.Stat(abs)
	if err != nil {
		log.Fatal(err.Error())
	}
	*file = abs
}

func assertFSFile(file string) string {
	assertFSFilePtr(&file)
	return file
}
