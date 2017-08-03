package main

import (
	"html/template"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

type response struct {
	QuickPathNavigation template.HTML
	Index               []afyFile
	DirSize             int64

	Package          bool
	PackageName      string
	PackageHierarchy string
	PackageDependsOn template.HTML

	Git        bool
	GitHash    string
	GitLink    string
	GitMessage string

	CI          bool
	CIJob       string
	CILink      string
	CIBuildTime string
	CICoverage  string
	CIHasReport bool
}

type afyFile struct {
	Name     string
	Link     string
	Modified string
	Size     int64
}

func getDirList(auth string, p string) (r *response, isDownload bool, isForbidden bool, is404 bool, is500 bool) {
	abs := path.Join(path.Dir(config.Root), p)
	if !strings.HasPrefix(abs+"/", config.Root) {
		return nil, false, true, false, false
	}

	stat, err := os.Stat(abs)
	if err != nil {
		return nil, false, false, true, false
	}

	if !stat.IsDir() {
		return nil, true, false, false, false
	}

	abs += "/"

	r = &response{}

	//
	// Quick path navigation
	//
	parts := strings.Split("/"+strings.TrimSuffix(strings.TrimPrefix(abs, config.Root), "/"), "/")
	qpn := ""
	pathYet := strings.TrimSuffix(auth, "/")
	for _, item := range parts {
		itemFull := item + "/"
		pathYet += itemFull
		qpn += "<a href=\"" + pathYet + "\">" + itemFull + "</a>"
	}
	log.Info(qpn)
	r.QuickPathNavigation = template.HTML(qpn)

	//
	// Index
	//
	fis, err := ioutil.ReadDir(abs)
	if err != nil {
		log.Error(err.Error())
		return nil, false, false, false, true
	}
	for _, fi := range fis {
		name := fi.Name()
		if fi.IsDir() {
			name += "/"
		}
		r.DirSize += fi.Size()
		r.Index = append(r.Index, afyFile{
			Name:     name,
			Link:     "./" + name,
			Modified: fi.ModTime().Format("02-Jan-2006 15:04"),
			Size:     fi.Size(),
		})
	}

	//
	// Package
	//
	_, err = os.Stat(abs + "_package.afy")
	if err == nil {
		r.Package = true
		content, err := ioutil.ReadFile(abs + "_package.afy")
		if err != nil {
			return nil, false, false, false, true
		}
		lines := strings.Split(string(content), "\n")
		r.PackageName = lines[0]
		r.PackageHierarchy = lines[1]
		r.PackageDependsOn = template.HTML(strings.Join(lines[2:], ", "))
	}

	//
	// Git
	//
	_, err = os.Stat(abs + "_git.afy")
	if err == nil {
		r.Git = true
		content, err := ioutil.ReadFile(abs + "_git.afy")
		if err != nil {
			return nil, false, false, false, true
		}
		lines := strings.Split(string(content), "\n")
		r.GitHash = lines[0]
		r.GitLink = lines[1]
		r.GitMessage = lines[2]
	}

	//
	// CI
	//
	_, err = os.Stat(abs + "_ci.afy")
	if err == nil {
		r.CI = true
		content, err := ioutil.ReadFile(abs + "_ci.afy")
		if err != nil {
			return nil, false, false, false, true
		}
		lines := strings.Split(string(content), "\n")
		r.CIJob = lines[0]
		r.CILink = lines[1]
		r.CIBuildTime = lines[2]
		if len(lines) > 3 {
			r.CICoverage = lines[3]
			r.CIHasReport = true
		}
	}

	return r, false, false, false, false
}
