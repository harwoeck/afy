package main

import (
	"io/ioutil"
	"os"
	"strings"
	"time"
)

func cacheworker() {
	for {
		time.Sleep(time.Second * time.Duration(config.FSCache.RefreshTime))

	}
}

func buildIndexDTree(f os.FileInfo, abs string) []*indexd {
	fi, err := ioutil.ReadDir(abs)
	// Handle normal and filesystem permission errors
	if err != nil && strings.Contains(err.Error(), "permission denied") {
		resp := make([]*indexd, 1)
		resp[0] = &indexd{
			Title:    strings.TrimPrefix(abs, config.Mnt) + "/",
			Modified: f.ModTime(),
			Access:   false,
		}
		return resp
	} else if err != nil {
		log.Error(err.Error())
		return nil
	}

	return nil
}
