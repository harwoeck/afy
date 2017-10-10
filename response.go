package main

import (
	"html/template"
	"time"
)

// indexd(irectory)
type indexd struct {
	Title    string
	Link     string
	Modified time.Time
	Size     int64
	Access   bool

	QuickPathNavigation template.HTML

	Dirs  []indexd
	Files []indexf

	Pkg pkg
	CI  ci
	Git git
}

type pkg struct {
	Name        string
	Link        string
	Description string
	DependsOn   template.HTML
}

type git struct {
	Hash    string
	Link    string
	Message string
	Branch  string
	Tag     string
	Author  struct {
		Name  string
		Email string
		Link  string
	}
}

type ci struct {
	Job       string
	Link      string
	BuildTime string
	Coverage  []struct {
		Percentage string
		Report     string
	}
}

// indexf(ile)
type indexf struct {
}
