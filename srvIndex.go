package main

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func srvIndex(w http.ResponseWriter, r *http.Request, a string, p string) {
	jnd := path.Join(config.Mnt, p)
	abs, err := filepath.Abs(jnd)
	if err != nil {
		sendError(w, http.StatusForbidden)
		return
	}
	if !strings.HasPrefix(abs, config.Mnt) {
		sendError(w, http.StatusForbidden)
		return
	}

	stat, err := os.Stat(abs)
	if err != nil {
		sendError(w, http.StatusNotFound)
		return
	}
	if !stat.IsDir() {
		srvStatic(w, r, abs)
		return
	}

	pipe := response{}

	//
	// Quick path navigation
	//
	qpn := "<a href=\"" + a + "\">/</a>"
	if strings.TrimSpace(p) != "" {
		pathYet := a
		parts := strings.Split(strings.TrimSuffix(p, "/"), "/")
		for _, item := range parts {
			itemFull := item + "/"
			pathYet += itemFull
			qpn += "<a href=\"" + pathYet + "\">" + itemFull + "</a>"
		}
	}
	pipe.QuickPathNavigation = template.HTML(qpn)

	//
	// Index
	//
	fis, err := ioutil.ReadDir(abs)
	if err != nil && strings.Contains(err.Error(), "permission denied") {
		sendError(w, http.StatusForbidden)
		return
	} else if err != nil {
		log.Error(err.Error())
		sendError(w, http.StatusInternalServerError)
		return
	}
	for _, fi := range fis {
		pipe.DirSize += fi.Size()
		ii := indexItem{
			Name:     fi.Name(),
			Link:     "./" + fi.Name(),
			Modified: fi.ModTime().Format("02-Jan-2006 15:04"),
			Size:     fi.Size(),
		}
		if fi.IsDir() {
			ii.Name += "/"
			ii.Link += "/"
			pipe.Index = append(pipe.Index, ii)
		} else {
			pipe.IndexF = append(pipe.IndexF, ii)
		}
	}
	pipe.Index = append(pipe.Index, pipe.IndexF...)

	//
	// Package
	//
	_, err = os.Stat(abs + "/_package.afy")
	if err == nil {
		pipe.Package = true
		content, err := ioutil.ReadFile(abs + "/_package.afy")
		if err != nil {
			log.Error(err.Error())
			sendError(w, http.StatusInternalServerError)
			return
		}
		lines := strings.Split(string(content), "\n")
		pipe.PackageName = lines[0]
		pipe.PackageHierarchy = lines[1]
		pipe.PackageDependsOn = template.HTML(strings.Join(lines[2:], ", "))
	}

	//
	// Git
	//
	_, err = os.Stat(abs + "/_git.afy")
	if err == nil {
		pipe.Git = true
		content, err := ioutil.ReadFile(abs + "/_git.afy")
		if err != nil {
			log.Error(err.Error())
			sendError(w, http.StatusInternalServerError)
			return
		}
		lines := strings.Split(string(content), "\n")
		pipe.GitHash = lines[0]
		pipe.GitLink = lines[1]
		pipe.GitMessage = lines[2]
	}

	//
	// CI
	//
	_, err = os.Stat(abs + "/_ci.afy")
	if err == nil {
		pipe.CI = true
		content, err := ioutil.ReadFile(abs + "/_ci.afy")
		if err != nil {
			log.Error(err.Error())
			sendError(w, http.StatusInternalServerError)
			return
		}
		lines := strings.Split(string(content), "\n")
		pipe.CIJob = lines[0]
		pipe.CILink = lines[1]
		pipe.CIBuildTime = lines[2]
		if len(lines) > 3 {
			pipe.CICoverage = lines[3]
			pipe.CIHasReport = true
		}
	}

	// Branding
	if config.Branding.Name == "" {
		pipe.Title = "/" + p + " - " + config.Host
	} else {
		pipe.Title = "/" + p + " - " + config.Branding.Name
	}
	pipe.Description = config.Branding.Description
	pipe.Keywords = config.Branding.Keywords
	if config.Branding.Favicon == "" {
		pipe.Favicon = "https://afy.io/content/afyio_logo.png"
	} else {
		pipe.Favicon = config.Branding.Favicon
	}

	tmpl.Execute(w, pipe)
}
