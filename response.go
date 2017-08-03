package main

import "html/template"

type response struct {
	QuickPathNavigation template.HTML
	Index               []indexItem
	IndexF              []indexItem
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

type indexItem struct {
	Name     string
	Link     string
	Modified string
	Size     int64
}
